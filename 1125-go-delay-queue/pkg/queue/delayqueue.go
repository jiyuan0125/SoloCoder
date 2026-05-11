package queue

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrTaskAlreadyExec = errors.New("task already executed")
)

type DelayQueue struct {
	mu    sync.Mutex
	heap  taskHeap
	index map[string]*Task
}

func New() *DelayQueue {
	q := &DelayQueue{
		heap:  taskHeap{},
		index: make(map[string]*Task),
	}
	heap.Init(&q.heap)
	return q
}

func (q *DelayQueue) Push(task *Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.index[task.ID]; exists {
		return errors.New("task already exists")
	}

	task.index = -1
	heap.Push(&q.heap, task)
	q.index[task.ID] = task
	return nil
}

func (q *DelayQueue) Peek() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.heap.Len() == 0 {
		return nil
	}

	task := q.heap[0]
	copy := *task
	copy.index = -1
	return &copy
}

func (q *DelayQueue) Poll(now time.Time) *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.heap.Len() == 0 {
		return nil
	}

	task := q.heap[0]
	if !task.IsExpired(now) {
		return nil
	}

	heap.Pop(&q.heap)
	delete(q.index, task.ID)
	return task
}

func (q *DelayQueue) Drain(now time.Time) []*Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	var tasks []*Task
	for q.heap.Len() > 0 {
		task := q.heap[0]
		if !task.IsExpired(now) {
			break
		}
		heap.Pop(&q.heap)
		delete(q.index, task.ID)
		tasks = append(tasks, task)
	}
	return tasks
}

func (q *DelayQueue) ModifyDelay(id string, newExecuteAt time.Time) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.index[id]
	if !exists {
		return ErrTaskNotFound
	}

	if task.index < 0 {
		return ErrTaskAlreadyExec
	}

	task.ExecuteAt = newExecuteAt
	heap.Fix(&q.heap, task.index)
	return nil
}

func (q *DelayQueue) Promote(id string, newPriority Priority) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.index[id]
	if !exists {
		return ErrTaskNotFound
	}

	if task.index < 0 {
		return ErrTaskAlreadyExec
	}

	if newPriority <= task.Priority {
		return nil
	}

	task.Priority = newPriority
	heap.Fix(&q.heap, task.index)
	return nil
}

func (q *DelayQueue) Get(id string) *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.index[id]
	if !exists {
		return nil
	}

	copy := *task
	copy.index = -1
	return &copy
}

func (q *DelayQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.heap.Len()
}
