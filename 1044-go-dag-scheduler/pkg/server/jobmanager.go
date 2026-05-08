package server

import (
	"dag-scheduler/common"
	"dag-scheduler/core"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID        string
	DAG       *core.DAG
	Scheduler *core.Scheduler
	Config    core.SchedulerConfig
	Tasks     []*core.Task
	mu        sync.Mutex
}

type JobManager struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*Job),
	}
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func (jm *JobManager) Submit(req common.SubmitRequest) (*common.SubmitResponse, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	tasks := make([]*core.Task, len(req.Tasks))
	for i, td := range req.Tasks {
		tasks[i] = &core.Task{
			ID:            td.ID,
			Dependencies:  td.Dependencies,
			Duration:      td.Duration,
			Retries:       td.Retries,
			RetryInterval: td.RetryInterval,
			Status:        core.StatusPending,
			ShouldFail:    td.ShouldFail,
		}
	}

	dag := core.NewDAG()
	if err := dag.Build(tasks); err != nil {
		return nil, err
	}

	if cycle, hasCycle := dag.DetectCycle(); hasCycle {
		return &common.SubmitResponse{
			Success: false,
			Error:   fmt.Sprintf("cycle detected in task graph"),
			Cycle:   cycle,
		}, nil
	}

	jobID := generateJobID()
	config := core.SchedulerConfig{
		MaxConcurrency:        req.MaxConcurrency,
		DefaultRetries:        req.DefaultRetries,
		DefaultRetryInterval:  req.DefaultRetryInterval,
	}

	scheduler := core.NewScheduler(dag, config)

	job := &Job{
		ID:        jobID,
		DAG:       dag,
		Scheduler: scheduler,
		Config:    config,
		Tasks:     tasks,
	}

	jm.jobs[jobID] = job

	return &common.SubmitResponse{
		Success: true,
		JobID:   jobID,
	}, nil
}

func (jm *JobManager) GetJob(jobID string) (*Job, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	job, exists := jm.jobs[jobID]
	return job, exists
}

func (jm *JobManager) Start(jobID string) (*common.StartResponse, error) {
	job, exists := jm.GetJob(jobID)
	if !exists {
		return &common.StartResponse{
			Success: false,
			Error:   fmt.Sprintf("job %s not found", jobID),
		}, nil
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	if job.Scheduler.IsRunning() {
		return &common.StartResponse{
			Success: false,
			Error:   "job is already running",
		}, nil
	}

	go func() {
		_ = job.Scheduler.Run()
	}()

	return &common.StartResponse{
		Success: true,
	}, nil
}

func (jm *JobManager) GetStatus(jobID string) (*common.StatusResponse, error) {
	job, exists := jm.GetJob(jobID)
	if !exists {
		return &common.StatusResponse{
			Success: false,
			Error:   fmt.Sprintf("job %s not found", jobID),
		}, nil
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	tasksMap := job.Scheduler.GetTasks()
	taskInfos := make(map[string]common.TaskInfo)

	for id, task := range tasksMap {
		taskInfos[id] = common.TaskInfo{
			ID:           id,
			Status:       task.Status,
			StartTime:    task.StartTime,
			EndTime:      task.EndTime,
			DurationMs:   task.DurationMs,
			Attempts:     task.Attempts,
			LastError:    task.LastError,
			Dependencies: task.Dependencies,
		}
	}

	return &common.StatusResponse{
		Success:   true,
		JobID:     jobID,
		IsRunning: job.Scheduler.IsRunning(),
		Tasks:     taskInfos,
	}, nil
}

func (jm *JobManager) Cancel(jobID string) (*common.CancelResponse, error) {
	job, exists := jm.GetJob(jobID)
	if !exists {
		return &common.CancelResponse{
			Success: false,
			Error:   fmt.Sprintf("job %s not found", jobID),
		}, nil
	}

	job.Scheduler.Cancel()

	return &common.CancelResponse{
		Success: true,
	}, nil
}
