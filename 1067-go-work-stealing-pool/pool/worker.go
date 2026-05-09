package pool

import (
	"sync"
	"time"
)

type Worker struct {
	id        int
	queue     *deque
	pool      *Pool
	running   bool
	stopping  bool
	stopped   chan struct{}
	mu        sync.Mutex
}

func newWorker(id int, pool *Pool) *Worker {
	return &Worker{
		id:      id,
		queue:   newDeque(),
		pool:    pool,
		running: true,
		stopped: make(chan struct{}, 1),
	}
}

func (w *Worker) start() {
	go w.run()
}

func (w *Worker) stopGracefully() {
	w.mu.Lock()
	w.stopping = true
	w.mu.Unlock()
	w.queue.signal()
}

func (w *Worker) run() {
	for {
		w.mu.Lock()
		if !w.running {
			w.mu.Unlock()
			break
		}
		w.mu.Unlock()

		task, ok := w.queue.popFront()
		if ok {
			w.executeTask(task)
			continue
		}

		w.mu.Lock()
		if w.stopping {
			w.running = false
			w.mu.Unlock()
			break
		}
		w.mu.Unlock()

		stolen := w.trySteal()
		if stolen != nil {
			w.executeTask(stolen)
			continue
		}

		select {
		case <-w.pool.closeCh:
			w.mu.Lock()
			w.running = false
			w.mu.Unlock()
			w.queue.broadcast()
			return
		case <-time.After(10 * time.Millisecond):
		}
	}

	close(w.stopped)
}

func (w *Worker) trySteal() *Task {
	workers := w.pool.getWorkers()
	for i := 0; i < len(workers); i++ {
		other := workers[(w.id+1+i)%len(workers)]
		if other.id == w.id {
			continue
		}
		if task, ok := other.queue.popBack(); ok {
			return task
		}
	}
	return nil
}

func (w *Worker) executeTask(task *Task) {
	task.run(w.pool.activeTasks)
}

func (w *Worker) queueLen() int {
	return w.queue.len()
}

func (w *Worker) submit(task *Task) {
	w.queue.pushBack(task)
}

func (w *Worker) drain() []*Task {
	return w.queue.drain()
}
