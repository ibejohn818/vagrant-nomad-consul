package main

import (
	"context"
	"flag"
	"fmt"
	"gotooling/johnhardy.io/pkg/cli"
	"gotooling/johnhardy.io/pkg/hashi"
	"gotooling/johnhardy.io/pkg/utils"
	"io"
	"log"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	nomad "github.com/hashicorp/nomad/api"
)

var (
	sources   cli.ListFlag = make([]string, 0)
	follow    bool         = false
	saveTo    string
	ns        cli.ListFlag = make([]string, 0)
	taskHosts bool         = false
	out       string       = "stderr"
	offset    int64        = 0
	origin    string
	output    string = "stderr"
)

const usage = `
  -f, --follow  follow the log stream
  --origin 		where to pace the cursor when starting to read the log stream,
				(valid: "start" or "end") (default: "end" if --follow and "start" if !--follow)
  --offset 		line offset to place cursor relative to --origin (default: 0)
  -o, --out 	output log stream to tty, ("stderr" or "stdout", default: "stderr")
				(NOTE: only valid if --follow flag is set)
  -t, --type    logs source, can be "stderr" or "stdout" 
                (use option multiple times for both, default: stderr)
  -s, --save    dir path to save logs, each stream will create a file, path will be created
                (NOTE: empty value omits writing logs to file)
  -n, --ns      nomad namespace(s) to query, (set multiple times for multiple namespaces)
				(default: "default")
  --hostFilter  flag to enable hostname filter in selection wizard
				(default: selected tasks include all hosts)
                (TODO: not yet implemented)

EXAMPLE - Multiple namespaces:
  {CMD} -n default --ns system --ns utils
  `

func init() {

	flag.StringVar(&origin, "origin", "", "")
	flag.Int64Var(&offset, "offset", 0, "")

	flag.Var(&sources, "type", "either 'stdout' or 'stderr' (can set this option multiple times)")
	flag.Var(&sources, "t", "either 'stdout' or 'stderr' (can set this option multiple times)")

	flag.Var(&ns, "n", "")
	flag.Var(&ns, "ns", "")

	flag.BoolVar(&follow, "follow", false, "stream and follow logs?")
	flag.BoolVar(&follow, "f", false, "stream and follow logs?")

	flag.BoolVar(&taskHosts, "hostFilter", false, "filter tasks hostname? (default: selected tasks include all hosts)")

	flag.StringVar(&saveTo, "save", "", fmt.Sprintf("dir path to save logs (dir will be created, default: %s)", saveTo))
	flag.StringVar(&saveTo, "s", "", fmt.Sprintf("dir path to save logs (dir will be created, default: %s)", saveTo))

}

func main() {

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if len(sources) <= 0 {
		sources.Set("stderr")
	}

	// check origin value
	if len(origin) <= 0 {
		if follow {
			origin = "end"
		} else {
			origin = "start"
		}
	}

	if !utils.StrInArray(origin, []string{"start", "end"}) {
		log.Fatalln("--origin can only be 'start' or 'end'")
	}

	if !utils.StrInArray(output, []string{"stderr", "stdout"}) {
		log.Fatalln("--out can only be 'stderr' or 'stdout'")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := hashi.NomadClient()
	jobsc := client.Jobs()

	// select {namespace}.{job}
	jobStubs := nomadSelectJob(jobsc)

	selectedJobs := getSelectedJobs(jobsc, jobStubs)

	distinctStates := getDistinctTaskStatuses(selectedJobs)

	var states []string
	if len(distinctStates) <= 1 {
		states = []string{distinctStates[0]}
	} else {
		states = selectTaskStates(distinctStates)
	}

	fmt.Printf("Selected states: %v \n", states)

	// TODO: option to filter all allocation tasks by hostname
	// taskAllocMap := hashi.FilterTaskAllocsByStates(states, selectedJobs)
	// spew.Dump(taskAllocMap)

	// right now, get all tasks that match selected state,
	taskMap := hashi.DistinctTaskByStates(states, selectedJobs)
	selectedTasks := selectTasks(taskMap)

	// spew.Dump("Selected Tasks: ", selectedTasks)

	selectedAllocs := hashi.FilterJobSelectByStateAndTask(states, selectedTasks, selectedJobs)

  if len(selectedAllocs) <= 0 {
    log.Println("no allocations selected, exiting ...") 
    return
  }

	log.Printf("%d selected allocation(s)", len(selectedAllocs))


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

	// spew.Dump("selectedJobs", selectedJobs)

	var streamOutput io.Writer
	switch output {
	case "stderr":
		streamOutput = os.Stderr
		break
	case "stdout":
		streamOutput = os.Stdout
	}

	config := hashi.StreamLogsConfig{
		LogType:      sources,
		Follow:       follow,
		StreamWaiter: new(sync.WaitGroup),
		StreamOutput: streamOutput,
		SavePath:     saveTo,
		Origin:       origin,
		OriginOffset: offset,
	}

	logWriters, writersErr := hashi.NewLogWriters(&config)
	if writersErr != nil {
		log.Fatalf("error creating writers: %v", writersErr)
	}
	streamer := hashi.NewStreamLogs(client, selectedAllocs, &config, logWriters)

	go streamer.Run(ctx)

	<-ctx.Done()
}

func getDistinctTaskStatuses(jobs []*hashi.JobSelect) []string {

	r := make([]string, 0)

	for _, v := range jobs {
		for _, vv := range v.Allocs {
			for _, vvv := range vv.TaskStates {
				st := vvv.State
				if !utils.StrInArray(st, r) {
					r = append(r, st)
				}
			}
		}
	}

	return r
}

func getSelectedJobs(c hashi.NomadJobs, stubs []*nomad.JobListStub) []*hashi.JobSelect {

	selectedJobs := make([]*hashi.JobSelect, 0)

	for _, v := range stubs {

		qops := &nomad.QueryOptions{
			Namespace: v.Namespace,
		}

		allocs, _, allocsErr := c.Allocations(v.ID, false, qops)
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

func selectTaskStates(states []string) []string {

	var askRes []string

	sort.Slice(states, func(a, b int) bool {
		return states[a] < states[b]
	})

	sel := survey.MultiSelect{
		Message: "Select task states",
		Options: states,
	}

	survey.AskOne(&sel, &askRes, survey.WithPageSize(15))

	return askRes
}

func selectTasks(jt map[string]*hashi.JobTask) []*hashi.JobTask {
	r := make([]*hashi.JobTask, 0)
	keys := jobTaskKeys(jt)
	askRes := make([]string, 0)

	sel := survey.MultiSelect{
		Message: "Select task states",
		Options: keys,
	}

	survey.AskOne(&sel, &askRes, survey.WithPageSize(15))

	for _, v := range askRes {
		sjt := jt[v]
		r = append(r, sjt)
	}

	return r
}

func jobTaskKeys(jt map[string]*hashi.JobTask) []string {
	r := make([]string, 0)

	for k := range jt {
		r = append(r, k)
	}
	return r
}
