package queue

import (
	"sync"
)

type LockedQueue struct {
	mu    sync.Mutex
	items []interface{}
}

func NewLocked() *LockedQueue {
	return &LockedQueue{
		items: make([]interface{}, 0),
	}
}

func (q *LockedQueue) Enqueue(v interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, v)
}

func (q *LockedQueue) Dequeue() (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil, false
	}
	v := q.items[0]
	q.items = q.items[1:]
	return v, true
}

func (q *LockedQueue) Depth() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int64(len(q.items))
}
