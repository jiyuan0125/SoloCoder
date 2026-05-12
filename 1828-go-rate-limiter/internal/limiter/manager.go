package limiter

import (
	"sync"
	"time"
)

type Manager struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
	stopCh  chan struct{}
}

func NewManager() *Manager {
	m := &Manager{
		buckets: make(map[string]*Bucket),
		stopCh:  make(chan struct{}),
	}
	go m.refiller()
	return m
}

func (m *Manager) refiller() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			for _, b := range m.buckets {
				b.Notify()
			}
			m.mu.RUnlock()
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) SetRule(cfg *BucketConfig) error {
	if cfg.Capacity <= 0 {
		return &ConfigError{Field: "capacity", Reason: "must be greater than 0"}
	}
	if cfg.Rate <= 0 {
		return &ConfigError{Field: "rate", Reason: "must be greater than 0"}
	}
	if cfg.WaitTimeout < 0 {
		return &ConfigError{Field: "waitTimeout", Reason: "cannot be negative"}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.buckets[cfg.Path]; ok {
		existing.Reset(cfg)
	} else {
		m.buckets[cfg.Path] = NewBucket(cfg)
	}
	return nil
}

func (m *Manager) DeleteRule(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[path]
	if !ok {
		return false
	}

	b.MarkDeleted()
	b.Notify()
	delete(m.buckets, path)
	return true
}

func (m *Manager) GetBucket(path string) *Bucket {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.buckets[path]
}

func (m *Manager) ListStats() []*BucketStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]*BucketStats, 0, len(m.buckets))
	for path, b := range m.buckets {
		stats = append(stats, b.Stats(path))
	}
	return stats
}

func (m *Manager) Close() {
	close(m.stopCh)
}

type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Reason
}
