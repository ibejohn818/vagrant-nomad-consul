package hashi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gotooling/johnhardy.io/pkg/utils"

	nomad "github.com/hashicorp/nomad/api"
	jobspec "github.com/hashicorp/nomad/jobspec2"
	"github.com/pkg/errors"
)

const defaultMaxAttemps uint32 = 30

// deploy callbacks
type JobDeployMonitorProgress func(*JobDeployMonitor, *nomad.Deployment)
type JobDeployMonitorCompleted func(*JobDeployMonitor, *nomad.Deployment)

var (
	defaultEvalWait time.Duration = 250 * time.Millisecond
)

type JobTask struct {
	Namespace string
	Job       string
	Task      string
}

type JobTaskAlloc struct {
	Namespace string
	Job       string
	Task      string
	AllocID   string
	Host      string
}

func DistinctTaskByStates(states []string, jobs []*JobSelect) map[string]*JobTask {
	m := make(map[string]*JobTask)

	for _, v := range jobs {
		for _, vv := range v.Allocs {
			for taskName, vvv := range vv.TaskStates {
				if !utils.StrInArray(vvv.State, states) {
					continue
				}
				key := fmt.Sprintf("%s.%s.%s", v.Stub.Namespace, v.Stub.ID, taskName)
				m[key] = &JobTask{
					Namespace: v.Stub.Namespace,
					Job:       v.Stub.ID,
					Task:      taskName,
				}
			}
		}
	}

	return m
}

func FilterTaskAllocsByStates(states []string, jobs []*JobSelect) map[string]*JobTaskAlloc {
	m := make(map[string]*JobTaskAlloc)

	for _, v := range jobs {

		jobID := strings.Replace(v.Stub.ID, ".", "-", -1)
		ns := strings.Replace(v.Stub.Namespace, ".", "-", -1)

		for _, vv := range v.Allocs {

			host := vv.NodeName

			for taskName, vvv := range vv.TaskStates {
				st := vvv.State
				if utils.StrInArray(st, states) {
					mapKey := fmt.Sprintf("%s.%s.%s", ns, jobID, taskName)
					m[mapKey] = &JobTaskAlloc{
						Namespace: v.Stub.Namespace,
						Job:       v.Stub.ID,
						Task:      taskName,
						AllocID:   vv.ID,
						Host:      host,
					}
				}
			}
		}
	}

	return m
}

func IsInJobTasks(namespace, jobId, task string, jts []*JobTask) bool {
  
  for _, v := range jts {
    nsChk := v.Namespace == namespace
    jobIdChk := v.Job == jobId
    taskChk := v.Task == task

    if nsChk && jobIdChk && taskChk {
      return true
    }

  }


  return false
}

func FilterJobSelectByStateAndTask(states []string, jts []*JobTask, jobs []*JobSelect) []*JobTaskAlloc {
  r := make([]*JobTaskAlloc, 0)

  for _, v := range jobs {
    for _, vv := range v.Allocs {
      for taskName, vvv := range vv.TaskStates {
        jobTaskChk := IsInJobTasks(v.Stub.Namespace, v.Stub.ID, taskName, jts)
        if jobTaskChk && utils.StrInArray(vvv.State, states) {
          r = append(r, &JobTaskAlloc{
            Namespace: v.Stub.Namespace,
            Job: v.Stub.ID,
            AllocID: vv.ID,
            Task: taskName, 
            Host: vv.NodeName,
          })
        }
      }
    }
  }

  return r
}

/*
nomad deployment status vars
const (
	DeploymentStatusRunning    = "running"
	DeploymentStatusPaused     = "paused"
	DeploymentStatusFailed     = "failed"
	DeploymentStatusSuccessful = "successful"
	DeploymentStatusCancelled  = "cancelled"
	DeploymentStatusPending    = "pending"
	DeploymentStatusBlocked    = "blocked"
	DeploymentStatusUnblocking = "unblocking"
)
*/

type jobDeployMonitorCallbacks struct {
	progress  JobDeployMonitorProgress
	completed JobDeployMonitorCompleted
}

type JobDeployMonitor struct {
	//
	job      *nomad.Job
	evalID   string
	deployID string
	status   string
	err      error

	// interfaces
	nomadJobs    NomadJobs
	nomadEvals   NomadEvaluations
	nomadDeploys NomadDeployments

	// message buffer
	msg bytes.Buffer

	// save deploy eval
	deployEval *nomad.Evaluation
	// save final deployment info
	deployLookup *nomad.Deployment
	callbacks    *jobDeployMonitorCallbacks
}

func NewJobDeployMonitor(
	job *nomad.Job,
	nomadJobs NomadJobs,
	nomadEvals NomadEvaluations,
	nomadDeploys NomadDeployments) *JobDeployMonitor {

	d := JobDeployMonitor{
		status:       nomad.DeploymentStatusPending,
		job:          job,
		nomadJobs:    nomadJobs,
		nomadEvals:   nomadEvals,
		nomadDeploys: nomadDeploys,
		callbacks: &jobDeployMonitorCallbacks{
			progress:  nil,
			completed: nil,
		},
	}
	return &d
}

func (j *JobDeployMonitor) JobName() string {
	if j.job != nil && j.job.Name != nil {
		return *j.job.Name
	}

	return "n/a"
}

func (j *JobDeployMonitor) Status() string {
	return j.status
}

func (j *JobDeployMonitor) Deploy(ctx context.Context) {
	go j.DeployWithCallbacks(ctx, nil, nil)
}

func (j *JobDeployMonitor) DeployWithCallbacks(
	ctx context.Context,
	prog JobDeployMonitorProgress,
	completed JobDeployMonitorCompleted) {
	j.handleDeploy(ctx, prog, completed)
}

func (j *JobDeployMonitor) handleDeploy(
	_ context.Context,
	prog JobDeployMonitorProgress,
	completed JobDeployMonitorCompleted) {

	j.callbacks.progress = prog
	j.callbacks.completed = completed

	var err error

	// TODO: expose namespace param
	jobRes, err := JobRegister(j.nomadJobs, j.job, "")

	if err != nil {
		j.err = err
		j.status = nomad.DeploymentStatusFailed
		if completed != nil {
			completed(j, nil)
		}
		return
	}

	j.evalID = jobRes.EvalID
	j.deployID, err = TryGetDeploymentID(j.nomadEvals, j.evalID, 0)

	if err != nil {
		j.err = err
		j.status = nomad.DeploymentStatusFailed
		if completed != nil {
			completed(j, nil)
		}
		return
	}

	j.watchDeployment()
}

func (j *JobDeployMonitor) watchDeployment() {

	// interval := time.Duration(500 * time.Millisecond)
	interval := time.Duration(1000 * time.Millisecond)
	expireTime := time.Now().Add(5 * time.Minute)
	var err error

	for {

		j.deployLookup, err = DeployInfo(j.nomadDeploys, j.deployID)
		j.status = j.deployLookup.Status

		if j.callbacks.progress != nil {
			j.callbacks.progress(j, j.deployLookup)
		}

		if err != nil {
			j.err = err
			j.status = nomad.DeploymentStatusFailed
			if j.callbacks.completed != nil {
				j.callbacks.completed(j, j.deployLookup)
			}
			return
		}

		if time.Now().After(expireTime) {

			j.err = errors.New("deployment timed out")
			j.status = nomad.DeploymentStatusFailed
			if j.callbacks.completed != nil {
				j.callbacks.completed(j, j.deployLookup)
			}
			return
		}

		continueStatus := []string{
			nomad.DeploymentStatusPending,
			nomad.DeploymentStatusRunning,
		}

		if !utils.StrInArray(j.deployLookup.Status, continueStatus) {
			if j.callbacks.completed != nil {
				j.callbacks.completed(j, j.deployLookup)
			}
			return
		}

		time.Sleep(interval)
	}
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
func JobFromHcl(hclPath string) (*nomad.Job, error) {

	fd, err := os.Open(hclPath)
	if err != nil {
		return nil, err
	}

	// hcl, hclErr := jobspec.Parse(hclPath, fd)
	//  if hclErr != nil {
	//    panic(hclErr)
	//  }

	jobBytes, _ := io.ReadAll(fd)
	c := jobspec.ParseConfig{
		Path:    hclPath,
		BaseDir: filepath.Dir(hclPath),
		Body:    jobBytes,
		AllowFS: true,
	}
	job, jobErr := jobspec.ParseWithConfig(&c)

	return job, jobErr
}

func PlanJob(jobs NomadJobs, job *nomad.Job) (*nomad.JobPlanResponse, error) {

	ops := nomad.WriteOptions{}

	res, _, err := jobs.Plan(job, true, &ops)

	return res, err
}

func JobRegister(jobs NomadJobs, job *nomad.Job, ns string) (*nomad.JobRegisterResponse, error) {

	if len(ns) <= 0 {
		ns = "default"
	}

	ops := nomad.WriteOptions{
		Namespace: ns,
	}

	res, _, err := jobs.Register(job, &ops)

	return res, err
}

func DeployInfo(deps NomadDeployments, deployID string) (*nomad.Deployment, error) {

	ops := nomad.QueryOptions{}

	res, _, err := deps.Info(deployID, &ops)

	return res, err
}

func JobList(jobs NomadJobs) ([]*nomad.JobListStub, error) {

	ops := nomad.QueryOptions{
		Namespace: "*",
	}

	res, _, err := jobs.List(&ops)

	return res, err
}

func JobInfo(jcli NomadJobs, id, namespace string) (*nomad.Job, error) {
	ops := nomad.QueryOptions{
		Namespace: namespace,
	}

	res, _, err := jcli.Info(id, &ops)

	return res, err
}
