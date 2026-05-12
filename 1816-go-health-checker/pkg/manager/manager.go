package manager

import (
	"sync"
	"time"

	"health-checker/pkg/checker"
	"health-checker/pkg/models"
)

const (
	DefaultTimeout  = 5 * time.Second
	DefaultInterval = 10 * time.Second
	HistoryRetention = 10 * time.Minute
	MaxHistoryRecords = 20
	StateChangeThreshold = 2
)

type Manager struct {
	components map[string]*models.Component
	mu         sync.RWMutex
}

func New() *Manager {
	return &Manager{
		components: make(map[string]*models.Component),
	}
}

func (m *Manager) Register(req *models.RegisterRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.components[req.Name]; ok {
		close(existing.StopChan)
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	interval := req.Interval
	if interval <= 0 {
		interval = DefaultInterval
	}

	component := &models.Component{
		Name:         req.Name,
		CheckType:    req.CheckType,
		Address:      req.Address,
		Timeout:      timeout,
		Interval:     interval,
		Status:       models.StatusUnknown,
		RegisterTime: time.Now(),
		CheckResults: make([]models.CheckResult, 0),
		StopChan:     make(chan struct{}),
	}

	m.components[req.Name] = component
	go m.startChecking(component)

	return nil
}

func (m *Manager) Deregister(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	component, ok := m.components[name]
	if !ok {
		return false
	}

	component.Deregistered = true
	component.DeregisterTime = time.Now()

	go func() {
		time.Sleep(HistoryRetention)
		m.mu.Lock()
		defer m.mu.Unlock()
		if c, exists := m.components[name]; exists && c.Deregistered && time.Since(c.DeregisterTime) >= HistoryRetention {
			delete(m.components, name)
		}
	}()

	return true
}

func (m *Manager) Get(name string) (*models.Component, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	component, ok := m.components[name]
	if !ok {
		return nil, false
	}

	if component.Deregistered {
		return nil, false
	}

	return component, true
}

func (m *Manager) GetAll() []*models.Component {
	m.mu.RLock()
	defer m.mu.RUnlock()

	components := make([]*models.Component, 0)
	for _, c := range m.components {
		if !c.Deregistered {
			components = append(components, c)
		}
	}

	return components
}

func (m *Manager) GetOverallStatus() models.ComponentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allHealthy := true
	hasComponents := false

	for _, c := range m.components {
		if c.Deregistered {
			continue
		}
		hasComponents = true

		displayStatus := c.Status
		if displayStatus == models.StatusUnknown {
			displayStatus = models.StatusUnhealthy
		}

		if displayStatus != models.StatusHealthy {
			allHealthy = false
		}
	}

	if !hasComponents {
		return models.StatusHealthy
	}

	if allHealthy {
		return models.StatusHealthy
	}

	return models.StatusUnhealthy
}

func (m *Manager) startChecking(component *models.Component) {
	ticker := time.NewTicker(component.Interval)
	defer ticker.Stop()

	m.doCheck(component)

	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			if component.Deregistered {
				m.mu.RUnlock()
				return
			}
			m.mu.RUnlock()
			m.doCheck(component)
		case <-component.StopChan:
			return
		}
	}
}

func (m *Manager) doCheck(component *models.Component) {
	status, message := checker.Check(component)

	result := models.CheckResult{
		Timestamp: time.Now(),
		Status:    status,
		Message:   message,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	component.LastCheckTime = result.Timestamp
	component.CheckResults = append(component.CheckResults, result)

	if len(component.CheckResults) > MaxHistoryRecords {
		component.CheckResults = component.CheckResults[len(component.CheckResults)-MaxHistoryRecords:]
	}

	if status == models.StatusHealthy {
		component.ConsecutiveGood++
		component.ConsecutiveBad = 0
	} else if status == models.StatusUnhealthy {
		component.ConsecutiveBad++
		component.ConsecutiveGood = 0
	}

	if component.Status == models.StatusUnknown {
		if status == models.StatusHealthy {
			component.ConsecutiveGood++
			if component.ConsecutiveGood >= StateChangeThreshold {
				component.Status = models.StatusHealthy
			}
		} else {
			component.ConsecutiveBad++
			if component.ConsecutiveBad >= StateChangeThreshold {
				component.Status = models.StatusUnhealthy
			}
		}
	} else if component.Status == models.StatusUnhealthy {
		if component.ConsecutiveGood >= StateChangeThreshold {
			component.Status = models.StatusHealthy
			component.ConsecutiveGood = 0
		}
	} else if component.Status == models.StatusHealthy {
		if component.ConsecutiveBad >= StateChangeThreshold {
			component.Status = models.StatusUnhealthy
			component.ConsecutiveBad = 0
		}
	}
}
