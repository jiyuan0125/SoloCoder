package cache

import (
	"container/list"
	"sync"
	"time"
)

type Status int

const (
	StatusActive Status = iota
	StatusExpired
	StatusEvicted
)

type Item struct {
	Key       string
	Value     string
	ExpiresAt time.Time
	Status    Status
}

type Stats struct {
	HitCount    uint64
	TotalCount  uint64
	CurrentSize int
}

type Cache struct {
	mu         sync.RWMutex
	lruList    *list.List
	itemMap    map[string]*list.Element
	capacity   int
	defaultTTL time.Duration
	stats      Stats
}

func New(capacity int, defaultTTLSeconds int) *Cache {
	return &Cache{
		lruList:    list.New(),
		itemMap:    make(map[string]*list.Element),
		capacity:   capacity,
		defaultTTL: time.Duration(defaultTTLSeconds) * time.Second,
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.TotalCount++

	elem, ok := c.itemMap[key]
	if !ok {
		return "", false
	}

	item := elem.Value.(*Item)

	if item.Status == StatusEvicted {
		delete(c.itemMap, key)
		c.lruList.Remove(elem)
		return "", false
	}

	if item.Status == StatusExpired || time.Now().After(item.ExpiresAt) {
		item.Status = StatusExpired
		delete(c.itemMap, key)
		c.lruList.Remove(elem)
		return "", false
	}

	c.stats.HitCount++
	c.lruList.MoveToFront(elem)
	return item.Value, true
}

func (c *Cache) Set(key, value string, ttlSeconds *int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ttl time.Duration
	if ttlSeconds != nil && *ttlSeconds > 0 {
		ttl = time.Duration(*ttlSeconds) * time.Second
	} else {
		ttl = c.defaultTTL
	}

	if elem, ok := c.itemMap[key]; ok {
		item := elem.Value.(*Item)
		item.Value = value
		item.ExpiresAt = time.Now().Add(ttl)
		item.Status = StatusActive
		c.lruList.MoveToFront(elem)
		return
	}

	if c.lruList.Len() >= c.capacity {
		c.evictLocked()
	}

	item := &Item{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		Status:    StatusActive,
	}
	elem := c.lruList.PushFront(item)
	c.itemMap[key] = elem
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.itemMap[key]; ok {
		item := elem.Value.(*Item)
		item.Status = StatusEvicted
		delete(c.itemMap, key)
		c.lruList.Remove(elem)
	}
}

func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lruList = list.New()
	c.itemMap = make(map[string]*list.Element)
}

func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		HitCount:    c.stats.HitCount,
		TotalCount:  c.stats.TotalCount,
		CurrentSize: c.lruList.Len(),
	}
}

func (c *Cache) SetConfig(capacity int, defaultTTLSeconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.capacity = capacity
	c.defaultTTL = time.Duration(defaultTTLSeconds) * time.Second

	for c.lruList.Len() > c.capacity {
		c.evictLocked()
	}
}

func (c *Cache) Capacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capacity
}

func (c *Cache) DefaultTTL() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int(c.defaultTTL / time.Second)
}

func (c *Cache) evictLocked() {
	for {
		back := c.lruList.Back()
		if back == nil {
			return
		}
		item := back.Value.(*Item)
		if item.Status != StatusActive || time.Now().After(item.ExpiresAt) {
			item.Status = StatusExpired
		} else {
			item.Status = StatusEvicted
		}
		delete(c.itemMap, item.Key)
		c.lruList.Remove(back)

		if c.lruList.Len() <= c.capacity {
			return
		}
	}
}
