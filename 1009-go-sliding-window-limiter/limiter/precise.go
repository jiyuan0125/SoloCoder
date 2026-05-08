package limiter

import (
	"time"
)

type preciseLimiter struct {
	baseLimiter
	timestamps []time.Time
}

func newPreciseLimiter(rule LimitRule) *preciseLimiter {
	return &preciseLimiter{
		baseLimiter: baseLimiter{rule: rule},
		timestamps:  make([]time.Time, 0),
	}
}

func (p *preciseLimiter) cleanupExpired(now time.Time) {
	cutoff := now.Add(-p.rule.Window)
	idx := 0
	for i, ts := range p.timestamps {
		if ts.After(cutoff) {
			idx = i
			break
		}
		idx = i + 1
	}
	if idx > 0 {
		p.timestamps = p.timestamps[idx:]
	}
}

func (p *preciseLimiter) calculateRetryAfter(now time.Time) time.Duration {
	if len(p.timestamps) == 0 {
		return 0
	}

	oldest := p.timestamps[0]
	retryTime := oldest.Add(p.rule.Window)
	waitTime := retryTime.Sub(now)
	if waitTime < 0 {
		waitTime = 0
	}
	return waitTime
}

func (p *preciseLimiter) Allow() Result {
	return p.AllowN(1)
}

func (p *preciseLimiter) AllowN(n int) Result {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	p.cleanupExpired(now)

	currentCount := len(p.timestamps)

	if currentCount+n > p.rule.Limit {
		retryAfter := p.calculateRetryAfter(now)
		return Result{
			Allowed:    false,
			Remaining:  p.rule.Limit - currentCount,
			Limit:      p.rule.Limit,
			RetryAfter: retryAfter,
			WindowEnd:  now.Add(p.rule.Window),
		}
	}

	for i := 0; i < n; i++ {
		p.timestamps = append(p.timestamps, now)
	}

	newCount := len(p.timestamps)
	return Result{
		Allowed:    true,
		Remaining:  p.rule.Limit - newCount,
		Limit:      p.rule.Limit,
		RetryAfter: 0,
		WindowEnd:  now.Add(p.rule.Window),
	}
}

func (p *preciseLimiter) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timestamps = make([]time.Time, 0)
}

func (p *preciseLimiter) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now()
	p.cleanupExpired(now)

	currentCount := len(p.timestamps)
	var lastAccess time.Time
	if currentCount > 0 {
		lastAccess = p.timestamps[currentCount-1]
	}

	return map[string]interface{}{
		"mode":       "precise",
		"limit":      p.rule.Limit,
		"window":     p.rule.Window.String(),
		"current":    currentCount,
		"remaining":  p.rule.Limit - currentCount,
		"last_access": lastAccess,
	}
}
