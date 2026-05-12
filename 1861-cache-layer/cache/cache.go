package cache

import (
	"cache-layer/policy"
	"container/list"
	"sync"
	"time"
)

type Entry struct {
	Key   string
	Value string
}

type Stats struct {
	HitRate   float64
	ItemCount int
	TotalHits int
}

type Cache interface {
	Get(key string) (string, bool)
	Set(key, value string, ttl time.Duration) bool
	Delete(key string) bool
	Flush()
	Capacity() int
	SetCapacity(capacity int)
	Stats() Stats
}

type entryInternal struct {
	key   string
	value string
	node  *list.Element
}

type lruCache struct {
	mu          sync.RWMutex
	capacity    int
	defaultTTL  time.Duration
	lruList     *list.List
	items       map[string]*entryInternal
	ttlCache    policy.TTLCache
	totalHits   int64
	totalMisses int64
}

func NewLRUCache(capacity int, defaultTTL time.Duration) Cache {
	return &lruCache{
		capacity:   capacity,
		defaultTTL: defaultTTL,
		lruList:    list.New(),
		items:      make(map[string]*entryInternal),
		ttlCache:   policy.NewTTLCache(),
	}
}

func (c *lruCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalHits++

	entry, ok := c.items[key]
	if !ok {
		c.totalMisses++
		return "", false
	}

	if c.ttlCache.IsExpired(key) {
		c.removeEntry(entry)
		c.totalMisses++
		return "", false
	}

	c.lruList.MoveToFront(entry.node)
	return entry.value, true
}

func (c *lruCache) Set(key, value string, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.items[key]; ok {
		existing.value = value
		c.lruList.MoveToFront(existing.node)
		c.setTTL(key, ttl)
		return true
	}

	if len(c.items) >= c.capacity && c.capacity > 0 {
		c.evictLRU()
	}

	node := c.lruList.PushFront(key)
	entry := &entryInternal{
		key:   key,
		value: value,
		node:  node,
	}
	c.items[key] = entry
	c.setTTL(key, ttl)
	return true
}

func (c *lruCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok {
		return false
	}
	c.removeEntry(entry)
	return true
}

func (c *lruCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lruList.Init()
	c.items = make(map[string]*entryInternal)
	c.ttlCache = policy.NewTTLCache()
}

func (c *lruCache) Capacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capacity
}

func (c *lruCache) SetCapacity(capacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if capacity < 0 {
		capacity = 0
	}
	c.capacity = capacity

	for len(c.items) > c.capacity && c.capacity > 0 {
		c.evictLRU()
	}
}

func (c *lruCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.totalHits + c.totalMisses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.totalHits) / float64(total)
	}

	return Stats{
		HitRate:   hitRate,
		ItemCount: len(c.items),
		TotalHits: int(total),
	}
}

func (c *lruCache) setTTL(key string, ttl time.Duration) {
	if ttl < 0 {
		ttl = 0
	}
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	c.ttlCache.Put(key, ttl)
}

func (c *lruCache) removeEntry(entry *entryInternal) {
	c.lruList.Remove(entry.node)
	delete(c.items, entry.key)
	c.ttlCache.Remove(entry.key)
}

func (c *lruCache) evictLRU() {
	back := c.lruList.Back()
	if back == nil {
		return
	}
	key := back.Value.(string)
	if entry, ok := c.items[key]; ok {
		c.removeEntry(entry)
	}
}
