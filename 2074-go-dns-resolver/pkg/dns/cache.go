package dns

import (
	"sync"
	"time"
)

type cacheItem struct {
	result      *DNSQueryResult
	expiresAt   time.Time
}

type Cache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

func NewCache() *Cache {
	c := &Cache{
		items: make(map[string]cacheItem),
	}
	go c.cleanupExpired()
	return c
}

func (c *Cache) Get(key string) (*DNSQueryResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	result := *item.result
	return &result, true
}

func (c *Cache) Set(key string, result *DNSQueryResult) {
	if len(result.Records) == 0 {
		return
	}

	minTTL := result.Records[0].TTL
	for _, r := range result.Records {
		if r.TTL < minTTL {
			minTTL = r.TTL
		}
	}

	if minTTL == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		result:    result,
		expiresAt: time.Now().Add(time.Duration(minTTL) * time.Second),
	}
}

func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheItem)
}
