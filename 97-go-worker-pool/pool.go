package workerpool

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultInitialWorkers = 4
	DefaultMaxWorkers     = 32
	DefaultIdleTimeout    = 10 * time.Second
	DefaultThresholdFactor = 10
)

type Config struct {
	InitialWorkers int
	MaxWorkers     int
	IdleTimeout    time.Duration
	Threshold      int
}

func DefaultConfig() *Config {
	return &Config{
		InitialWorkers: DefaultInitialWorkers,
		MaxWorkers:     DefaultMaxWorkers,
		IdleTimeout:    DefaultIdleTimeout,
		Threshold:      DefaultInitialWorkers * DefaultThresholdFactor,
	}
}

type Pool struct {
	config       *Config
	taskChan     chan *task
	workers      map[int]*worker
	workersMu    sync.RWMutex
	nextWorkerID int32
	closed       int32
	closing      int32

	completedCount int64
	droppedCount   int64

	stopOnce sync.Once
}

func NewPool(cfg *Config) *Pool {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.InitialWorkers <= 0 {
		cfg.InitialWorkers = DefaultInitialWorkers
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = DefaultMaxWorkers
	}
	if cfg.MaxWorkers < cfg.InitialWorkers {
		cfg.MaxWorkers = cfg.InitialWorkers
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = DefaultIdleTimeout
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = cfg.InitialWorkers * DefaultThresholdFactor
	}

	p := &Pool{
		config:       cfg,
		taskChan:     make(chan *task, cfg.Threshold+cfg.MaxWorkers),
		workers:      make(map[int]*worker),
		nextWorkerID: 1,
	}

	for i := 0; i < cfg.InitialWorkers; i++ {
		p.createWorker(false)
	}

	return p
}

func (p *Pool) createWorker(isExtra bool) *worker {
	id := int(atomic.AddInt32(&p.nextWorkerID, 1)) - 1
	w := newWorker(id, p.taskChan, p, isExtra)
	p.workersMu.Lock()
	p.workers[id] = w
	p.workersMu.Unlock()
	w.start()
	return w
}

func (p *Pool) workerCount() int {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()
	return len(p.workers)
}

func (p *Pool) scaleUpIfNeeded() {
	pending := len(p.taskChan)
	if pending < p.config.Threshold {
		return
	}

	currentWorkers := p.workerCount()
	if currentWorkers >= p.config.MaxWorkers {
		return
	}

	needed := min(p.config.MaxWorkers-currentWorkers, (pending-p.config.Threshold)/p.config.Threshold+1)
	for i := 0; i < needed; i++ {
		p.createWorker(true)
	}
}

func (p *Pool) workerExit(w *worker) {
	p.workersMu.Lock()
	delete(p.workers, w.id)
	p.workersMu.Unlock()
}

func (p *Pool) replaceWorker(w *worker) {
	p.workersMu.Lock()
	delete(p.workers, w.id)
	p.workersMu.Unlock()
	p.createWorker(w.isExtra)
}

func (p *Pool) Submit(task func() error) (Future, error) {
	return p.SubmitWithTimeout(task, 0)
}

func (p *Pool) SubmitWithTimeout(task func() error, timeout time.Duration) (Future, error) {
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, ErrPoolClosed
	}

	t := newTask(task, timeout)

	p.scaleUpIfNeeded()

	select {
	case p.taskChan <- t:
		return t.future, nil
	default:
		return nil, ErrPoolClosed
	}
}

func (p *Pool) Stop() (int64, int64) {
	return p.StopTimeout(0)
}

func (p *Pool) StopTimeout(timeout time.Duration) (int64, int64) {
	if !atomic.CompareAndSwapInt32(&p.closing, 0, 1) {
		return atomic.LoadInt64(&p.completedCount), atomic.LoadInt64(&p.droppedCount)
	}

	atomic.StoreInt32(&p.closed, 1)

	var remainingTasks []*task
	drainLoop:
	for {
		select {
		case t := <-p.taskChan:
			remainingTasks = append(remainingTasks, t)
		default:
			break drainLoop
		}
	}

	waitChan := make(chan struct{})
	go func() {
		p.waitAllWorkers()
		close(waitChan)
	}()

	if timeout > 0 {
		select {
		case <-waitChan:
		case <-time.After(timeout):
			p.stopAllWorkers()
			for _, t := range remainingTasks {
				t.future.complete(ErrDropped)
				atomic.AddInt64(&p.droppedCount, 1)
			}
		}
	} else {
		<-waitChan
		for _, t := range remainingTasks {
			t.future.complete(ErrDropped)
			atomic.AddInt64(&p.droppedCount, 1)
		}
	}

	return atomic.LoadInt64(&p.completedCount), atomic.LoadInt64(&p.droppedCount)
}

func (p *Pool) waitAllWorkers() {
	for {
		p.workersMu.RLock()
		workers := make([]*worker, 0, len(p.workers))
		for _, w := range p.workers {
			workers = append(workers, w)
		}
		p.workersMu.RUnlock()

		if len(workers) == 0 {
			break
		}

		for _, w := range workers {
			for w.isRunning() {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (p *Pool) stopAllWorkers() {
	p.workersMu.RLock()
	workers := make([]*worker, 0, len(p.workers))
	for _, w := range p.workers {
		workers = append(workers, w)
	}
	p.workersMu.RUnlock()

	for _, w := range workers {
		w.stop()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
