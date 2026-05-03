package limiter

import (
	"net/http"
	"sync"
	"time"
)

type SlidingWindowLimiter struct {
	rate       int
	windowSize time.Duration
	requests   []time.Time
	mu         sync.Mutex
}

func NewSlidingWindowLimiter(rate int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		rate:       rate,
		windowSize: time.Second,
		requests:   make([]time.Time, 0, rate),
	}
}

func (sw *SlidingWindowLimiter) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.windowSize)

	validIdx := 0
	for i, t := range sw.requests {
		if t.After(cutoff) {
			validIdx = i
			break
		}
	}

	if validIdx > 0 {
		sw.requests = sw.requests[validIdx:]
	} else if len(sw.requests) > 0 && !sw.requests[0].After(cutoff) {
		sw.requests = sw.requests[:0]
	}

	if len(sw.requests) >= sw.rate {
		return false
	}

	sw.requests = append(sw.requests, now)
	return true
}

func (sw *SlidingWindowLimiter) WithKey(key string) Limiter {
	return sw
}

func (sw *SlidingWindowLimiter) Stats() (total, allowed, blocked int64) {
	return 0, 0, 0
}

func (sw *SlidingWindowLimiter) SetGlobalMax(max int64) {
}

func (sw *SlidingWindowLimiter) Use(next http.Handler) Limiter {
	return sw
}

func (sw *SlidingWindowLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
