package hashi

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/nomad/api"
)

type JobSelect struct {
	Stub   *api.JobListStub
	Allocs []*api.AllocationListStub
}

type LogPayload struct {
	alloc    *api.Allocation
	taskName string
	frame    *api.StreamFrame
	logType  string
}

func NewStreamLogs(client *api.Client, j []*JobSelect, config *StreamLogsConfig) *StreamLogs {
	s := StreamLogs{
		jobs:         j,
		outputCh:     make(chan *LogPayload),
		nomadAllocFS: client.AllocFS(),
		nomadAllocs:  client.Allocations(),
		allocCache:   make(map[string]*api.Allocation),
		output:       os.Stderr,
		config:       config,
	}

	return &s
}

type StreamLogsConfig struct {
	Follow bool
	// Origin is either "start" or "end"
	Origin string
	// Source "stdout" or "stderr"
	LogType []string
	SaveDir string
}

type StreamLogs struct {
	jobs         []*JobSelect
	outputCh     chan *LogPayload
	nomadAllocFS NomadAllocFS
	nomadAllocs  NomadAllocations
	allocCache   map[string]*api.Allocation
	output       io.Writer
	config       *StreamLogsConfig
}

func (l *StreamLogs) startStream(ctx context.Context, logType, namespace, allocID, taskName string) {

	alloc := l.getAlloc(namespace, allocID)

	// spew.Dump(alloc.ID)

	cancelStream := make(chan struct{})

	ops := api.QueryOptions{
		Namespace:  namespace,
		AllowStale: true,
	}

	stream, streamErr := l.nomadAllocFS.Logs(alloc, l.config.Follow, taskName, logType, "end", 0, cancelStream, &ops)

	for {
		select {
		case logErr := <-streamErr:
			log.Printf("logs error, task: %s, error: %s", taskName, logErr.Error())
			return
		case frame := <-stream:
			if frame == nil {
				continue
			}

			l.outputCh <- &LogPayload{
				alloc:    alloc,
				frame:    frame,
				taskName: taskName,
				logType:  logType,
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
			id := strings.Split(payload.alloc.ID, "-")[0]
			hn := payload.alloc.NodeName
			jn := payload.alloc.JobID
			lt := payload.logType
      tn := payload.taskName

			lines := strings.Split(string(payload.frame.Data), "\n")

			for _, ln := range lines {
				ts := time.Now().Format(time.RFC3339)
				line := fmt.Sprintf("[%s][allocID:%s][host:%s][job:%s][task:%s][%s] - %s\n", ts, id, hn, jn, tn, lt, ln)
				l.output.Write([]byte(line))
			}
			break
		case <-ctx.Done():
			return
		}
	}

}

func (l *StreamLogs) Run(ctx context.Context) {

	for _, js := range l.jobs {
		for _, al := range js.Allocs {
			for taskName, st := range al.TaskStates {
				if strings.ToLower(st.State) == "running" {
					for _, logType := range l.config.LogType {
						spew.Dump(logType)
						go l.startStream(ctx, logType, js.Stub.Namespace, al.ID, taskName)
					}
				}
			}
		}
	}

	l.outputHandler(ctx)
	log.Println("StreamLogs.Run() exiting ...")
}
