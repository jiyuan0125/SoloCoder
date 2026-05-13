package pool

import (
	"errors"
	"log"
	"net"
	"sync"
)

type Manager struct {
	pools sync.Map
}

func NewManager() *Manager {
	return &Manager{}
}

type RegisterRequest struct {
	Name   string
	Config PoolConfig
}

func (m *Manager) Register(req RegisterRequest) error {
	if req.Name == "" {
		return errors.New("pool name is required")
	}

	if _, exists := m.pools.Load(req.Name); exists {
		return errors.New("pool with this name already exists")
	}

	dialFunc := func(network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	}

	pool := NewPool(req.Name, req.Config, dialFunc)

	if err := pool.WarmUp(); err != nil {
		log.Printf("[Manager] WarmUp failed for pool %s: %v", req.Name, err)
	}

	pool.StartHealthCheck()
	m.pools.Store(req.Name, pool)
	log.Printf("[Manager] Pool %s registered successfully", req.Name)
	return nil
}

func (m *Manager) Unregister(name string) error {
	value, exists := m.pools.Load(name)
	if !exists {
		return errors.New("pool not found")
	}

	pool := value.(*Pool)
	m.pools.Delete(name)

	log.Printf("[Manager] Starting unregister process for pool %s", name)
	return pool.Close()
}

func (m *Manager) Get(name string) (*Pool, error) {
	value, exists := m.pools.Load(name)
	if !exists {
		return nil, errors.New("pool not found")
	}
	return value.(*Pool), nil
}

func (m *Manager) List() []*Pool {
	pools := make([]*Pool, 0)
	m.pools.Range(func(key, value interface{}) bool {
		pools = append(pools, value.(*Pool))
		return true
	})
	return pools
}

func (m *Manager) GetConnection(name string) (*Conn, error) {
	pool, err := m.Get(name)
	if err != nil {
		return nil, err
	}
	return pool.Get()
}

func (m *Manager) GetAllStats() map[string]PoolStats {
	stats := make(map[string]PoolStats)
	m.pools.Range(func(key, value interface{}) bool {
		pool := value.(*Pool)
		stats[pool.Name()] = pool.GetStats()
		return true
	})
	return stats
}

func (m *Manager) CloseAll() {
	m.pools.Range(func(key, value interface{}) bool {
		pool := value.(*Pool)
		go pool.Close()
		return true
	})
}
