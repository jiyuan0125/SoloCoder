package limiter

import (
	"sync"
	"time"

	"multirate-limiter/common"
)

const (
	cleanupIntervalRequests = 100
	cleanupIntervalSeconds  = 1
)

type ruleLimiter struct {
	rule       common.Rule
	windowSec  int64
	buffer     *ringBuffer
	mu         sync.Mutex

	requestCount    int
	lastCleanupTime int64
}

func newRuleLimiter(rule common.Rule) *ruleLimiter {
	windowSec := rule.Granularity.Duration()
	capacity := rule.Quota * 2
	if capacity < 64 {
		capacity = 64
	}
	return &ruleLimiter{
		rule:            rule,
		windowSec:       windowSec,
		buffer:          newRingBuffer(capacity),
		requestCount:    0,
		lastCleanupTime: 0,
	}
}

func (r *ruleLimiter) shouldCleanup(nowSec int64) bool {
	r.requestCount++
	if r.requestCount >= cleanupIntervalRequests {
		r.requestCount = 0
		return true
	}
	if nowSec-r.lastCleanupTime >= cleanupIntervalSeconds {
		return true
	}
	return false
}

func (r *ruleLimiter) cleanup(nowSec int64) {
	threshold := nowSec - r.windowSec
	r.buffer.clearBefore(threshold)
	r.lastCleanupTime = nowSec
}

func (r *ruleLimiter) allow(forceCleanup bool) (allowed bool, current int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nowSec := time.Now().Unix()

	if forceCleanup || r.shouldCleanup(nowSec) {
		r.cleanup(nowSec)
		r.requestCount = 0
	}

	current = r.buffer.size()
	if current >= r.rule.Quota {
		r.cleanup(nowSec)
		current = r.buffer.size()
		if current >= r.rule.Quota {
			return false, current
		}
	}

	r.buffer.add(nowSec)
	return true, current
}

func (r *ruleLimiter) stats() (current int, quota int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nowSec := time.Now().Unix()
	r.cleanup(nowSec)
	current = r.buffer.size()
	quota = r.rule.Quota
	return current, quota
}

func (r *ruleLimiter) getRule() common.Rule {
	return r.rule
}
