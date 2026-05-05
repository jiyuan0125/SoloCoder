package ratelimiter

import (
	"sync"
	"time"
)

// LeakyBucketLimiter implements rate limiting using a leaky bucket algorithm.
// Requests are queued and processed at a constant rate, regardless of incoming
// request burstiness. This smooths out traffic spikes.
type LeakyBucketLimiter struct {
	mu     sync.RWMutex
	config Config
	store  Storage
}

type leakyBucketState struct {
	queueSize  int64
	lastLeak   time.Time
}

// NewLeakyBucketLimiter creates a new LeakyBucketLimiter with the given configuration.
func NewLeakyBucketLimiter(config Config) (*LeakyBucketLimiter, error) {
	config.Mode = ModeLeakyBucket
	if err := config.Validate(); err != nil {
		return nil, err
	}

	store := NewMemoryStorage()

	return &LeakyBucketLimiter{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a request is allowed.
func (l *LeakyBucketLimiter) Allow(key string) Result {
	return l.AllowN(key, 1)
}

// AllowN checks if n requests are allowed.
func (l *LeakyBucketLimiter) AllowN(key string, n int64) Result {
	l.mu.RLock()
	rate := l.config.Rate
	capacity := l.config.Capacity
	l.mu.RUnlock()

	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	stateAny, exists := l.store.Get(key)
	var state leakyBucketState

	if exists {
		state = stateAny.(leakyBucketState)
		elapsed := now.Sub(state.lastLeak).Seconds()
		leaked := int64(elapsed * rate)
		if leaked > state.queueSize {
			state.queueSize = 0
		} else {
			state.queueSize -= leaked
		}
	} else {
		state = leakyBucketState{
			queueSize: 0,
			lastLeak:  now,
		}
	}

	state.lastLeak = now
	allowed := state.queueSize+n <= capacity

	if allowed {
		state.queueSize += n
	}

	l.store.Set(key, state)

	remaining := capacity - state.queueSize
	if remaining < 0 {
		remaining = 0
	}

	waitTime := time.Duration(0)
	resetAt := now

	if !allowed {
		needed := state.queueSize + n - capacity
		waitSeconds := float64(needed) / rate
		waitTime = time.Duration(waitSeconds * float64(time.Second))
		resetAt = now.Add(waitTime)
	}

	return Result{
		Allowed:   allowed,
		Limited:   !allowed,
		Remaining: remaining,
		WaitTime:  waitTime,
		ResetAt:   resetAt,
	}
}

// Stats returns statistics for the given key.
func (l *LeakyBucketLimiter) Stats(key string) Stats {
	l.mu.RLock()
	rate := l.config.Rate
	capacity := l.config.Capacity
	l.mu.RUnlock()

	now := time.Now()

	l.mu.RLock()
	defer l.mu.RUnlock()

	stateAny, exists := l.store.Get(key)
	var queueSize int64 = 0

	if exists {
		state := stateAny.(leakyBucketState)
		elapsed := now.Sub(state.lastLeak).Seconds()
		leaked := int64(elapsed * rate)
		if leaked > state.queueSize {
			queueSize = 0
		} else {
			queueSize = state.queueSize - leaked
		}
	}

	remaining := capacity - queueSize
	if remaining < 0 {
		remaining = 0
	}

	return Stats{
		Key:       key,
		Mode:      ModeLeakyBucket,
		Used:      queueSize,
		Limit:     capacity,
		Remaining: remaining,
		ResetAt:   now.Add(time.Duration(float64(queueSize)/rate) * time.Second),
	}
}

// UpdateConfig updates the configuration.
func (l *LeakyBucketLimiter) UpdateConfig(config Config) error {
	if config.Mode != ModeLeakyBucket {
		return ErrModeMismatch
	}
	if err := config.Validate(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.config.Rate = config.Rate
	l.config.Capacity = config.Capacity

	return nil
}

// GetConfig returns the current configuration.
func (l *LeakyBucketLimiter) GetConfig() Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// Reset resets the rate limiter for the given key.
func (l *LeakyBucketLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Delete(key)
}

// ResetAll resets all rate limiters.
func (l *LeakyBucketLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Clear()
}
