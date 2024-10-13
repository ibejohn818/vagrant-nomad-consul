package hashi

import (
	consul "github.com/hashicorp/consul/api"
	nomad "github.com/hashicorp/nomad/api"
)

var (
	nomadJobs        NomadJobs
	nomadEvaluations NomadEvaluations
	nomadDeployments NomadDeployments
	nomadNamespaces  NomadNamespaces
)

func SetNomadServices(client *nomad.Client) (NomadJobs, NomadEvaluations, NomadDeployments, NomadNamespaces) {
	nomadJobs = client.Jobs()
	nomadEvaluations = client.Evaluations()
	nomadDeployments = client.Deployments()
	nomadNamespaces = client.Namespaces()

	return nomadJobs, nomadEvaluations, nomadDeployments, nomadNamespaces
}

type NomadAllocations interface {
	Info(allocID string, q *nomad.QueryOptions) (*nomad.Allocation, *nomad.QueryMeta, error)
}

type NomadAllocFS interface {
	Logs(alloc *nomad.Allocation, follow bool, task, logType, origin string,
		offset int64, cancel <-chan struct{}, q *nomad.QueryOptions) (<-chan *nomad.StreamFrame, <-chan error)
}

type NomadJobs interface {
	Plan(job *nomad.Job, diff bool, q *nomad.WriteOptions) (*nomad.JobPlanResponse, *nomad.WriteMeta, error)
	Register(job *nomad.Job, q *nomad.WriteOptions) (*nomad.JobRegisterResponse, *nomad.WriteMeta, error)
	List(q *nomad.QueryOptions) ([]*nomad.JobListStub, *nomad.QueryMeta, error)
	Info(jobID string, q *nomad.QueryOptions) (*nomad.Job, *nomad.QueryMeta, error)
	Allocations(jobID string, allAllocs bool, q *nomad.QueryOptions) ([]*nomad.AllocationListStub, *nomad.QueryMeta, error)
}

type NomadNamespaces interface {
	List(q *nomad.QueryOptions) ([]*nomad.Namespace, *nomad.QueryMeta, error)
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
