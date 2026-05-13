package cache

import (
	"container/list"
	"sync"
	"time"
)

type item struct {
	key       string
	value     string
	expiresAt time.Time
}

type Cache struct {
	mu         sync.RWMutex
	ll         *list.List
	items      map[string]*list.Element
	capacity   int
	defaultTTL time.Duration
	hits       int64
	misses     int64
}

func New(capacity int, defaultTTL int) *Cache {
	return &Cache{
		ll:         list.New(),
		items:      make(map[string]*list.Element),
		capacity:   capacity,
		defaultTTL: time.Duration(defaultTTL) * time.Second,
	}
}

func (c *Cache) Set(key, value string, ttlSeconds *int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ttl time.Duration
	if ttlSeconds != nil {
		ttl = time.Duration(*ttlSeconds) * time.Second
	} else {
		ttl = c.defaultTTL
	}

	if elem, ok := c.items[key]; ok {
		elem.Value.(*item).value = value
		elem.Value.(*item).expiresAt = time.Now().Add(ttl)
		c.ll.MoveToFront(elem)
		return
	}

	elem := c.ll.PushFront(&item{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})
	c.items[key] = elem

	if c.ll.Len() > c.capacity {
		c.evictOldest()
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		c.misses++
		return "", false
	}

	it := elem.Value.(*item)
	if time.Now().After(it.expiresAt) {
		c.removeElement(elem)
		c.misses++
		return "", false
	}

	c.ll.MoveToFront(elem)
	c.hits++
	return it.value, true
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return false
	}

	c.removeElement(elem)
	return true
}

func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll = list.New()
	c.items = make(map[string]*list.Element)
}

type Stats struct {
	Count       int   `json:"count"`
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Capacity    int   `json:"capacity"`
	DefaultTTL  int   `json:"default_ttl"`
}

func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Count:      c.ll.Len(),
		Hits:       c.hits,
		Misses:     c.misses,
		Capacity:   c.capacity,
		DefaultTTL: int(c.defaultTTL / time.Second),
	}
}

func (c *Cache) SetCapacity(newCapacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldCapacity := c.capacity
	c.capacity = newCapacity

	if newCapacity < oldCapacity {
		for c.ll.Len() > c.capacity {
			c.evictOldest()
		}
	}
}

func (c *Cache) SetDefaultTTL(ttl int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTTL = time.Duration(ttl) * time.Second
}

func (c *Cache) evictOldest() {
	elem := c.ll.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *Cache) removeElement(elem *list.Element) {
	c.ll.Remove(elem)
	delete(c.items, elem.Value.(*item).key)
}
