package limiter

import (
	"sort"
	"sync"
)

type LimiterManager struct {
	mu           sync.RWMutex
	limiters     map[string]Limiter
	rules        []LimitRule
	sortedRules  []LimitRule
}

func NewLimiterManager() *LimiterManager {
	return &LimiterManager{
		limiters:    make(map[string]Limiter),
		rules:       make([]LimitRule, 0),
		sortedRules: make([]LimitRule, 0),
	}
}

func (m *LimiterManager) sortRules() {
	sorted := make([]LimitRule, len(m.rules))
	copy(sorted, m.rules)
	
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}
		return sorted[i].Window < sorted[j].Window
	})
	
	m.sortedRules = sorted
}

func (m *LimiterManager) AddRule(rule LimitRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i, r := range m.rules {
		if r.Name == rule.Name {
			m.rules[i] = rule
			m.sortRules()
			return
		}
	}
	
	m.rules = append(m.rules, rule)
	m.sortRules()
}

func (m *LimiterManager) RemoveRule(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i, r := range m.rules {
		if r.Name == name {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			delete(m.limiters, name)
			m.sortRules()
			return true
		}
	}
	return false
}

func (m *LimiterManager) GetRules() []LimitRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]LimitRule, len(m.rules))
	copy(result, m.rules)
	return result
}

func (m *LimiterManager) getOrCreateLimiter(rule LimitRule) Limiter {
	if l, exists := m.limiters[rule.Name]; exists {
		return l
	}
	l := NewLimiter(rule)
	m.limiters[rule.Name] = l
	return l
}

func (m *LimiterManager) Allow() (Result, string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.sortedRules) == 0 {
		return Result{Allowed: true}, "", true
	}
	
	for _, rule := range m.sortedRules {
		limiter := m.getOrCreateLimiter(rule)
		result := limiter.Allow()
		if !result.Allowed {
			return result, rule.Name, false
		}
	}
	
	return Result{Allowed: true}, "", true
}

func (m *LimiterManager) AllowN(n int) (Result, string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.sortedRules) == 0 {
		return Result{Allowed: true}, "", true
	}
	
	for _, rule := range m.sortedRules {
		limiter := m.getOrCreateLimiter(rule)
		result := limiter.AllowN(n)
		if !result.Allowed {
			return result, rule.Name, false
		}
	}
	
	return Result{Allowed: true}, "", true
}

func (m *LimiterManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, l := range m.limiters {
		l.Reset()
	}
}

func (m *LimiterManager) Stats() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make(map[string]map[string]interface{})
	for name, l := range m.limiters {
		stats[name] = l.Stats()
	}
	return stats
}

func (m *LimiterManager) GetLimiter(name string) (Limiter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	l, exists := m.limiters[name]
	return l, exists
}
