package circuitbreaker

import (
	"log"
	"sync"
	"time"
)

type Manager struct {
	mu        sync.RWMutex
	breakers  map[string]*CircuitBreaker
}

func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

func (m *Manager) Register(config *Config) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cb, ok := m.breakers[config.ServiceName]; ok {
		return cb
	}
	cb := NewCircuitBreaker(config)
	m.breakers[config.ServiceName] = cb
	log.Printf("[INFO] Registered circuit breaker for service: %s", config.ServiceName)
	return cb
}

func (m *Manager) Get(serviceName string) *CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.breakers[serviceName]
}

func (m *Manager) List() []*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*CircuitBreaker, 0, len(m.breakers))
	for _, cb := range m.breakers {
		list = append(list, cb)
	}
	return list
}

func (m *Manager) Reset(serviceName string, operatorIP string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	cb, ok := m.breakers[serviceName]
	if !ok {
		return false
	}
	log.Printf("[AUDIT] [%s] Manual reset requested by %s for service: %s", time.Now().Format(time.RFC3339), operatorIP, serviceName)
	cb.Reset()
	return true
}

func (m *Manager) Allow(serviceName string) bool {
	cb := m.Get(serviceName)
	if cb == nil {
		return true
	}
	return cb.Allow()
}

func (m *Manager) Record(serviceName string, success bool, duration time.Duration) {
	cb := m.Get(serviceName)
	if cb == nil {
		return
	}
	cb.Record(success, duration)
}
