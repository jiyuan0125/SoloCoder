package limiter

import (
	"math"
	"net/http"
	"sync"
	"time"
)

type TokenBucketLimiter struct {
	baseLimiter
	rate       int
	burst      int
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucketLimiter(rate, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucketLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	if elapsed > 0 {
		newTokens := float64(elapsed.Seconds()) * float64(tb.rate)
		tb.tokens = math.Min(tb.tokens+newTokens, float64(tb.burst))
		tb.lastRefill = now
	}
}

func (tb *TokenBucketLimiter) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.incTotal()
	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens--
		tb.incAllowed()
		return true
	}

	tb.incBlocked()
	return false
}

func (tb *TokenBucketLimiter) WithKey(key string) Limiter {
	return tb
}

func (tb *TokenBucketLimiter) SetGlobalMax(max int64) {
}

func (tb *TokenBucketLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
