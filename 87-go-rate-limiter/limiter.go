package limiter

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	ModeSlidingWindow = "sliding_window"
	ModeTokenBucket   = "token_bucket"
)

type Limiter interface {
	http.Handler
	Allow() bool
	WithKey(key string) Limiter
	Stats() (total, allowed, blocked int64)
	SetGlobalMax(max int64)
}

type RateLimiter struct {
	mode         string
	rate         int
	burst        int
	globalMax    int64
	globalActive int64
	keyedLimiters map[string]Limiter
	defaultLimiter Limiter
	mu           sync.RWMutex
	globalMu     sync.Mutex
	total        int64
	allowed      int64
	blocked      int64
	statsMu      sync.RWMutex
	next         http.Handler
}

func New(mode string, rate, burst int) Limiter {
	rl := &RateLimiter{
		mode:          mode,
		rate:          rate,
		burst:         burst,
		globalMax:     0,
		globalActive:  0,
		keyedLimiters: make(map[string]Limiter),
	}
	rl.defaultLimiter = rl.createLimiter()
	return rl
}

func (rl *RateLimiter) createLimiter() Limiter {
	switch rl.mode {
	case ModeSlidingWindow:
		return NewSlidingWindowLimiter(rl.rate)
	case ModeTokenBucket:
		return NewTokenBucketLimiter(rl.rate, rl.burst)
	default:
		return NewSlidingWindowLimiter(rl.rate)
	}
}

func (rl *RateLimiter) WithKey(key string) Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.keyedLimiters[key]; exists {
		return limiter
	}

	keyed := &KeyedLimiter{
		parent: rl,
		key:    key,
		inner:  rl.createLimiter(),
	}
	rl.keyedLimiters[key] = keyed
	return keyed
}

func (rl *RateLimiter) SetGlobalMax(max int64) {
	rl.globalMu.Lock()
	rl.globalMax = max
	rl.globalMu.Unlock()
}

func (rl *RateLimiter) checkGlobalLimit() bool {
	if rl.globalMax <= 0 {
		return true
	}
	rl.globalMu.Lock()
	defer rl.globalMu.Unlock()
	return rl.globalActive < rl.globalMax
}

func (rl *RateLimiter) incGlobalActive() {
	rl.globalMu.Lock()
	rl.globalActive++
	rl.globalMu.Unlock()
}

func (rl *RateLimiter) decGlobalActive() {
	rl.globalMu.Lock()
	rl.globalActive--
	rl.globalMu.Unlock()
}

func (rl *RateLimiter) Allow() bool {
	rl.statsMu.Lock()
	rl.total++
	rl.statsMu.Unlock()

	if !rl.checkGlobalLimit() {
		rl.statsMu.Lock()
		rl.blocked++
		rl.statsMu.Unlock()
		return false
	}

	allowed := rl.defaultLimiter.Allow()
	if allowed {
		rl.statsMu.Lock()
		rl.allowed++
		rl.statsMu.Unlock()
		rl.incGlobalActive()
	} else {
		rl.statsMu.Lock()
		rl.blocked++
		rl.statsMu.Unlock()
	}
	return allowed
}

func (rl *RateLimiter) Stats() (total, allowed, blocked int64) {
	rl.statsMu.RLock()
	defer rl.statsMu.RUnlock()
	return rl.total, rl.allowed, rl.blocked
}

func (rl *RateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := extractIP(r)
	keyedLimiter := rl.WithKey(key)
	
	if !keyedLimiter.Allow() {
		writeTooManyRequests(w)
		return
	}
	
	defer rl.decGlobalActive()
	
	if rl.next != nil {
		rl.next.ServeHTTP(w, r)
	}
}

type KeyedLimiter struct {
	parent *RateLimiter
	key    string
	inner  Limiter
}

func (kl *KeyedLimiter) Allow() bool {
	if !kl.parent.checkGlobalLimit() {
		kl.parent.statsMu.Lock()
		kl.parent.blocked++
		kl.parent.statsMu.Unlock()
		return false
	}
	
	allowed := kl.inner.Allow()
	if allowed {
		kl.parent.statsMu.Lock()
		kl.parent.allowed++
		kl.parent.statsMu.Unlock()
		kl.parent.incGlobalActive()
	} else {
		kl.parent.statsMu.Lock()
		kl.parent.blocked++
		kl.parent.statsMu.Unlock()
	}
	return allowed
}

func (kl *KeyedLimiter) WithKey(key string) Limiter {
	return kl.parent.WithKey(key)
}

func (kl *KeyedLimiter) Stats() (total, allowed, blocked int64) {
	return kl.parent.Stats()
}

func (kl *KeyedLimiter) SetGlobalMax(max int64) {
	kl.parent.SetGlobalMax(max)
}

func (kl *KeyedLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kl.parent.ServeHTTP(w, r)
}

func extractIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip
			}
		}
	}
	
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeTooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "1")
	w.WriteHeader(http.StatusTooManyRequests)
	
	response := map[string]interface{}{
		"error":       "too many requests",
		"retry_after": 1,
	}
	json.NewEncoder(w).Encode(response)
}

type baseLimiter struct {
	mu      sync.Mutex
	total   int64
	allowed int64
	blocked int64
}

func (bl *baseLimiter) Stats() (total, allowed, blocked int64) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	return bl.total, bl.allowed, bl.blocked
}

func (bl *baseLimiter) incTotal() {
	bl.mu.Lock()
	bl.total++
	bl.mu.Unlock()
}

func (bl *baseLimiter) incAllowed() {
	bl.mu.Lock()
	bl.allowed++
	bl.mu.Unlock()
}

func (bl *baseLimiter) incBlocked() {
	bl.mu.Lock()
	bl.blocked++
	bl.mu.Unlock()
}
