package hashi

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/nomad/api"
)

type JobSelect struct {
	Stub   *api.JobListStub
	Allocs []*api.AllocationListStub
}

type LogPayload struct {
	frame        *api.StreamFrame
	logType      string
	jobTaskAlloc *JobTaskAlloc
	writers      []io.Writer
}

type StreamLogs struct {
	allocs        []*JobTaskAlloc
	outputCh      chan *LogPayload
	nomadAllocFS  NomadAllocFS
	nomadAllocs   NomadAllocations
	allocCache    map[string]*api.Allocation
	output        io.Writer
	config        *StreamLogsConfig
	globalCancel  chan struct{}
	outputStreams LogStreamIO
}

func NewStreamLogs(client *api.Client, j []*JobTaskAlloc, config *StreamLogsConfig, outputStreams LogStreamIO) *StreamLogs {
	s := StreamLogs{
		allocs:        j,
		outputCh:      make(chan *LogPayload),
		nomadAllocFS:  client.AllocFS(),
		nomadAllocs:   client.Allocations(),
		allocCache:    make(map[string]*api.Allocation),
		output:        os.Stderr,
		config:        config,
		globalCancel:  make(chan struct{}),
		outputStreams: outputStreams,
	}

	return &s
}

type LogStreamIO interface {
	GetStreams(jta *JobTaskAlloc, logSrc string) []io.Writer
}

type LogWriters struct {
	globalWriters []io.Writer
	savePath      string
	fileRefs      sync.Map // key: filePath => *os.File
}

func (l *LogWriters) ensureFileDescriptor(savePath string) (io.Writer, error) {

	var writer io.Writer

	if fdRaw, chk := l.fileRefs.Load(savePath); chk {
		writer = fdRaw.(*os.File)
	} else {
		fd, err := os.Create(savePath)
		if err != nil {
			return nil, err
		}

		l.fileRefs.Store(savePath, fd)
		writer = fd
	}

	return writer, nil
}

func (l *LogWriters) GetStreams(jta *JobTaskAlloc, logSrc string) []io.Writer {
	s := make([]io.Writer, 0)

	// get/set/cache file descriptor
	if len(l.savePath) > 0 {
		//  create filename and path
		allocID := strings.SplitN(jta.AllocID, "-", 1)[0]
		fn := fmt.Sprintf("%s.%s.%s.%s.%s.%s.log",
			jta.Host, allocID, jta.Namespace, jta.Job, jta.Task, logSrc)
		fullPath := fmt.Sprintf("%s/%s", l.savePath, fn)

		// ensure that the file descriptor is created
		fileWriter, err := l.ensureFileDescriptor(fullPath)

		if err != nil {
			log.Printf("unable to ensure path: %s, error: %s", fullPath, err.Error())
		}

		if fileWriter != nil {
			s = append(s, fileWriter)
		}

	}

	for _, wr := range l.globalWriters {
		s = append(s, wr)
	}

	return s
}

func (l *LogWriters) CloseAll() {
	l.fileRefs.Range(func(_, rawFd any) bool {
		fd := rawFd.(*os.File)
		fd.Close()
		return true
	})
}

func NewLogWriters(c *StreamLogsConfig) (*LogWriters, error) {
	l := LogWriters{}

	if len(c.SavePath) > 0 {
		l.savePath = c.SavePath
		savePathErr := l.ensureSavePath()
		if savePathErr != nil {
			return nil, savePathErr
		}
	}

	ws := make([]io.Writer, 0)

	if c.StreamOutput != nil {
		ws = append(ws, c.StreamOutput)
	}

	l.globalWriters = ws

	return &l, nil
}

func (l *LogWriters) ensureSavePath() error {
	err := os.MkdirAll(l.savePath, 0755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	return nil
}

type StreamLogsConfig struct {
	// Follow keep stream open while writing to io.Writer's
	Follow bool
	// Origin where in the log stream to start (values: "start" or "end")
	Origin string
	// OriginOffset move cursor in the stream buffer to the offset of `Origin`
	OriginOffset int64
	// Source "stdout" or "stderr"
	LogType      []string
	SavePath     string
	StreamOutput io.Writer
	StreamWaiter *sync.WaitGroup
}

func (l *StreamLogs) watchDog(ctx context.Context) {
	for ctx.Err() == nil {
		<-time.After(1 * time.Second)
	}
}

// FIXME: this can use some refactoring of the []io.Writers and *LogPayload handling
func (l *StreamLogs) startStream(ctx context.Context, logType string, jta *JobTaskAlloc) {

	alloc := l.getAlloc(jta.Namespace, jta.AllocID)

	spew.Dump("AllocID: ", alloc.ID)

	// FIXME: centralize cancel chan in parent dereferenced struct
	cancelStream := make(chan struct{})

	ops := api.QueryOptions{
		Namespace: jta.Namespace,
		// TODO: parameterize
		// AllowStale: true,
	}

	// get io.Writers
	writers := l.outputStreams.GetStreams(jta, logType)

	stream, streamErr := l.nomadAllocFS.Logs(alloc, l.config.Follow, jta.Task, logType, "end", 0, cancelStream, &ops)

	for {
		select {
		case logErr := <-streamErr:
			log.Printf("logs error, task: %s, error: %s", jta.Task, logErr.Error())
		case frame := <-stream:
			if frame == nil {
				continue
			}

			// send to the output handler
			l.outputCh <- &LogPayload{
				frame:        frame,
				jobTaskAlloc: jta,
				logType:      logType,
				writers:      writers,
			}
			break
		case <-ctx.Done():
			log.Println("startStream exit")
			return
		}
	}

}

func (l *StreamLogs) getAlloc(namespace, allocID string) *api.Allocation {
	if a, c := l.allocCache[allocID]; c {
		return a
	}

	ops := api.QueryOptions{
		Namespace: namespace,
	}
	alloc, _, err := l.nomadAllocs.Info(allocID, &ops)
	if err != nil {
		return nil
	}

	l.allocCache[allocID] = alloc
	return alloc
}

func (l *StreamLogs) outputHandler(ctx context.Context) {

	for {
		select {
		case payload := <-l.outputCh:
			id := strings.Split(payload.jobTaskAlloc.AllocID, "-")[0]
			hn := payload.jobTaskAlloc.Host
			jn := payload.jobTaskAlloc.Job
			lt := payload.logType
			tn := payload.jobTaskAlloc.Task

			lines := strings.Split(string(payload.frame.Data), "\n")

			for _, ln := range lines {
				ts := time.Now().Format(time.RFC3339)
				prefix := fmt.Sprintf("[%s][allocID:%s][host:%s][job:%s][task:%s][%s]", ts, id, hn, jn, tn, lt)
				line := fmt.Sprintf("%s - %s\n", prefix, ln)
        for _, wr := range payload.writers {
          wr.Write([]byte(line))
        }
			}
			break
		case <-ctx.Done():
			return
		}
	}

}

func (l *StreamLogs) Run(ctx context.Context) {

	// start watch dog
	// FIXME: needs implementation
	// go l.watchDog(ctx)

	for _, jta := range l.allocs {
		for _, logType := range l.config.LogType {
			go l.startStream(ctx, logType, jta)
		}
	}

	l.outputHandler(ctx)
	log.Println("StreamLogs.Run() exiting ...")
}
