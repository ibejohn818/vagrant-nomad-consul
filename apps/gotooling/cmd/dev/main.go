package main

import (
	"context"
	"gotooling/johnhardy.io/pkg/hashi"
	"log"

	nomad "github.com/hashicorp/nomad/api"
)

func progress(jdm *hashi.JobDeployMonitor, d *nomad.Deployment) {
  name := jdm.JobName()
	log.Printf("%s Deploy Status: %s \r", name,  d.Status)
}

// ingressPath := "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/ingress.hcl"
// registryPath := "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/registry.hcl"
// promPath := "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/prometheus.hcl"
var jobs map[string]string = map[string]string{
	"ingress":    "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/ingress.hcl",
	"registry":   "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/registry.hcl",
	"prometheus": "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/prometheus.hcl",
	"grafana": "/home/jhardy/projects/lab/vagrant-nomad-consul/jobs/grafana.hcl",
}

var ncli *nomad.Client = hashi.NomadClient()
var njobs hashi.NomadJobs = ncli.Jobs()
var nevals hashi.NomadEvaluations = ncli.Evaluations()
var ndeploy hashi.NomadDeployments = ncli.Deployments()

func deploy(ctx context.Context, hclPath string) {

	var err error

	jobData, err := hashi.JobFromHcl(hclPath)
	if err != nil {
		panic(err)
	}

	mon := hashi.NewJobDeployMonitor(jobData, njobs, nevals, ndeploy)
	go mon.DeployWithCallbacks(ctx, progress, nil)

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for k, v := range jobs {
		log.Println("Deploy: ", k)
		go deploy(ctx, v)
	}

	<-ctx.Done()
}
