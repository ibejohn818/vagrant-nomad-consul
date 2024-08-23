package hashi

import (
	"bytes"
	"context"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	nomad "github.com/hashicorp/nomad/api"
	jobspec "github.com/hashicorp/nomad/jobspec2"
	"github.com/pkg/errors"
)

const defaultMaxAttemps uint32 = 30

var (
	defaultEvalWait time.Duration = 250 * time.Millisecond
)

type DeployStatus string

const (
	DeployStatusPending   = DeployStatus("pending")
	DeployStatusRunning   = DeployStatus("running")
	DeployStatusCompleted = DeployStatus("Completed")
	DeployStatusError     = DeployStatus("error")
)

type JobDeployMonitor struct {
	// inner context derived from incoming ctx
	ctx       context.Context
	ctxCancel context.CancelFunc
	//
	job      *nomad.Job
	evalID   string
	deployID string
	status   DeployStatus
	err      error

	// interfaces
	nomadJobs    NomadJobs
	nomadEvals   NomadEvaluations
	nomadDeploys NomadDeployments

	// message buffer
	msg bytes.Buffer
}

func NewJobDeployMonitor(
	ctx context.Context,
	job *nomad.Job,
	nomadJobs NomadJobs,
	nomadEvals NomadEvaluations,
	nomadDeploys NomadDeployments) *JobDeployMonitor {

	ctx, ctxCancel := context.WithCancel(ctx)

	d := JobDeployMonitor{
		ctx:          ctx,
		ctxCancel:    ctxCancel,
		status:       DeployStatusPending,
		job:          job,
		nomadJobs:    nomadJobs,
		nomadEvals:   nomadEvals,
		nomadDeploys: nomadDeploys,
	}
	return &d
}

func JobEval(evals NomadEvaluations, evalID string) (*nomad.Evaluation, error) {
	ops := nomad.QueryOptions{}

	res, _, err := evals.Info(evalID, &ops)

	return res, err
}

func TryGetDeploymentID(evals NomadEvaluations, evalID string, maxAttempts uint32) (string, error) {

	if maxAttempts == 0 {
		maxAttempts = defaultMaxAttemps
	}

	var attempts uint32 = 0

	for attempts < maxAttempts {

		res, err := JobEval(evals, evalID)

		if err != nil {
			return "", err
		}

		if len(res.DeploymentID) > 0 {
			return res.DeploymentID, nil
		}

		time.Sleep(defaultEvalWait)
		attempts += 1
	}

	return "", errors.New("unable to get deploymentID")

}

// ParseJob parses HCL file to structure for registering
// VIA sdk
func ParseJob(hclPath string) (*nomad.Job, error) {

	fd, err := os.Open(hclPath)
	if err != nil {
		return nil, err
	}

	job, jobErr := jobspec.Parse(hclPath, fd)

	return job, jobErr
}

func PlanJob(jobs NomadJobs, job *nomad.Job) (*nomad.JobPlanResponse, error) {

	ops := nomad.WriteOptions{}

	res, _, err := jobs.Plan(job, true, &ops)

	return res, err
}

func RunJob(jobs NomadJobs, job *nomad.Job, namespace string) (*nomad.JobRegisterResponse, error) {

	if len(namespace) <= 0 {
		namespace = "default"
	}

	ops := nomad.WriteOptions{
		Namespace: namespace,
	}

	res, _, err := jobs.Register(job, &ops)

	return res, err
}

func DeployInfo(deps NomadDeployments, deployID string) {

	ops := nomad.QueryOptions{}

	res, meta, err := deps.Info(deployID, &ops)

	spew.Dump("Meta: ", meta)
	spew.Dump("Res: ", res)
	spew.Dump("Err: ", err)
}
