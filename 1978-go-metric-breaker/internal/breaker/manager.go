package breaker

import "sync"

type Manager struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

func (m *Manager) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	m.mu.RLock()
	if cb, ok := m.breakers[name]; ok {
		m.mu.RUnlock()
		return cb
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, ok := m.breakers[name]; ok {
		return cb
	}

	if config == nil {
		config = &CircuitBreakerConfig{}
	}
	config.Name = name

	cb := NewCircuitBreaker(config)
	m.breakers[name] = cb
	return cb
}

func (m *Manager) Get(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cb, ok := m.breakers[name]
	return cb, ok
}

func (m *Manager) All() []*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*CircuitBreaker, 0, len(m.breakers))
	for _, cb := range m.breakers {
		result = append(result, cb)
	}
	return result
}
