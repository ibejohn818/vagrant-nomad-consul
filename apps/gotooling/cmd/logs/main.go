package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gotooling/johnhardy.io/pkg/hashi"
	"gotooling/johnhardy.io/pkg/utils"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	nomad "github.com/hashicorp/nomad/api"
)

type listFlag []string

func (l *listFlag) String() string {
	return strings.Join(*l, ", ")
}

func (l *listFlag) Set(val string) error {
	if len(*l) >= 2 {
		return errors.New("too many values")
	}

	if !utils.StrInArray(val, []string{"stderr", "stdout"}) {
		return errors.New("invalid value!")
	}

	*l = append(*l, val)

	return nil
}

var (
	sources listFlag = make([]string, 0)
	follow  bool     = false
	saveTo  string   = "/tmp/logs"
	ns      listFlag = make([]string, 0)
)

const usage = `
  -f, --follow  follow the log stream
  -t, --type  where to get logs, can be "stderr" or "stdout" 
                (use option multiple times for both, default: stderr)
  -s, --save    dir path to save logs, (will create path if not present),
                (TODO: not yet implemented)
  -n, --ns      namespace(s) to query, (set multiple times for multiple namespaces)

EXAMPLE - Multiple namespaces:
  {CMD} -n default --ns system --ns utils
  `

func init() {

	flag.Var(&sources, "type", "either 'stdout' or 'stderr' (can set this option multiple times)")
	flag.Var(&sources, "t", "either 'stdout' or 'stderr' (can set this option multiple times)")

	flag.BoolVar(&follow, "follow", false, "stream and follow logs?")
	flag.BoolVar(&follow, "f", false, "stream and follow logs?")

	flag.StringVar(&saveTo, "save", "/tmp/logs", fmt.Sprintf("dir path to save logs (dir will be created, default: %s)", saveTo))
	flag.StringVar(&saveTo, "s", "/tmp/logs", fmt.Sprintf("dir path to save logs (dir will be created, default: %s)", saveTo))

}

func main() {

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if len(sources) <= 0 {
		sources.Set("stderr")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := hashi.NomadClient()
	jobsc := client.Jobs()

	jobStubs := nomadSelectJob(jobsc)

	selectedJobs := getSelectedJobs(jobsc, jobStubs)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sig
		switch s {
		case syscall.SIGTERM:
			log.Println("SIGTERM, terminating immediately")
			os.Exit(0)

		case syscall.SIGINT:
			log.Println("SIGINT, shutting down")
			// cancel the root context
			cancel()
		}
	}()

	config := hashi.StreamLogsConfig{
    LogType: sources,
    Follow: follow,
  }

	streamer := hashi.NewStreamLogs(client, selectedJobs, &config)

	go streamer.Run(ctx)

	<-ctx.Done()
}

func getSelectedJobs(c hashi.NomadJobs, stubs []*nomad.JobListStub) []*hashi.JobSelect {

	selectedJobs := make([]*hashi.JobSelect, 0)

	for _, v := range stubs {

		aops := &nomad.QueryOptions{
			Namespace: v.Namespace,
		}
		allocs, _, allocsErr := c.Allocations(v.ID, false, aops)
		if allocsErr != nil {
			continue
		}

		selectedJobs = append(selectedJobs, &hashi.JobSelect{
			Stub:   v,
			Allocs: allocs,
		})
	}

	return selectedJobs
}

func jobListMap(jobs hashi.NomadJobs) map[string]*nomad.JobListStub {
	res := make(map[string]*nomad.JobListStub)

	ls, err := hashi.JobList(jobs)

	if err != nil {
		panic(err)
	}

	for _, v := range ls {
		key := fmt.Sprintf("%s.%s", v.Namespace, v.Name)
		res[key] = v
	}

	return res
}

// nomadSelectJob prompts user to select jobs and returns a list of Job structs
func nomadSelectJob(jc hashi.NomadJobs) []*nomad.JobListStub {

	jobs := jobListMap(jc)

	var res []*nomad.JobListStub
	var askRes []string

	ops := make([]string, len(jobs))

	idx := 0
	for k := range jobs {
		ops[idx] = k
		idx++
	}

	sort.Slice(ops, func(a, b int) bool {
		return ops[a] < ops[b]
	})

	sel := survey.MultiSelect{
		Message: "Select jobs {namespace}.{job-name}",
		Options: ops,
	}

	survey.AskOne(&sel, &askRes, survey.WithPageSize(15))

	for _, v := range askRes {
		res = append(res, jobs[v])
	}

	return res
}
