package limiter

import (
	"math"
	"net/http"
	"sync"
	"time"
)

type TokenBucketLimiter struct {
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

	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucketLimiter) WithKey(key string) Limiter {
	return tb
}

func (tb *TokenBucketLimiter) Stats() (total, allowed, blocked int64) {
	return 0, 0, 0
}

func (tb *TokenBucketLimiter) SetGlobalMax(max int64) {
}

func (tb *TokenBucketLimiter) Use(next http.Handler) Limiter {
	return tb
}

func (tb *TokenBucketLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
