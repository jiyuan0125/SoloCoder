package ratelimiter

import (
	"sync"
	"time"
)

// SlidingWindowLimiter implements rate limiting using a sliding window algorithm.
// It maintains timestamps of requests and counts those within the window duration.
// This provides more accurate rate limiting than fixed window but uses more memory.
type SlidingWindowLimiter struct {
	mu     sync.RWMutex
	config Config
	store  Storage
}

type slidingWindowState struct {
	timestamps []time.Time
}

// NewSlidingWindowLimiter creates a new SlidingWindowLimiter with the given configuration.
func NewSlidingWindowLimiter(config Config) (*SlidingWindowLimiter, error) {
	config.Mode = ModeSlidingWindow
	if err := config.Validate(); err != nil {
		return nil, err
	}

	store := NewMemoryStorage()

	return &SlidingWindowLimiter{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a request is allowed.
func (l *SlidingWindowLimiter) Allow(key string) Result {
	return l.AllowN(key, 1)
}

// AllowN checks if n requests are allowed.
func (l *SlidingWindowLimiter) AllowN(key string, n int64) Result {
	l.mu.RLock()
	limit := l.config.Limit
	window := l.config.Window
	l.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-window)

	l.mu.Lock()
	defer l.mu.Unlock()

	stateAny, exists := l.store.Get(key)
	var state slidingWindowState
	if exists {
		state = stateAny.(slidingWindowState)
	}

	validTimestamps := make([]time.Time, 0, len(state.timestamps))
	for _, ts := range state.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	currentCount := int64(len(validTimestamps))
	allowed := currentCount+n <= limit

	if allowed {
		for i := int64(0); i < n; i++ {
			validTimestamps = append(validTimestamps, now)
		}
		state.timestamps = validTimestamps
		l.store.Set(key, state)
	} else {
		state.timestamps = validTimestamps
		l.store.Set(key, state)
	}

	remaining := limit - int64(len(validTimestamps))
	if remaining < 0 {
		remaining = 0
	}

	waitTime := time.Duration(0)
	resetAt := now.Add(window)

	if !allowed && len(validTimestamps) > 0 {
		oldest := validTimestamps[0]
		waitTime = oldest.Add(window).Sub(now)
		if waitTime < 0 {
			waitTime = 0
		}
		resetAt = oldest.Add(window)
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
func (l *SlidingWindowLimiter) Stats(key string) Stats {
	l.mu.RLock()
	limit := l.config.Limit
	window := l.config.Window
	l.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-window)

	l.mu.RLock()
	defer l.mu.RUnlock()

	stateAny, exists := l.store.Get(key)
	var count int64 = 0
	var oldest time.Time

	if exists {
		state := stateAny.(slidingWindowState)
		for _, ts := range state.timestamps {
			if ts.After(cutoff) {
				count++
				if oldest.IsZero() || ts.Before(oldest) {
					oldest = ts
				}
			}
		}
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(window)
	if !oldest.IsZero() {
		resetAt = oldest.Add(window)
	}

	return Stats{
		Key:       key,
		Mode:      ModeSlidingWindow,
		Used:      count,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}
}

// UpdateConfig updates the configuration.
func (l *SlidingWindowLimiter) UpdateConfig(config Config) error {
	if config.Mode != ModeSlidingWindow {
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
func (l *SlidingWindowLimiter) GetConfig() Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// Reset resets the rate limiter for the given key.
func (l *SlidingWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Delete(key)
}

// ResetAll resets all rate limiters.
func (l *SlidingWindowLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Clear()
}
