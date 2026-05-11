package idtoken

import (
	"sync"
	"time"
)

type NonceCache struct {
	mu    sync.RWMutex
	store map[string]time.Time
	ttl   time.Duration
}

func NewNonceCache() *NonceCache {
	return &NonceCache{
		store: make(map[string]time.Time),
		ttl:   NonceTTL,
	}
}

func NewNonceCacheWithTTL(ttl time.Duration) *NonceCache {
	return &NonceCache{
		store: make(map[string]time.Time),
		ttl:   ttl,
	}
}

func (c *NonceCache) Add(nonce string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[nonce] = time.Now().Add(c.ttl)
}

func (c *NonceCache) Validate(nonce string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	expiry, ok := c.store[nonce]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}

func (c *NonceCache) Remove(nonce string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, nonce)
}

func (c *NonceCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for nonce, expiry := range c.store {
		if now.After(expiry) {
			delete(c.store, nonce)
		}
	}
}

func (c *NonceCache) StartCleanup(interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-stop:
				return
			}
		}
	}()
	return stop
}
