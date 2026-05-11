package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type SchedulerConfig struct {
	MaxConcurrency       int
	DefaultRetries       int
	DefaultRetryInterval int
}

type Scheduler struct {
	dag        *DAG
	config     SchedulerConfig
	tasks      map[string]*Task
	mu         sync.RWMutex
	status     map[string]TaskStatus
	completed  map[string]bool
	failed     map[string]bool
	cond       *sync.Cond
	running    bool
	cancelled  bool
	wg         sync.WaitGroup
	semaphore  chan struct{}
	ctx        context.Context
	cancelFunc context.CancelFunc
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

	var mu sync.Mutex
	s.cond = sync.NewCond(&mu)

	s.mu.Lock()
	for id := range s.tasks {
		s.status[id] = StatusPending
	}
	s.mu.Unlock()

	return s
}

func (s *Scheduler) GetTaskStatus(taskID string) (TaskStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, exists := s.status[taskID]
	return status, exists
}

func (s *Scheduler) GetAllStatus() map[string]TaskStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]TaskStatus)
	for k, v := range s.status {
		result[k] = v
	}
	return result
}

func (s *Scheduler) GetTasks() map[string]*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*Task)
	for k, v := range s.tasks {
		taskCopy := *v
		result[k] = &taskCopy
	}
	return result
}

func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

func (s *Scheduler) isCancelled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cancelled
}

func (s *Scheduler) setRunning(val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = val
}

func (s *Scheduler) allDependenciesCompleted(taskID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
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
	s.mu.RLock()
	defer s.mu.RUnlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()

	ready := []string{}
	for id, task := range s.tasks {
		if s.status[id] != StatusPending {
			continue
		}

		anyFailed := false
		for _, depID := range task.Dependencies {
			if s.failed[depID] {
				anyFailed = true
				break
			}
		}
		if anyFailed {
			s.status[id] = StatusSkipped
			task.Status = StatusSkipped
			continue
		}

		allCompleted := true
		for _, depID := range task.Dependencies {
			if !s.completed[depID] {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	return ready
}

func (s *Scheduler) checkAllDone() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, status := range s.status {
		if status == StatusPending || status == StatusRunning {
			return false
		}
	}
	return true
}

func (s *Scheduler) markTaskRunning(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[taskID] = StatusRunning
	if task, exists := s.tasks[taskID]; exists {
		task.Status = StatusRunning
	}
}

func (s *Scheduler) executeTask(taskID string) {
	defer s.wg.Done()
	defer func() { <-s.semaphore }()

	s.mu.RLock()
	task, exists := s.tasks[taskID]
	if !exists {
		s.mu.RUnlock()
		return
	}
	retries := task.Retries
	if retries < 0 {
		retries = s.config.DefaultRetries
	}
	retryInterval := task.RetryInterval
	if retryInterval <= 0 {
		retryInterval = s.config.DefaultRetryInterval
	}
	shouldFail := task.ShouldFail
	s.mu.RUnlock()

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		select {
		case <-s.ctx.Done():
			s.mu.Lock()
			s.status[taskID] = StatusSkipped
			if t, ok := s.tasks[taskID]; ok {
				t.Status = StatusSkipped
			}
			s.mu.Unlock()
			return
		default:
		}

		now := time.Now().UnixMilli()
		s.mu.Lock()
		if t, ok := s.tasks[taskID]; ok {
			t.Attempts = attempt + 1
			t.Status = StatusRunning
			t.StartTime = now
		}
		s.status[taskID] = StatusRunning
		s.mu.Unlock()

		s.mu.RLock()
		duration := task.Duration
		if duration <= 0 {
			duration = 1 + int(time.Now().UnixNano()%5)
		}
		s.mu.RUnlock()

		select {
		case <-s.ctx.Done():
			endNow := time.Now().UnixMilli()
			s.mu.Lock()
			if t, ok := s.tasks[taskID]; ok {
				t.EndTime = endNow
				t.DurationMs = endNow - now
				t.Status = StatusSkipped
			}
			s.status[taskID] = StatusSkipped
			s.mu.Unlock()
			return
		case <-time.After(time.Duration(duration) * time.Second):
		}

		if shouldFail {
			lastErr = fmt.Errorf("task execution failed")
			endNow := time.Now().UnixMilli()

			s.mu.Lock()
			if t, ok := s.tasks[taskID]; ok {
				t.LastError = lastErr.Error()
				t.EndTime = endNow
				t.DurationMs = endNow - now
			}
			s.mu.Unlock()

			if attempt < retries {
				s.mu.Lock()
				s.status[taskID] = StatusPending
				if t, ok := s.tasks[taskID]; ok {
					t.Status = StatusPending
				}
				s.mu.Unlock()
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}

			s.mu.Lock()
			s.status[taskID] = StatusFailed
			s.failed[taskID] = true
			if t, ok := s.tasks[taskID]; ok {
				t.Status = StatusFailed
			}
			s.mu.Unlock()
			s.cond.Broadcast()
			return
		}

		endNow := time.Now().UnixMilli()
		s.mu.Lock()
		if t, ok := s.tasks[taskID]; ok {
			t.EndTime = endNow
			t.DurationMs = endNow - now
			t.Status = StatusCompleted
		}
		s.status[taskID] = StatusCompleted
		s.completed[taskID] = true
		s.mu.Unlock()
		s.cond.Broadcast()
		return
	}

	s.mu.Lock()
	s.status[taskID] = StatusFailed
	s.failed[taskID] = true
	if t, ok := s.tasks[taskID]; ok {
		t.Status = StatusFailed
		if lastErr != nil {
			t.LastError = lastErr.Error()
		}
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

	defer s.setRunning(false)

	for {
		if s.isCancelled() {
			break
		}

		ready := s.getReadyTasks()
		allDone := s.checkAllDone()

		if allDone {
			break
		}

		if len(ready) == 0 {
			s.cond.L.Lock()
			s.cond.Wait()
			s.cond.L.Unlock()
			continue
		}

		for _, taskID := range ready {
			if s.isCancelled() {
				break
			}

			select {
			case s.semaphore <- struct{}{}:
				s.markTaskRunning(taskID)
				s.wg.Add(1)
				go s.executeTask(taskID)
			case <-s.ctx.Done():
				goto cleanup
			}
		}
	}

cleanup:
	s.wg.Wait()
	return nil
}
