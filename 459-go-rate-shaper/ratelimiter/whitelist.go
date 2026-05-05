package ratelimiter

import (
	"sync"
	"time"
)

// WhitelistLimiter wraps a Limiter with whitelist support.
// Keys in the whitelist bypass rate limit checks and are always allowed.
type WhitelistLimiter struct {
	mu        sync.RWMutex
	limiter   Limiter
	whitelist map[string]struct{}
}

// NewWhitelistLimiter creates a new WhitelistLimiter that wraps the given limiter.
func NewWhitelistLimiter(limiter Limiter) *WhitelistLimiter {
	return &WhitelistLimiter{
		limiter:   limiter,
		whitelist: make(map[string]struct{}),
	}
}

// AddToWhitelist adds a key to the whitelist.
func (w *WhitelistLimiter) AddToWhitelist(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.whitelist[key] = struct{}{}
}

// RemoveFromWhitelist removes a key from the whitelist.
func (w *WhitelistLimiter) RemoveFromWhitelist(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.whitelist, key)
}

// IsWhitelisted checks if a key is in the whitelist.
func (w *WhitelistLimiter) IsWhitelisted(key string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, exists := w.whitelist[key]
	return exists
}

// GetWhitelist returns a copy of the current whitelist.
func (w *WhitelistLimiter) GetWhitelist() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	keys := make([]string, 0, len(w.whitelist))
	for key := range w.whitelist {
		keys = append(keys, key)
	}
	return keys
}

// Allow checks if a request is allowed.
// If the key is in the whitelist, it is always allowed.
func (w *WhitelistLimiter) Allow(key string) Result {
	if w.IsWhitelisted(key) {
		return Result{
			Allowed:   true,
			Limited:   false,
			Remaining: -1,
			WaitTime:  0,
			ResetAt:   time.Time{},
		}
	}
	return w.limiter.Allow(key)
}

// AllowN checks if n requests are allowed.
// If the key is in the whitelist, it is always allowed.
func (w *WhitelistLimiter) AllowN(key string, n int64) Result {
	if w.IsWhitelisted(key) {
		return Result{
			Allowed:   true,
			Limited:   false,
			Remaining: -1,
			WaitTime:  0,
			ResetAt:   time.Time{},
		}
	}
	return w.limiter.AllowN(key, n)
}

// Stats returns statistics for the given key.
func (w *WhitelistLimiter) Stats(key string) Stats {
	return w.limiter.Stats(key)
}

// UpdateConfig updates the configuration of the underlying limiter.
func (w *WhitelistLimiter) UpdateConfig(config Config) error {
	return w.limiter.UpdateConfig(config)
}

// GetConfig returns the current configuration.
func (w *WhitelistLimiter) GetConfig() Config {
	return w.limiter.GetConfig()
}

// Reset resets the rate limiter for the given key.
func (w *WhitelistLimiter) Reset(key string) {
	w.limiter.Reset(key)
}

// ResetAll resets all rate limiters.
func (w *WhitelistLimiter) ResetAll() {
	w.limiter.ResetAll()
}
