package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go-batch-scheduler/db"
	"go-batch-scheduler/models"
)

type Scheduler struct {
	db          *db.Database
	workerCount int
	taskQueue   *PriorityQueue
	mu          sync.Mutex
	running     bool
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

type PQItem struct {
	taskID   string
	priority models.Priority
	index    int
}

type PriorityQueue []*PQItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PQItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func New(database *db.Database, workerCount int) *Scheduler {
	if workerCount <= 0 {
		workerCount = 4
	}
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	return &Scheduler{
		db:          database,
		workerCount: workerCount,
		taskQueue:   &pq,
		running:     false,
		stopChan:    make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true
	s.stopChan = make(chan struct{})

	s.loadPendingTasks()

	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	go s.monitorNewTasks()

	log.Printf("Scheduler started with %d workers", s.workerCount)
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) loadPendingTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := s.db.GetPendingTasks(ctx)
	if err != nil {
		log.Printf("Error loading pending tasks: %v", err)
		return
	}

	for _, task := range tasks {
		heap.Push(s.taskQueue, &PQItem{
			taskID:   task.ID,
			priority: task.Priority,
		})
	}

	log.Printf("Loaded %d pending tasks into queue", len(tasks))
}

func (s *Scheduler) monitorNewTasks() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndEnqueueNewTasks()
		}
	}
}

func (s *Scheduler) checkAndEnqueueNewTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := s.db.GetPendingTasks(ctx)
	if err != nil {
		log.Printf("Error checking new tasks: %v", err)
		return
	}

	currentInQueue := make(map[string]bool)
	for _, item := range *s.taskQueue {
		currentInQueue[item.taskID] = true
	}

	for _, task := range tasks {
		if !currentInQueue[task.ID] {
			heap.Push(s.taskQueue, &PQItem{
				taskID:   task.ID,
				priority: task.Priority,
			})
			log.Printf("Enqueued task: %s", task.ID)
		}
	}
}

func (s *Scheduler) AddTask(taskID string, priority models.Priority) {
	s.mu.Lock()
	defer s.mu.Unlock()

	heap.Push(s.taskQueue, &PQItem{
		taskID:   taskID,
		priority: priority,
	})
}

func (s *Scheduler) worker(id int) {
	defer s.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-s.stopChan:
			return
		default:
			taskID := s.getNextTask()
			if taskID == "" {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			s.executeTask(id, taskID)
		}
	}
}

func (s *Scheduler) getNextTask() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.taskQueue.Len() == 0 {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tasks, err := s.db.GetPendingTasks(ctx)
	if err != nil {
		log.Printf("Error getting pending tasks: %v", err)
		return ""
	}

	for _, task := range tasks {
		ok, err := s.db.AreAllDependenciesSuccessful(task.ID)
		if err != nil {
			log.Printf("Error checking dependencies for %s: %v", task.ID, err)
			continue
		}

		if ok {
			s.removeFromQueue(task.ID)
			return task.ID
		}
	}

	return ""
}

func (s *Scheduler) removeFromQueue(taskID string) {
	for i, item := range *s.taskQueue {
		if item.taskID == taskID {
			heap.Remove(s.taskQueue, i)
			return
		}
	}
}

func (s *Scheduler) executeTask(workerID int, taskID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	task, err := s.db.GetTask(taskID)
	cancel()

	if err != nil || task == nil {
		log.Printf("Worker %d: Task %s not found: %v", workerID, taskID, err)
		return
	}

	log.Printf("Worker %d: Starting task %s", workerID, task.ID)

	now := time.Now()
	task.Status = models.StatusRunning
	task.FlowStatus = models.FlowExecuting
	task.StartedAt = &now

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.db.UpdateTask(ctx, task); err != nil {
		cancel()
		log.Printf("Worker %d: Failed to update task %s: %v", workerID, task.ID, err)
		return
	}
	cancel()

	timeout := task.Timeout
	if timeout < 1*time.Second {
		timeout = 30 * time.Second
	}

	done := make(chan bool, 1)
	var failedReason string

	go func() {
		defer close(done)
		err := s.runTaskLogic(task)
		if err != nil {
			failedReason = err.Error()
			done <- false
			return
		}
		done <- true
	}()

	select {
	case success := <-done:
		now = time.Now()
		task.CompletedAt = &now
		if success {
			task.Status = models.StatusSuccess
			task.FlowStatus = models.FlowCompleted
			task.FailedReason = ""
			task.IsFinalFailure = false
			log.Printf("Worker %d: Task %s completed successfully", workerID, task.ID)
		} else {
			s.handleFailure(task, failedReason)
			log.Printf("Worker %d: Task %s failed: %s", workerID, task.ID, failedReason)
		}
	case <-time.After(timeout):
		now = time.Now()
		task.CompletedAt = &now
		s.handleFailure(task, fmt.Sprintf("timeout after %v", timeout))
		log.Printf("Worker %d: Task %s timed out after %v", workerID, task.ID, timeout)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.db.UpdateTask(ctx, task); err != nil {
		cancel()
		log.Printf("Worker %d: Failed to finalize task %s: %v", workerID, task.ID, err)
		return
	}
	cancel()
}

func (s *Scheduler) runTaskLogic(task *models.Task) error {
	switch task.Type {
	case "data_processing":
		time.Sleep(2 * time.Second)
		return nil
	case "file_conversion":
		time.Sleep(3 * time.Second)
		return nil
	case "quick":
		time.Sleep(500 * time.Millisecond)
		return nil
	case "long":
		time.Sleep(10 * time.Second)
		return nil
	default:
		time.Sleep(1 * time.Second)
		return nil
	}
}

func (s *Scheduler) handleFailure(task *models.Task, reason string) {
	task.FailedReason = reason

	if task.RetryCount < task.MaxRetries {
		task.RetryCount++
		task.Status = models.StatusQueued
		task.FlowStatus = models.FlowApproved
		task.StartedAt = nil
		task.CompletedAt = nil
		task.IsFinalFailure = false

		s.AddTask(task.ID, task.Priority)
		log.Printf("Task %s scheduled for retry %d/%d", task.ID, task.RetryCount, task.MaxRetries)
	} else {
		task.Status = models.StatusFailed
		task.FlowStatus = models.FlowCompleted
		task.IsFinalFailure = true
	}
}

func (s *Scheduler) RetryTask(taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	task, err := s.db.GetTask(taskID)
	cancel()

	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	if task.Status != models.StatusFailed {
		return fmt.Errorf("only failed tasks can be retried")
	}

	task.RetryCount = 0
	task.Status = models.StatusQueued
	task.FlowStatus = models.FlowApproved
	task.StartedAt = nil
	task.CompletedAt = nil
	task.FailedReason = ""
	task.IsFinalFailure = false

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.db.UpdateTask(ctx, task); err != nil {
		cancel()
		return err
	}
	cancel()

	s.AddTask(task.ID, task.Priority)
	return nil
}
