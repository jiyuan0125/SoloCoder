package alert

import (
	"sync"
	"time"

	"monitor-alert/internal/types"
)

type AlertManager struct {
	mu          sync.RWMutex
	activeAlerts map[string]*types.Alert
	alertHistory []types.Alert
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		activeAlerts: make(map[string]*types.Alert),
		alertHistory: make([]types.Alert, 0),
	}
}

func (m *AlertManager) ActiveAlerts() []types.Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	alerts := make([]types.Alert, 0, len(m.activeAlerts))
	for _, a := range m.activeAlerts {
		alerts = append(alerts, *a)
	}
	return alerts
}

func (m *AlertManager) AlertHistory() []types.Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	history := make([]types.Alert, len(m.alertHistory))
	copy(history, m.alertHistory)
	return history
}

func (m *AlertManager) IsActive(ruleName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.activeAlerts[ruleName]
	return exists
}

func (m *AlertManager) GetActive(ruleName string) (*types.Alert, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alert, exists := m.activeAlerts[ruleName]
	if exists {
		cpy := *alert
		return &cpy, true
	}
	return nil, false
}

func (m *AlertManager) ShouldRemind(ruleName string, interval time.Duration) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	alert, exists := m.activeAlerts[ruleName]
	if !exists {
		return false
	}
	if interval <= 0 {
		return false
	}
	return time.Since(alert.LastNotifiedAt) >= interval
}

func (m *AlertManager) Trigger(rule types.AlertRule, message string) *types.Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	alert := &types.Alert{
		RuleName:       rule.Name,
		Severity:       rule.Severity,
		Status:         types.AlertStatusFiring,
		StartedAt:      now,
		LastNotifiedAt: now,
		Message:        message,
	}
	
	m.activeAlerts[rule.Name] = alert
	m.alertHistory = append(m.alertHistory, *alert)
	
	return alert
}

func (m *AlertManager) UpdateLastNotified(ruleName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if alert, exists := m.activeAlerts[ruleName]; exists {
		alert.LastNotifiedAt = time.Now()
	}
}

func (m *AlertManager) Resolve(ruleName string) *types.Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	alert, exists := m.activeAlerts[ruleName]
	if !exists {
		return nil
	}
	
	now := time.Now()
	alert.Status = types.AlertStatusResolved
	alert.ResolvedAt = &now
	
	m.alertHistory = append(m.alertHistory, *alert)
	delete(m.activeAlerts, ruleName)
	
	return alert
}
