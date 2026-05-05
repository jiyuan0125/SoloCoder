package cron

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Expr       string
	ParsedExpr *CronExpr
	LastAccess time.Time
}

type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
	}
}

func (c *Cache) Get(expr string) (*CronExpr, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[expr]
	if !exists {
		return nil, false
	}

	entry.LastAccess = time.Now()
	return entry.ParsedExpr, true
}

func (c *Cache) Set(expr string, parsedExpr *CronExpr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[expr] = &CacheEntry{
		Expr:       expr,
		ParsedExpr: parsedExpr,
		LastAccess: time.Now(),
	}
}

func (c *Cache) Cleanup(olderThan time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.LastAccess) > olderThan {
			delete(c.entries, key)
		}
	}
}
