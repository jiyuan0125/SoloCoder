package limiter

import (
	"sync"

	"multirate-limiter/common"
)

type limiterKey struct {
	path     string
	clientID string
	hasID    bool
}

type pathConfig struct {
	rules []common.Rule
}

type Manager struct {
	mu        sync.RWMutex
	configs   map[string]*pathConfig
	limiters  map[limiterKey]*MultiLimiter
}

func NewManager() *Manager {
	return &Manager{
		configs:  make(map[string]*pathConfig),
		limiters: make(map[limiterKey]*MultiLimiter),
	}
}

func (m *Manager) SetRules(path string, rules []common.Rule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[path] = &pathConfig{rules: rules}

	for key := range m.limiters {
		if key.path == path {
			delete(m.limiters, key)
		}
	}
}

func (m *Manager) GetRules(path string) ([]common.Rule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, ok := m.configs[path]
	if !ok {
		return nil, false
	}
	return cfg.rules, true
}

func (m *Manager) getOrCreateLimiter(path string, clientID string, hasClientID bool) (*MultiLimiter, bool) {
	m.mu.RLock()
	cfg, ok := m.configs[path]
	if !ok {
		m.mu.RUnlock()
		return nil, false
	}
	rules := cfg.rules
	m.mu.RUnlock()

	key := limiterKey{
		path:     path,
		clientID: clientID,
		hasID:    hasClientID,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if limiter, exists := m.limiters[key]; exists {
		return limiter, true
	}

	limiter := NewMultiLimiter(rules)
	m.limiters[key] = limiter
	return limiter, true
}

func (m *Manager) Allow(path string, clientID string, hasClientID bool) (allowed bool, results []common.RuleResult, found bool) {
	limiter, ok := m.getOrCreateLimiter(path, clientID, hasClientID)
	if !ok {
		return false, nil, false
	}
	allowed, results = limiter.Allow()
	return allowed, results, true
}

func (m *Manager) Stats(path string, clientID string, hasClientID bool) (stats []common.RuleResult, found bool) {
	limiter, ok := m.getOrCreateLimiter(path, clientID, hasClientID)
	if !ok {
		return nil, false
	}
	return limiter.Stats(), true
}
