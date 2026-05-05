package ratelimiter

import "sync"

// MemoryStorage is an in-memory implementation of the Storage interface.
// It uses a map to store key-value pairs and is safe for concurrent use.
type MemoryStorage struct {
	mu sync.RWMutex
	m  map[string]interface{}
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		m: make(map[string]interface{}),
	}
}

// Get retrieves the value associated with the given key.
func (s *MemoryStorage) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.m[key]
	return value, exists
}

// Set sets the value associated with the given key.
func (s *MemoryStorage) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

// Delete deletes the value associated with the given key.
func (s *MemoryStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}

// Clear removes all keys and values.
func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[string]interface{})
}
