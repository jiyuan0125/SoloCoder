package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

type Pool[T any] struct {
	config      Config[T]
	mu          sync.Mutex
	free        []entry[T]
	closed      atomic.Bool
	closeOnce   sync.Once
	created     atomic.Int64
	discarded   atomic.Int64
	hits        atomic.Int64
	misses      atomic.Int64
	lastCleanup atomic.Int64
}

type entry[T any] struct {
	value      T
	putTime    time.Time
}

type Config[T any] struct {
	MaxSize        int
	Factory        func() T
	Reset          func(T)
	IdleTimeout    time.Duration
	CleanupOnGet   bool
}

func DefaultConfig[T any](factory func() T) Config[T] {
	return Config[T]{
		MaxSize:     100,
		Factory:     factory,
		Reset:       nil,
		IdleTimeout: 5 * time.Minute,
		CleanupOnGet: true,
	}
}

func New[T any](config Config[T]) *Pool[T] {
	if config.Factory == nil {
		panic("pool: factory function is required")
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	p := &Pool[T]{
		config: config,
		free:   make([]entry[T], 0),
	}
	p.lastCleanup.Store(time.Now().UnixNano())
	return p
}

func (p *Pool[T]) Get() T {
	if p.closed.Load() {
		return p.config.Factory()
	}

	if p.config.CleanupOnGet {
		p.tryCleanupOnGet()
	}

	p.mu.Lock()
	if len(p.free) > 0 {
		e := p.free[len(p.free)-1]
		p.free = p.free[:len(p.free)-1]
		p.mu.Unlock()
		p.hits.Add(1)
		return e.value
	}
	p.mu.Unlock()

	p.misses.Add(1)
	p.created.Add(1)
	return p.config.Factory()
}

func (p *Pool[T]) Put(value T) {
	if p.closed.Load() {
		return
	}

	if p.config.Reset != nil {
		p.config.Reset(value)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.free) >= p.config.MaxSize {
		p.discarded.Add(1)
		return
	}

	p.free = append(p.free, entry[T]{
		value:   value,
		putTime: time.Now(),
	})
}

func (p *Pool[T]) tryCleanupOnGet() {
	now := time.Now()
	last := p.lastCleanup.Load()
	lastTime := time.Unix(0, last)

	if now.Sub(lastTime) < 10*time.Second {
		return
	}

	if !p.lastCleanup.CompareAndSwap(last, now.UnixNano()) {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	cleaned := 0
	maxClean := 10
	deadline := now.Add(-p.config.IdleTimeout)

	for len(p.free) > 0 && cleaned < maxClean {
		e := p.free[0]
		if e.putTime.After(deadline) {
			break
		}
		copy(p.free, p.free[1:])
		p.free = p.free[:len(p.free)-1]
		cleaned++
	}

	p.discarded.Add(int64(cleaned))
}

func (p *Pool[T]) Stats() Stats {
	p.mu.Lock()
	free := len(p.free)
	p.mu.Unlock()

	hits := p.hits.Load()
	misses := p.misses.Load()
	total := hits + misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return Stats{
		Free:      free,
		Created:   p.created.Load(),
		Discarded: p.discarded.Load(),
		Hits:      hits,
		Misses:    misses,
		HitRate:   hitRate,
	}
}

func (p *Pool[T]) Close() {
	p.closeOnce.Do(func() {
		p.closed.Store(true)

		p.mu.Lock()
		n := len(p.free)
		p.free = nil
		p.mu.Unlock()

		p.discarded.Add(int64(n))
	})
}
