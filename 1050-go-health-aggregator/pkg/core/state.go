package core

import (
	"sync"
	"time"

	"github.com/health-aggregator/pkg/common"
)

type serviceState struct {
	config             *common.ServiceConfig
	healthy            bool
	currentStatus      common.HealthStatus
	consecutiveSuccess int
	consecutiveFail    int
	history            []*common.ProbeResult
	historySize        int
	mu                 sync.RWMutex
}

func newServiceState(config *common.ServiceConfig, historySize int) *serviceState {
	return &serviceState{
		config:        config,
		healthy:       true,
		currentStatus: common.StatusHealthy,
		history:       make([]*common.ProbeResult, 0, historySize),
		historySize:   historySize,
	}
}

func (s *serviceState) update(result *common.ProbeResult, failureThreshold, recoveryThreshold int) (stateChanged bool, from, to common.HealthStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldStatus := s.currentStatus
	oldHealthy := s.healthy

	if result.Success {
		s.consecutiveSuccess++
		s.consecutiveFail = 0
	} else {
		s.consecutiveFail++
		s.consecutiveSuccess = 0
	}

	if s.healthy {
		if s.consecutiveFail >= failureThreshold {
			s.healthy = false
			s.currentStatus = common.StatusUnhealthy
		}
	} else {
		if s.consecutiveSuccess >= recoveryThreshold {
			s.healthy = true
			s.currentStatus = common.StatusHealthy
		}
	}

	if len(s.history) >= s.historySize {
		s.history = s.history[1:]
	}
	s.history = append(s.history, result)

	if oldStatus != s.currentStatus {
		return true, oldStatus, s.currentStatus
	}

	_ = oldHealthy
	return false, "", ""
}

func (s *serviceState) getStatus() *common.ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &common.ServiceStatus{
		Name:               s.config.Name,
		Group:              s.config.Group,
		ProbeType:          s.config.ProbeType,
		Target:             s.config.Target,
		Healthy:            s.healthy,
		CurrentStatus:      s.currentStatus,
		ConsecutiveSuccess: s.consecutiveSuccess,
		ConsecutiveFail:    s.consecutiveFail,
	}
	if len(s.history) > 0 {
		status.LastProbe = s.history[len(s.history)-1]
	}
	return status
}

func (s *serviceState) getHistory() []*common.ProbeResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*common.ProbeResult, len(s.history))
	copy(results, s.history)
	return results
}

type StateStore struct {
	services        map[string]*serviceState
	events          []*common.StateChangeEvent
	globalConfig    *GlobalConfig
	proberFactory   *ProberFactory
	mu              sync.RWMutex
	maxEvents       int
}

func NewStateStore(globalConfig *GlobalConfig) *StateStore {
	if globalConfig == nil {
		globalConfig = DefaultGlobalConfig()
	}
	return &StateStore{
		services:      make(map[string]*serviceState),
		events:        make([]*common.StateChangeEvent, 0),
		globalConfig:  globalConfig,
		proberFactory: NewProberFactory(),
		maxEvents:     100,
	}
}

func (s *StateStore) AddService(config *common.ServiceConfig) {
	ApplyDefaults(config, s.globalConfig)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[config.Name] = newServiceState(config, s.globalConfig.HistorySize)
}

func (s *StateStore) RemoveService(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.services[name]; exists {
		delete(s.services, name)
		return true
	}
	return false
}

func (s *StateStore) ServiceExists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.services[name]
	return exists
}

func (s *StateStore) GetService(name string) (*common.ServiceConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if state, exists := s.services[name]; exists {
		cfg := *state.config
		return &cfg, true
	}
	return nil, false
}

func (s *StateStore) GetAllServiceNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.services))
	for name := range s.services {
		names = append(names, name)
	}
	return names
}

func (s *StateStore) GetServiceStatus(name string) (*common.ServiceStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if state, exists := s.services[name]; exists {
		return state.getStatus(), true
	}
	return nil, false
}

func (s *StateStore) GetAllStatuses() []*common.ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	statuses := make([]*common.ServiceStatus, 0, len(s.services))
	for _, state := range s.services {
		statuses = append(statuses, state.getStatus())
	}
	return statuses
}

func (s *StateStore) GetServiceHistory(name string) ([]*common.ProbeResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if state, exists := s.services[name]; exists {
		return state.getHistory(), true
	}
	return nil, false
}

func (s *StateStore) GetEvents() []*common.StateChangeEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]*common.StateChangeEvent, len(s.events))
	copy(events, s.events)
	return events
}

func (s *StateStore) UpdateServiceStatus(name string, result *common.ProbeResult) bool {
	s.mu.Lock()
	state, exists := s.services[name]
	if !exists {
		s.mu.Unlock()
		return false
	}
	config := state.config
	s.mu.Unlock()

	stateChanged, from, to := state.update(
		result,
		config.FailureThreshold,
		config.RecoveryThreshold,
	)

	if stateChanged {
		event := &common.StateChangeEvent{
			Name:      name,
			From:      from,
			To:        to,
			Timestamp: time.Now(),
		}
		s.mu.Lock()
		if len(s.events) >= s.maxEvents {
			s.events = s.events[1:]
		}
		s.events = append(s.events, event)
		s.mu.Unlock()
	}

	return true
}

func (s *StateStore) GetProber(probeType common.ProbeType) Prober {
	return s.proberFactory.GetProber(probeType)
}

func (s *StateStore) GetServiceConfig(name string) (*common.ServiceConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if state, exists := s.services[name]; exists {
		cfg := *state.config
		return &cfg, true
	}
	return nil, false
}

func (s *StateStore) GetGlobalConfig() *GlobalConfig {
	return s.globalConfig
}
