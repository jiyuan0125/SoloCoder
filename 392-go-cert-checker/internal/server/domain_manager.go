package server

import (
	"sync"
	"time"

	"cert-checker/internal/protocol"
)

type DomainManager struct {
	domains map[string]*protocol.DomainStatus
	mu      sync.RWMutex
}

func NewDomainManager() *DomainManager {
	return &DomainManager{
		domains: make(map[string]*protocol.DomainStatus),
	}
}

func (m *DomainManager) AddDomain(domain string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.domains[domain]; exists {
		return false
	}

	m.domains[domain] = &protocol.DomainStatus{
		Domain:       domain,
		AddedAt:      time.Now(),
		CheckHistory: make([]protocol.CertInfo, 0),
	}
	return true
}

func (m *DomainManager) RemoveDomain(domain string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.domains[domain]; !exists {
		return false
	}

	delete(m.domains, domain)
	return true
}

func (m *DomainManager) ListDomains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]string, 0, len(m.domains))
	for domain := range m.domains {
		domains = append(domains, domain)
	}
	return domains
}

func (m *DomainManager) UpdateCheckResult(domain string, result protocol.CertInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if status, exists := m.domains[domain]; exists {
		status.LastCheck = &result
		status.CheckHistory = append(status.CheckHistory, result)
		if len(status.CheckHistory) > 100 {
			status.CheckHistory = status.CheckHistory[len(status.CheckHistory)-100:]
		}
	}
}

func (m *DomainManager) GetStatus(domain string) (*protocol.DomainStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, exists := m.domains[domain]; exists {
		copyStatus := *status
		return &copyStatus, true
	}
	return nil, false
}

func (m *DomainManager) GetHistory(domain string, limit int) ([]protocol.CertInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, exists := m.domains[domain]; exists {
		history := make([]protocol.CertInfo, len(status.CheckHistory))
		copy(history, status.CheckHistory)

		if limit > 0 && limit < len(history) {
			history = history[len(history)-limit:]
		}

		return history, true
	}
	return nil, false
}
