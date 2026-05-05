package ratelimiter

import (
	"sync"
	"time"
)

// SmoothWindowLimiter implements rate limiting using a smooth window algorithm.
// It divides the window into multiple small buckets to smooth out boundary bursts.
// This eliminates the "traffic doubling" problem at fixed window boundaries.
type SmoothWindowLimiter struct {
	mu        sync.RWMutex
	config    Config
	store     Storage
	bucketDur time.Duration
}

type smoothWindowState struct {
	buckets    []int64
	lastBucket int64
}

// NewSmoothWindowLimiter creates a new SmoothWindowLimiter with the given configuration.
func NewSmoothWindowLimiter(config Config) (*SmoothWindowLimiter, error) {
	config.Mode = ModeSmoothWindow
	if config.BucketCount < 2 {
		config.BucketCount = 10
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	bucketDur := config.Window / time.Duration(config.BucketCount)
	store := NewMemoryStorage()

	return &SmoothWindowLimiter{
		config:    config,
		store:     store,
		bucketDur: bucketDur,
	}, nil
}

// Allow checks if a request is allowed.
func (l *SmoothWindowLimiter) Allow(key string) Result {
	return l.AllowN(key, 1)
}

// AllowN checks if n requests are allowed.
func (l *SmoothWindowLimiter) AllowN(key string, n int64) Result {
	l.mu.RLock()
	limit := l.config.Limit
	window := l.config.Window
	bucketCount := l.config.BucketCount
	bucketDur := l.bucketDur
	l.mu.RUnlock()

	now := time.Now()
	currentBucket := now.UnixNano() / bucketDur.Nanoseconds()

	l.mu.Lock()
	defer l.mu.Unlock()

	stateAny, exists := l.store.Get(key)
	var state smoothWindowState
	if exists {
		state = stateAny.(smoothWindowState)
	}

	if state.buckets == nil {
		state.buckets = make([]int64, bucketCount)
		state.lastBucket = currentBucket
	}

	bucketsPassed := currentBucket - state.lastBucket
	if bucketsPassed > 0 {
		if bucketsPassed >= int64(bucketCount) {
			for i := range state.buckets {
				state.buckets[i] = 0
			}
		} else {
			for i := int64(0); i < bucketsPassed; i++ {
				idx := (state.lastBucket + 1 + i) % int64(bucketCount)
				state.buckets[idx] = 0
			}
		}
		state.lastBucket = currentBucket
	}

	var total int64 = 0
	for _, count := range state.buckets {
		total += count
	}

	allowed := total+n <= limit

	currentBucketIdx := currentBucket % int64(bucketCount)
	if allowed {
		state.buckets[currentBucketIdx] += n
		total += n
	}

	l.store.Set(key, state)

	remaining := limit - total
	if remaining < 0 {
		remaining = 0
	}

	waitTime := time.Duration(0)
	resetAt := now.Add(window)

	if !allowed {
		oldestBucketStart := (currentBucket - int64(bucketCount-1)) * bucketDur.Nanoseconds()
		waitTime = time.Duration(oldestBucketStart + bucketDur.Nanoseconds() - now.UnixNano())
		if waitTime < 0 {
			waitTime = 0
		}
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
func (l *SmoothWindowLimiter) Stats(key string) Stats {
	l.mu.RLock()
	limit := l.config.Limit
	window := l.config.Window
	bucketCount := l.config.BucketCount
	bucketDur := l.bucketDur
	l.mu.RUnlock()

	now := time.Now()
	currentBucket := now.UnixNano() / bucketDur.Nanoseconds()

	l.mu.RLock()
	defer l.mu.RUnlock()

	stateAny, exists := l.store.Get(key)
	var total int64 = 0

	if exists {
		state := stateAny.(smoothWindowState)
		if state.buckets != nil {
			bucketsPassed := currentBucket - state.lastBucket
			if bucketsPassed >= int64(bucketCount) {
				total = 0
			} else {
				for i := int64(0); i < bucketsPassed; i++ {
					idx := (state.lastBucket + 1 + i) % int64(bucketCount)
					state.buckets[idx] = 0
				}
				for _, count := range state.buckets {
					total += count
				}
			}
		}
	}

	remaining := limit - total
	if remaining < 0 {
		remaining = 0
	}

	return Stats{
		Key:       key,
		Mode:      ModeSmoothWindow,
		Used:      total,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   now.Add(window),
	}
}

// UpdateConfig updates the configuration.
func (l *SmoothWindowLimiter) UpdateConfig(config Config) error {
	if config.Mode != ModeSmoothWindow {
		return ErrModeMismatch
	}
	if config.BucketCount < 2 {
		config.BucketCount = 10
	}
	if err := config.Validate(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.config.Limit = config.Limit
	l.config.Window = config.Window
	if config.BucketCount != l.config.BucketCount {
		l.config.BucketCount = config.BucketCount
		l.bucketDur = config.Window / time.Duration(config.BucketCount)
	}

	return nil
}

// GetConfig returns the current configuration.
func (l *SmoothWindowLimiter) GetConfig() Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// Reset resets the rate limiter for the given key.
func (l *SmoothWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Delete(key)
}

// ResetAll resets all rate limiters.
func (l *SmoothWindowLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Clear()
}
