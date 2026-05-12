package manager

import (
	"conn-pool/internal/config"
	"conn-pool/internal/pool"
	"sync"
)

type PoolInfo struct {
	Config config.PoolConfig `json:"config"`
	Stats  pool.PoolStats    `json:"stats"`
}

type Manager struct {
	pools map[string]*pool.Pool
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]*pool.Pool),
	}
}

func (m *Manager) AddPool(cfg config.PoolConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.pools[cfg.Address]; exists {
		return
	}
	
	p := pool.NewPool(cfg)
	m.pools[cfg.Address] = p
}

func (m *Manager) GetPool(address string) (*pool.Pool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	p, exists := m.pools[address]
	return p, exists
}

func (m *Manager) ListAddresses() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	addresses := make([]string, 0, len(m.pools))
	for addr := range m.pools {
		addresses = append(addresses, addr)
	}
	return addresses
}

func (m *Manager) GetAllInfo() map[string]PoolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]PoolInfo)
	for addr, p := range m.pools {
		result[addr] = PoolInfo{
			Config: p.GetConfig(),
			Stats:  p.GetStats(),
		}
	}
	return result
}

func (m *Manager) SetMaxOpen(address string, newMax int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	p, exists := m.pools[address]
	if !exists {
		return false
	}
	
	p.SetMaxOpen(newMax)
	return true
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, p := range m.pools {
		p.Close()
	}
}
