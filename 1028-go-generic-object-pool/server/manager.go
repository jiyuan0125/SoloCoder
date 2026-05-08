package main

import (
	"generic-object-pool/common"
	"generic-object-pool/pool"
	"sync"
	"time"
)

type PoolManager struct {
	mu    sync.RWMutex
	pools map[string]*pool.Pool[*common.PoolObject]
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]*pool.Pool[*common.PoolObject]),
	}
}

func (m *PoolManager) Create(poolID string, maxSize int, idleTimeout time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[poolID]; exists {
		return false
	}

	if maxSize <= 0 {
		maxSize = 100
	}
	if idleTimeout <= 0 {
		idleTimeout = 5 * time.Minute
	}

	factory := func() *common.PoolObject {
		return &common.PoolObject{}
	}

	reset := func(obj *common.PoolObject) {
		obj.Data = ""
	}

	config := pool.Config[*common.PoolObject]{
		MaxSize:      maxSize,
		Factory:      factory,
		Reset:        reset,
		IdleTimeout:  idleTimeout,
		CleanupOnGet: true,
	}

	m.pools[poolID] = pool.New(config)
	return true
}

func (m *PoolManager) Get(poolID string) (*common.PoolObject, bool) {
	m.mu.RLock()
	p, exists := m.pools[poolID]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}

	return p.Get(), true
}

func (m *PoolManager) Put(poolID string, obj *common.PoolObject) bool {
	m.mu.RLock()
	p, exists := m.pools[poolID]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	p.Put(obj)
	return true
}

func (m *PoolManager) Stats(poolID string) (pool.Stats, bool) {
	m.mu.RLock()
	p, exists := m.pools[poolID]
	m.mu.RUnlock()

	if !exists {
		return pool.Stats{}, false
	}

	return p.Stats(), true
}

func (m *PoolManager) Close(poolID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.pools[poolID]
	if !exists {
		return false
	}

	p.Close()
	delete(m.pools, poolID)
	return true
}
