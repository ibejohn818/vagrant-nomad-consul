package main

import (
	"context"
	"gotooling/johnhardy.io/pkg/hashi"
	"log"

	nomad "github.com/hashicorp/nomad/api"
)

func progress(_ *hashi.JobDeployMonitor, d *nomad.Deployment) {
	log.Printf("Deploy Status: %s \r", d.Status)
}

func deploy(ctx context.Context) {


	ncli := hashi.NomadClient()
	njobs := ncli.Jobs()
	nevals := ncli.Evaluations()
	ndeploy := ncli.Deployments()

	var err error

	ingressPath := "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/ingress.hcl"
	registryPath := "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/registry.hcl"

	ingressJob, err := hashi.JobFromHcl(ingressPath)
	if err != nil {
		panic(err)
	}

	registryJob, err := hashi.JobFromHcl(registryPath)
	if err != nil {
		panic(err)
	}

  mon := hashi.NewJobDeployMonitor(ingressJob, njobs, nevals, ndeploy)
  go mon.DeployWithCallbacks(ctx, progress, nil)

  mon2 := hashi.NewJobDeployMonitor(registryJob, njobs, nevals, ndeploy)
  go mon2.DeployWithCallbacks(ctx, progress, nil)


}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

  go deploy(ctx)

	<-ctx.Done()
}
