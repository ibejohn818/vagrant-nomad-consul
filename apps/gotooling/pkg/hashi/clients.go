package hashi

import (
	consul "github.com/hashicorp/consul/api"
	nomad "github.com/hashicorp/nomad/api"
)

var (
  nomadJobs NomadJobs
  nomadEvaluations NomadEvaluations
  nomadDeployments NomadDeployments
)

func SetNomadServices(client *nomad.Client) {
  nomadJobs = client.Jobs()
  nomadEvaluations = client.Evaluations()
  nomadDeployments = client.Deployments()
}

type NomadJobs interface {
	Plan(job *nomad.Job, diff bool, q *nomad.WriteOptions) (*nomad.JobPlanResponse, *nomad.WriteMeta, error)
	Register(job *nomad.Job, q *nomad.WriteOptions) (*nomad.JobRegisterResponse, *nomad.WriteMeta, error)
}

type NomadEvaluations interface {
	Info(evalID string, q *nomad.QueryOptions) (*nomad.Evaluation, *nomad.QueryMeta, error)
}

type NomadDeployments interface {
	Info(deploymentID string, q *nomad.QueryOptions) (*nomad.Deployment, *nomad.QueryMeta, error)
}

func NomadClient() *nomad.Client {

	defConfig := nomad.DefaultConfig()
	defConfig.TLSConfig.Insecure = true

	client, err := nomad.NewClient(defConfig)
	if err != nil {
		panic(err)
	}

	return client
}

func ConsulClient() *consul.Client {
	defConfig := consul.DefaultConfig()
	defConfig.TLSConfig.InsecureSkipVerify = true

	client, err := consul.NewClient(defConfig)
	if err != nil {
		panic(err)
	}

	return client
}
