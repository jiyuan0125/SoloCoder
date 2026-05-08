package cache

import (
	"container/list"
	"sync"
	"time"
)

type memoryItem struct {
	key      string
	value    interface{}
	expireAt time.Time
}

type MemoryCache struct {
	capacity int
	items    map[string]*list.Element
	lru      *list.List
	mu       sync.RWMutex
}

func NewMemoryCache(capacity int) *MemoryCache {
	if capacity <= 0 {
		capacity = 1000
	}
	return &MemoryCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lru:      list.New(),
	}
}

func (m *MemoryCache) Get(key string) (interface{}, time.Time, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	elem, ok := m.items[key]
	if !ok {
		return nil, time.Time{}, false
	}

	item := elem.Value.(*memoryItem)
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		return nil, time.Time{}, false
	}

	m.lru.MoveToFront(elem)
	return item.value, item.expireAt, true
}

func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if elem, ok := m.items[key]; ok {
		m.lru.MoveToFront(elem)
		item := elem.Value.(*memoryItem)
		item.value = value
		if ttl > 0 {
			item.expireAt = time.Now().Add(ttl)
		} else {
			item.expireAt = time.Time{}
		}
		return
	}

	if m.lru.Len() >= m.capacity {
		m.evict()
	}

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	item := &memoryItem{
		key:      key,
		value:    value,
		expireAt: expireAt,
	}
	elem := m.lru.PushFront(item)
	m.items[key] = elem
}

func (m *MemoryCache) Delete(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if elem, ok := m.items[key]; ok {
		m.lru.Remove(elem)
		delete(m.items, key)
		return true
	}
	return false
}

func (m *MemoryCache) DeletePrefix(prefix string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for key, elem := range m.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			m.lru.Remove(elem)
			delete(m.items, key)
			count++
		}
	}
	return count
}

func (m *MemoryCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*list.Element)
	m.lru = list.New()
}

func (m *MemoryCache) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lru.Len()
}

func (m *MemoryCache) Capacity() int {
	return m.capacity
}

func (m *MemoryCache) Usage() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return float64(m.lru.Len()) / float64(m.capacity)
}

func (m *MemoryCache) evict() {
	if elem := m.lru.Back(); elem != nil {
		item := elem.Value.(*memoryItem)
		m.lru.Remove(elem)
		delete(m.items, item.key)
	}
}
