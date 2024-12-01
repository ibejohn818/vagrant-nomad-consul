package main

import (
	"context"
	"flag"
	"log"
	"os"
)

var (
	jobsJSON = flag.String("jobJson", "", "path to jobs.json file")
)

func main() {

	flag.Parse()

	if len(*jobsJSON) > 0 {

	}

	menu := jobMenu{
		state: jobMenuStateHome,
		jobs:  make([]*JobInput, 0),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runMenu(ctx, &menu)

	<-ctx.Done()

}

func runMenu(_ context.Context, menu *jobMenu) {
	for {

    clearTerm()

		switch menu.state {
		case jobMenuStateHome:
			homeMenu(menu)
			break
		case jobMenuStateExit:
			os.Exit(0)
			break
    case jobMenuStateChooseJob:

    break
		default:
			log.Println("invalid jobMenu state")
			os.Exit(1)
			break
		}
	}
}
