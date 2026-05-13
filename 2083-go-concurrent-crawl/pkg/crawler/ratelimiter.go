package crawler

import (
	"sync"
	"time"
)

type RateLimiter struct {
	limits   map[string]*domainLimiter
	limitsMu sync.RWMutex
	rate     int
}

type domainLimiter struct {
	lastRequest time.Time
	mu          sync.Mutex
}

func NewRateLimiter(rate int) *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*domainLimiter),
		rate:   rate,
	}
}

func (rl *RateLimiter) Wait(domain string) {
	limiter := rl.getLimiter(domain)
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	interval := time.Second / time.Duration(rl.rate)
	elapsed := time.Since(limiter.lastRequest)
	
	if elapsed < interval {
		time.Sleep(interval - elapsed)
	}
	
	limiter.lastRequest = time.Now()
}

func (rl *RateLimiter) getLimiter(domain string) *domainLimiter {
	rl.limitsMu.RLock()
	if limiter, exists := rl.limits[domain]; exists {
		rl.limitsMu.RUnlock()
		return limiter
	}
	rl.limitsMu.RUnlock()

	rl.limitsMu.Lock()
	defer rl.limitsMu.Unlock()

	if limiter, exists := rl.limits[domain]; exists {
		return limiter
	}

	limiter := &domainLimiter{}
	rl.limits[domain] = limiter
	return limiter
}
