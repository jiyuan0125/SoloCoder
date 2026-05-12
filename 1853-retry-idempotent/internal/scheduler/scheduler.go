package scheduler

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"idempotent-retry/internal/executor"
	"idempotent-retry/internal/models"
	"idempotent-retry/internal/store"
)

type Scheduler struct {
	store       *store.Store
	executor    *executor.Executor
	taskQueue   chan string
	stopChan    chan struct{}
	wg          sync.WaitGroup
	randSource  *rand.Rand
}

func New(s *store.Store, e *executor.Executor) *Scheduler {
	return &Scheduler{
		store:      s,
		executor:   e,
		taskQueue:  make(chan string, 1000),
		stopChan:   make(chan struct{}),
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Scheduler) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}
	log.Printf("Scheduler started with %d workers", workerCount)
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) Enqueue(taskID string) {
	select {
	case s.taskQueue <- taskID:
	case <-s.stopChan:
	}
}

func (s *Scheduler) worker() {
	defer s.wg.Done()

	for {
		select {
		case taskID := <-s.taskQueue:
			s.processTask(taskID)
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) processTask(taskID string) {
	task, exists := s.store.GetTask(taskID)
	if !exists {
		log.Printf("Task %s not found, skipping", taskID)
		return
	}

	if task.Status != models.StatusPending {
		log.Printf("Task %s is not pending (status: %s), skipping", taskID, task.Status)
		return
	}

	s.store.SetTaskStatus(taskID, models.StatusRunning)

	attempt := 0
	for {
		attempt++
		s.store.IncrementAttempt(taskID)

		record := s.executor.Execute(task, attempt)
		s.store.AddExecutionRecord(taskID, record)

		if record.Success {
			task, _ = s.store.GetTask(taskID)
			task.Result = &models.TaskResult{
				StatusCode: record.StatusCode,
				Body:       record.Body,
				Success:    record.Success,
				Error:      record.Error,
			}
			s.store.SetTaskResult(taskID, task.Result)
			s.store.SetTaskStatus(taskID, models.StatusSuccess)
			log.Printf("Task %s succeeded on attempt %d", taskID, attempt)
			s.notifyCallbacks(taskID)
			return
		}

		if attempt >= task.MaxRetries {
			task, _ = s.store.GetTask(taskID)
			task.Result = &models.TaskResult{
				StatusCode: record.StatusCode,
				Body:       record.Body,
				Success:    record.Success,
				Error:      record.Error,
			}
			s.store.SetTaskResult(taskID, task.Result)
			s.store.SetTaskStatus(taskID, models.StatusFailed)
			log.Printf("Task %s failed after %d attempts", taskID, attempt)
			s.notifyCallbacks(taskID)
			return
		}

		interval := s.calculateInterval(task, attempt)
		log.Printf("Task %s attempt %d failed, retrying in %v", taskID, attempt, interval)
		s.store.SetTaskStatus(taskID, models.StatusPending)

		time.Sleep(interval)
	}
}

func (s *Scheduler) calculateInterval(task *models.Task, attempt int) time.Duration {
	base := task.BaseInterval

	if base <= 0 {
		base = 1 * time.Second
	}

	if task.RetryStrategy == models.StrategyFixed {
		return base
	}

	multiplier := 1 << (attempt - 1)
	exponentialInterval := base * time.Duration(multiplier)
	jitter := time.Duration(s.randSource.Int63n(int64(base) + 1))

	return exponentialInterval + jitter
}

func (s *Scheduler) notifyCallbacks(taskID string) {
	task, exists := s.store.GetTask(taskID)
	if !exists {
		return
	}

	for _, callbackURL := range task.Callbacks {
		go func(url string) {
			result := s.executor.SendCallback(url, task)
			if !result.Success {
				log.Printf("Callback to %s for task %s failed: %s", url, taskID, result.Error)
			}
		}(callbackURL)
	}
}
