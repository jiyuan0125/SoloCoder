package pool

import (
	"errors"
	"sync"
)

type Manager struct {
	mu    sync.RWMutex
	pools map[string]*GenericPool
}

func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]*GenericPool),
	}
}

func (m *Manager) CreatePool(cfg PoolConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[cfg.Name]; exists {
		return errors.New("pool with this name already exists")
	}

	pool, err := NewGenericPool(cfg)
	if err != nil {
		return err
	}

	m.pools[cfg.Name] = pool
	return nil
}

func (m *Manager) GetPool(name string) (*GenericPool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, exists := m.pools[name]
	return pool, exists
}

func (m *Manager) GetAllStats() []PoolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]PoolStats, 0, len(m.pools))
	for _, pool := range m.pools {
		stats = append(stats, pool.Stats())
	}
	return stats
}

func (m *Manager) GetStats(name string) (PoolStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, exists := m.pools[name]
	if !exists {
		return PoolStats{}, false
	}
	return pool.Stats(), true
}

func (m *Manager) ListPools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.pools))
	for name := range m.pools {
		names = append(names, name)
	}
	return names
}

func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, pool := range m.pools {
		_ = pool.Close()
		delete(m.pools, name)
	}
}
