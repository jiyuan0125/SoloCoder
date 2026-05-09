package pool

import (
	"sync"
)

type deque struct {
	items []*Task
	mu    sync.Mutex
	cond  *sync.Cond
}

func newDeque() *deque {
	d := &deque{
		items: make([]*Task, 0),
	}
	d.cond = sync.NewCond(&d.mu)
	return d
}

func (d *deque) pushFront(t *Task) {
	d.mu.Lock()
	d.items = append([]*Task{t}, d.items...)
	d.cond.Signal()
	d.mu.Unlock()
}

func (d *deque) pushBack(t *Task) {
	d.mu.Lock()
	d.items = append(d.items, t)
	d.cond.Signal()
	d.mu.Unlock()
}

func (d *deque) popFront() (*Task, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.items) == 0 {
		return nil, false
	}
	t := d.items[0]
	d.items = d.items[1:]
	return t, true
}

func (d *deque) popBack() (*Task, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.items) == 0 {
		return nil, false
	}
	t := d.items[len(d.items)-1]
	d.items = d.items[:len(d.items)-1]
	return t, true
}

func (d *deque) len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.items)
}

func (d *deque) wait() {
	d.mu.Lock()
	d.cond.Wait()
	d.mu.Unlock()
}

func (d *deque) signal() {
	d.mu.Lock()
	d.cond.Signal()
	d.mu.Unlock()
}

func (d *deque) broadcast() {
	d.mu.Lock()
	d.cond.Broadcast()
	d.mu.Unlock()
}

func (d *deque) drain() []*Task {
	d.mu.Lock()
	defer d.mu.Unlock()
	items := d.items
	d.items = make([]*Task, 0)
	return items
}
