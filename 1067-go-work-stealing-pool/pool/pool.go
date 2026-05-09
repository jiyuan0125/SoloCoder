package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	workers      []*Worker
	workersMu    sync.RWMutex
	nextWorkerID int32
	closed       int32
	closeCh      chan struct{}
	activeTasks  *syncWaitGroup
}

type PoolOption func(*Pool)

func New(size int, options ...PoolOption) *Pool {
	if size <= 0 {
		size = 4
	}

	p := &Pool{
		workers:     make([]*Worker, 0, size),
		closeCh:     make(chan struct{}),
		activeTasks: &syncWaitGroup{},
	}

	for _, opt := range options {
		opt(p)
	}

	for i := 0; i < size; i++ {
		p.addWorker()
	}

	return p
}

func (p *Pool) addWorker() {
	id := atomic.AddInt32(&p.nextWorkerID, 1) - 1
	w := newWorker(int(id), p)
	p.workersMu.Lock()
	p.workers = append(p.workers, w)
	p.workersMu.Unlock()
	w.start()
}

func (p *Pool) getWorkers() []*Worker {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()
	workers := make([]*Worker, len(p.workers))
	copy(workers, p.workers)
	return workers
}

func (p *Pool) findLeastLoadedWorker() *Worker {
	workers := p.getWorkers()
	if len(workers) == 0 {
		return nil
	}

	minLen := -1
	var best *Worker
	for _, w := range workers {
		l := w.queueLen()
		if minLen == -1 || l < minLen {
			minLen = l
			best = w
		}
	}
	return best
}

func (p *Pool) Submit(fn func()) (*Future, error) {
	return p.SubmitWithTimeout(fn, 0)
}

func (p *Pool) SubmitWithTimeout(fn func(), timeout time.Duration) (*Future, error) {
	if atomic.LoadInt32(&p.closed) != 0 {
		return nil, ErrPoolClosed
	}

	future := newFuture()
	task := &Task{
		fn:      fn,
		future:  future,
		timeout: timeout,
	}

	w := p.findLeastLoadedWorker()
	if w == nil {
		return nil, ErrPoolClosed
	}

	w.submit(task)
	return future, nil
}

func (p *Pool) AddWorkers(count int) {
	for i := 0; i < count; i++ {
		p.addWorker()
	}
}

func (p *Pool) RemoveWorkers(count int) {
	if count <= 0 {
		return
	}

	for i := 0; i < count; i++ {
		var w *Worker
		p.workersMu.Lock()
		if len(p.workers) > 1 {
			w = p.workers[len(p.workers)-1]
			p.workers = p.workers[:len(p.workers)-1]
		}
		p.workersMu.Unlock()

		if w != nil {
			w.stopGracefully()
		} else {
			break
		}
	}
}

func (p *Pool) Size() int {
	return len(p.getWorkers())
}

func (p *Pool) Shutdown() {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return
	}

	close(p.closeCh)

	workers := p.getWorkers()
	for _, w := range workers {
		w.stopGracefully()
	}

	p.activeTasks.Wait()
}

func (p *Pool) ShutdownNow() []*Task {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return nil
	}

	close(p.closeCh)

	var pendingTasks []*Task
	workers := p.getWorkers()
	for _, w := range workers {
		tasks := w.drain()
		pendingTasks = append(pendingTasks, tasks...)
	}

	for _, t := range pendingTasks {
		t.future.complete(nil, ErrPoolClosed)
	}

	p.activeTasks.Wait()
	return pendingTasks
}
