package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type SchedulerConfig struct {
	MaxConcurrency    int
	DefaultRetries    int
	DefaultRetryInterval int
}

type Scheduler struct {
	dag          *DAG
	config       SchedulerConfig
	tasks        map[string]*Task
	status       map[string]TaskStatus
	completed    map[string]bool
	failed       map[string]bool
	mu           sync.Mutex
	cond         *sync.Cond
	running      bool
	cancelled    bool
	wg           sync.WaitGroup
	semaphore    chan struct{}
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

func NewScheduler(dag *DAG, config SchedulerConfig) *Scheduler {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 4
	}
	if config.DefaultRetries < 0 {
		config.DefaultRetries = 0
	}
	if config.DefaultRetryInterval < 0 {
		config.DefaultRetryInterval = 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		dag:        dag,
		config:     config,
		tasks:      dag.GetTasks(),
		status:     make(map[string]TaskStatus),
		completed:  make(map[string]bool),
		failed:     make(map[string]bool),
		running:    false,
		cancelled:  false,
		semaphore:  make(chan struct{}, config.MaxConcurrency),
		ctx:        ctx,
		cancelFunc: cancel,
	}
	s.cond = sync.NewCond(&s.mu)

	for id := range s.tasks {
		s.status[id] = StatusPending
	}

	return s
}

func (s *Scheduler) GetTaskStatus(taskID string) (TaskStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status, exists := s.status[taskID]
	return status, exists
}

func (s *Scheduler) GetAllStatus() map[string]TaskStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]TaskStatus)
	for k, v := range s.status {
		result[k] = v
	}
	return result
}

func (s *Scheduler) GetTasks() map[string]*Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]*Task)
	for k, v := range s.tasks {
		taskCopy := *v
		result[k] = &taskCopy
	}
	return result
}

func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) Cancel() {
	s.mu.Lock()
	if s.cancelled {
		s.mu.Unlock()
		return
	}
	s.cancelled = true
	s.mu.Unlock()
	s.cancelFunc()
	s.cond.Broadcast()
}

func (s *Scheduler) allDependenciesCompleted(taskID string) bool {
	task, exists := s.tasks[taskID]
	if !exists {
		return false
	}
	for _, depID := range task.Dependencies {
		if !s.completed[depID] {
			return false
		}
	}
	return true
}

func (s *Scheduler) anyDependencyFailed(taskID string) bool {
	task, exists := s.tasks[taskID]
	if !exists {
		return false
	}
	for _, depID := range task.Dependencies {
		if s.failed[depID] {
			return true
		}
	}
	return false
}

func (s *Scheduler) getReadyTasks() []string {
	ready := []string{}
	for id, task := range s.tasks {
		if s.status[id] != StatusPending {
			continue
		}
		if s.anyDependencyFailed(id) {
			s.status[id] = StatusSkipped
			task.Status = StatusSkipped
			continue
		}
		if s.allDependenciesCompleted(id) {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	return ready
}

func (s *Scheduler) executeTask(taskID string) {
	defer s.wg.Done()
	defer func() { <-s.semaphore }()

	task := s.tasks[taskID]
	retries := task.Retries
	if retries < 0 {
		retries = s.config.DefaultRetries
	}
	retryInterval := task.RetryInterval
	if retryInterval <= 0 {
		retryInterval = s.config.DefaultRetryInterval
	}

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		select {
		case <-s.ctx.Done():
			s.mu.Lock()
			s.status[taskID] = StatusSkipped
			task.Status = StatusSkipped
			s.mu.Unlock()
			return
		default:
		}

		s.mu.Lock()
		task.Attempts = attempt + 1
		task.Status = StatusRunning
		s.status[taskID] = StatusRunning
		task.StartTime = time.Now().UnixMilli()
		s.mu.Unlock()

		duration := task.Duration
		if duration <= 0 {
			duration = 1 + int(time.Now().UnixNano()%5)
		}

		select {
		case <-s.ctx.Done():
			s.mu.Lock()
			task.EndTime = time.Now().UnixMilli()
			task.DurationMs = task.EndTime - task.StartTime
			s.status[taskID] = StatusSkipped
			task.Status = StatusSkipped
			s.mu.Unlock()
			return
		case <-time.After(time.Duration(duration) * time.Second):
		}

		if task.ShouldFail {
			lastErr = fmt.Errorf("task execution failed")
			task.LastError = lastErr.Error()
			task.EndTime = time.Now().UnixMilli()
			task.DurationMs = task.EndTime - task.StartTime

			if attempt < retries {
				s.mu.Lock()
				s.status[taskID] = StatusPending
				task.Status = StatusPending
				s.mu.Unlock()
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}

			s.mu.Lock()
			s.status[taskID] = StatusFailed
			task.Status = StatusFailed
			s.failed[taskID] = true
			s.mu.Unlock()
			s.cond.Broadcast()
			return
		}

		s.mu.Lock()
		task.EndTime = time.Now().UnixMilli()
		task.DurationMs = task.EndTime - task.StartTime
		s.status[taskID] = StatusCompleted
		task.Status = StatusCompleted
		s.completed[taskID] = true
		s.mu.Unlock()
		s.cond.Broadcast()
		return
	}

	s.mu.Lock()
	s.status[taskID] = StatusFailed
	task.Status = StatusFailed
	s.failed[taskID] = true
	if lastErr != nil {
		task.LastError = lastErr.Error()
	}
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Scheduler) Run() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.cancelled = false
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	for {
		s.mu.Lock()
		if s.cancelled {
			s.mu.Unlock()
			break
		}

		ready := s.getReadyTasks()
		allDone := true
		for id, status := range s.status {
			if status == StatusPending || status == StatusRunning {
				allDone = false
				break
			}
			_ = id
		}

		if allDone {
			s.mu.Unlock()
			break
		}

		if len(ready) == 0 {
			s.cond.Wait()
			s.mu.Unlock()
			continue
		}

		for _, taskID := range ready {
			select {
			case s.semaphore <- struct{}{}:
				s.mu.Lock()
				s.status[taskID] = StatusRunning
				s.tasks[taskID].Status = StatusRunning
				s.mu.Unlock()
				s.wg.Add(1)
				go s.executeTask(taskID)
			case <-s.ctx.Done():
				s.mu.Unlock()
				goto cleanup
			}
		}
		s.mu.Unlock()
	}

cleanup:
	s.wg.Wait()
	return nil
}
