package pool

import (
	"container/heap"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed = errors.New("pool is closed")
	ErrZeroWorkers = errors.New("cannot submit to pool with zero workers")
)

type Pool struct {
	workerCount         int
	idleWorkerCount   int32
	queue               priorityQueue
	mu                  sync.Mutex
	cond                *sync.Cond
	closed              bool
	starvationThreshold time.Duration
	taskCounter         uint64
	starvedTasksCounter int32
	wg                  sync.WaitGroup
}

func New(workerCount int, starvationThreshold time.Duration) (*Pool, error) {
	if workerCount < 0 {
		return nil, errors.New("worker count cannot be negative")
	}

	p := &Pool{
		workerCount:         workerCount,
		starvationThreshold: starvationThreshold,
	}
	p.cond = sync.NewCond(&p.mu)
	p.queue = make(priorityQueue, 0)
	heap.Init(&p.queue)

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p, nil
}

func (p *Pool) worker() {
	defer p.wg.Done()
	atomic.AddInt32(&p.idleWorkerCount, 1)

	for {
		task := p.getNextTask()
		if task == nil {
			return
		}
		atomic.AddInt32(&p.idleWorkerCount, -1)
		task.fn()
		atomic.AddInt32(&p.idleWorkerCount, 1)
	}
}

func (p *Pool) Submit(priority Priority, fn func()) error {
	if p.workerCount == 0 {
		return ErrZeroWorkers
	}

	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}

	task := &Task{
		id:         atomic.AddUint64(&p.taskCounter, 1),
		priority:   priority,
		submitTime: time.Now(),
		fn:         fn,
	}

	heap.Push(&p.queue, task)
	p.mu.Unlock()

	p.cond.Signal()
	return nil
}

func (p *Pool) getNextTask() *Task {
	p.mu.Lock()
	defer p.mu.Unlock()

	for !p.closed && p.queue.Len() == 0 {
		p.cond.Wait()
	}

	if p.closed && p.queue.Len() == 0 {
		return nil
	}

	promoted := p.queue.checkAndPromote(p.starvationThreshold)
	if promoted > 0 {
		atomic.AddInt32(&p.starvedTasksCounter, int32(promoted))
	}

	return heap.Pop(&p.queue).(*Task)
}

func (p *Pool) Status() Status {
	p.mu.Lock()
	high, medium, low := p.queue.countByPriority()
	p.mu.Unlock()

	return Status{
		WorkerCount:        p.workerCount,
		IdleWorkerCount:  int(atomic.LoadInt32(&p.idleWorkerCount)),
		QueueHighPriority: high,
		QueueMediumPriority: medium,
		QueueLowPriority:    low,
		StarvedTasksCount:    int(atomic.LoadInt32(&p.starvedTasksCounter)),
	}
}

func (p *Pool) Shutdown() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	p.cond.Broadcast()

	p.wg.Wait()
}

type Status struct {
	WorkerCount        int
	IdleWorkerCount  int
	QueueHighPriority int
	QueueMediumPriority int
	QueueLowPriority    int
	StarvedTasksCount    int
}
