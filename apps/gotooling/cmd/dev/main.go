package main

import (
	"fmt"
	"gotooling/johnhardy.io/pkg/hashi"
	"time"
)

func main() {

	ncli := hashi.NomadClient()
	njobs := ncli.Jobs()
  nevals := ncli.Evaluations()
  ndeploy := ncli.Deployments()

	jobPath := "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/ingress.hcl"

	job, err := hashi.ParseJob(jobPath)

	if err != nil {
		panic(err)
	}

	plan, planErr := hashi.PlanJob(njobs, job)

	if planErr != nil {
		panic(planErr)
	}

	fmt.Printf("Plan index: %d \n", plan.JobModifyIndex)

	regRes, regErr := hashi.RunJob(njobs, job, "")

	if regErr != nil {
		panic(regErr)
	}

	fmt.Printf("EvalID: %s \n", regRes.EvalID)

  deployID, evalErr := hashi.TryGetDeploymentID(nevals, regRes.EvalID, 0)
	if evalErr != nil {
		panic(evalErr)
	}

	for i := 0; i < 20; i++ {
    hashi.DeployInfo(ndeploy, deployID)
		time.Sleep(1 * time.Second)
	}

}
