package taskqueue

import (
	"container/heap"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrQueueStopped = errors.New("queue is stopped")
)

var retryIntervals = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	5 * time.Second,
	10 * time.Second,
}

type Task struct {
	ID         uint64
	Priority   int
	Delay      time.Duration
	MaxRetries int
	Handler    func() error

	retryCount   int
	scheduledAt  time.Time
	lastErr      error
}

type TaskQueue struct {
	priorityQueue priorityQueue
	delayQueue    delayQueue

	deadLetters []*Task

	stats struct {
		submitted   uint64
		completed   uint64
		retries     uint64
		deadLetters uint64
	}

	mu           sync.Mutex
	cond         *sync.Cond
	workers      []*worker
	workerCount  int
	isRunning    bool
	isStopped    bool
	stopChan     chan struct{}
	stopWaitChan chan struct{}

	nextTaskID uint64
}

func NewTaskQueue() *TaskQueue {
	q := &TaskQueue{
		priorityQueue: make(priorityQueue, 0),
		delayQueue:    make(delayQueue, 0),
		deadLetters:   make([]*Task, 0),
		stopChan:      make(chan struct{}),
		stopWaitChan:  make(chan struct{}),
	}
	q.cond = sync.NewCond(&q.mu)
	heap.Init(&q.priorityQueue)
	heap.Init(&q.delayQueue)
	return q
}

func (q *TaskQueue) Submit(task *Task) (uint64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isStopped {
		return 0, ErrQueueStopped
	}

	task.ID = atomic.AddUint64(&q.nextTaskID, 1)

	if task.MaxRetries <= 0 {
		task.MaxRetries = 3
	}

	if task.Priority < 0 {
		task.Priority = 0
	} else if task.Priority > 9 {
		task.Priority = 9
	}

	task.scheduledAt = time.Now().Add(task.Delay)
	task.retryCount = 0

	if task.Delay > 0 {
		heap.Push(&q.delayQueue, task)
	} else {
		heap.Push(&q.priorityQueue, task)
	}

	atomic.AddUint64(&q.stats.submitted, 1)

	q.cond.Signal()

	return task.ID, nil
}

func (q *TaskQueue) Start(workerCount int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isRunning {
		return
	}

	if workerCount <= 0 {
		workerCount = 1
	}

	q.workerCount = workerCount
	q.isRunning = true
	q.workers = make([]*worker, workerCount)

	for i := 0; i < workerCount; i++ {
		w := newWorker(q, i)
		q.workers[i] = w
		go w.start()
	}

	go q.delayDispatcher()
}

func (q *TaskQueue) delayDispatcher() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopChan:
			return
		case <-ticker.C:
			q.mu.Lock()
			now := time.Now()
			for q.delayQueue.Len() > 0 {
				task := q.delayQueue[0]
				if now.After(task.scheduledAt) {
					heap.Pop(&q.delayQueue)
					heap.Push(&q.priorityQueue, task)
					q.cond.Signal()
				} else {
					break
				}
			}
			q.mu.Unlock()
		}
	}
}

func (q *TaskQueue) Stop() int {
	q.mu.Lock()
	if q.isStopped {
		q.mu.Unlock()
		return q.pendingTaskCountLocked()
	}
	q.isStopped = true
	q.mu.Unlock()

	close(q.stopChan)

	q.mu.Lock()
	q.cond.Broadcast()
	q.mu.Unlock()

	timeout := time.After(10 * time.Second)
	select {
	case <-q.stopWaitChan:
	case <-timeout:
	}

	return q.PendingCount()
}

func (q *TaskQueue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pendingTaskCountLocked()
}

func (q *TaskQueue) pendingTaskCountLocked() int {
	return q.priorityQueue.Len() + q.delayQueue.Len()
}

func (q *TaskQueue) DeadLetters() []*Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]*Task, len(q.deadLetters))
	copy(result, q.deadLetters)
	return result
}

func (q *TaskQueue) Stats() (submitted, completed, retries, deadLetters uint64) {
	submitted = atomic.LoadUint64(&q.stats.submitted)
	completed = atomic.LoadUint64(&q.stats.completed)
	retries = atomic.LoadUint64(&q.stats.retries)
	deadLetters = atomic.LoadUint64(&q.stats.deadLetters)
	return
}

func (q *TaskQueue) getTask() (*Task, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for {
		if q.isStopped {
			return nil, false
		}

		if q.priorityQueue.Len() > 0 {
			task := heap.Pop(&q.priorityQueue).(*Task)
			return task, true
		}

		q.cond.Wait()
	}
}

func (q *TaskQueue) taskCompleted(task *Task) {
	atomic.AddUint64(&q.stats.completed, 1)
}

func (q *TaskQueue) taskFailed(task *Task, err error) {
	task.lastErr = err
	task.retryCount++

	atomic.AddUint64(&q.stats.retries, 1)

	if task.retryCount >= task.MaxRetries {
		q.mu.Lock()
		q.deadLetters = append(q.deadLetters, task)
		q.mu.Unlock()
		atomic.AddUint64(&q.stats.deadLetters, 1)
		return
	}

	interval := q.getRetryInterval(task.retryCount - 1)
	q.mu.Lock()
	task.scheduledAt = time.Now().Add(interval)
	heap.Push(&q.delayQueue, task)
	q.mu.Unlock()
}

func (q *TaskQueue) getRetryInterval(retryIndex int) time.Duration {
	if retryIndex >= len(retryIntervals) {
		return retryIntervals[len(retryIntervals)-1]
	}
	return retryIntervals[retryIndex]
}

func (q *TaskQueue) workerStopped() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.workerCount--
	if q.workerCount == 0 {
		close(q.stopWaitChan)
	}
}

type priorityQueue []*Task

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*Task)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type delayQueue []*Task

func (dq delayQueue) Len() int { return len(dq) }

func (dq delayQueue) Less(i, j int) bool {
	return dq[i].scheduledAt.Before(dq[j].scheduledAt)
}

func (dq delayQueue) Swap(i, j int) {
	dq[i], dq[j] = dq[j], dq[i]
}

func (dq *delayQueue) Push(x interface{}) {
	item := x.(*Task)
	*dq = append(*dq, item)
}

func (dq *delayQueue) Pop() interface{} {
	old := *dq
	n := len(old)
	item := old[n-1]
	*dq = old[0 : n-1]
	return item
}
