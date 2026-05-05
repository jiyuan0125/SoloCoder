package ratelimiter

import (
	"sync"
	"time"
)

// FixedWindowLimiter implements rate limiting using fixed time windows.
// It counts requests in fixed intervals and resets the counter at window boundaries.
// Note: This mode may allow doubled traffic at window boundaries.
type FixedWindowLimiter struct {
	mu     sync.RWMutex
	config Config
	store  Storage
}

type fixedWindowState struct {
	count     int64
	startTime time.Time
}

// NewFixedWindowLimiter creates a new FixedWindowLimiter with the given configuration.
func NewFixedWindowLimiter(config Config) (*FixedWindowLimiter, error) {
	config.Mode = ModeFixedWindow
	if err := config.Validate(); err != nil {
		return nil, err
	}

	store := NewMemoryStorage()

	return &FixedWindowLimiter{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a request is allowed.
func (l *FixedWindowLimiter) Allow(key string) Result {
	return l.AllowN(key, 1)
}

// AllowN checks if n requests are allowed.
func (l *FixedWindowLimiter) AllowN(key string, n int64) Result {
	l.mu.RLock()
	limit := l.config.Limit
	window := l.config.Window
	l.mu.RUnlock()

	now := time.Now()
	windowStart := now.Truncate(window)

	l.mu.Lock()
	defer l.mu.Unlock()

	stateAny, exists := l.store.Get(key)
	var state fixedWindowState
	if exists {
		state = stateAny.(fixedWindowState)
	}

	if !exists || state.startTime.Before(windowStart) {
		state = fixedWindowState{
			count:     0,
			startTime: windowStart,
		}
	}

	allowed := state.count+n <= limit
	if allowed {
		state.count += n
		l.store.Set(key, state)
	}

	remaining := limit - state.count
	if remaining < 0 {
		remaining = 0
	}

	resetAt := windowStart.Add(window)
	waitTime := time.Duration(0)
	if !allowed {
		waitTime = resetAt.Sub(now)
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
func (l *FixedWindowLimiter) Stats(key string) Stats {
	l.mu.RLock()
	limit := l.config.Limit
	window := l.config.Window
	l.mu.RUnlock()

	now := time.Now()
	windowStart := now.Truncate(window)

	l.mu.RLock()
	defer l.mu.RUnlock()

	stateAny, exists := l.store.Get(key)
	var count int64 = 0
	if exists {
		state := stateAny.(fixedWindowState)
		if !state.startTime.Before(windowStart) {
			count = state.count
		}
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return Stats{
		Key:       key,
		Mode:      ModeFixedWindow,
		Used:      count,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   windowStart.Add(window),
	}
}

// UpdateConfig updates the configuration.
func (l *FixedWindowLimiter) UpdateConfig(config Config) error {
	if config.Mode != ModeFixedWindow {
		return ErrModeMismatch
	}
	if err := config.Validate(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.config.Limit = config.Limit
	l.config.Window = config.Window

	return nil
}

// GetConfig returns the current configuration.
func (l *FixedWindowLimiter) GetConfig() Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// Reset resets the rate limiter for the given key.
func (l *FixedWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Delete(key)
}

// ResetAll resets all rate limiters.
func (l *FixedWindowLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Clear()
}
