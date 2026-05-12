package circuitbreaker

import (
	"sync"
	"time"
)

type State string

const (
	StateClosed   State = "Closed"
	StateOpen     State = "Open"
	StateHalfOpen State = "HalfOpen"
	StateUnmonitored State = "未监控"
)

type CallRecord struct {
	Timestamp time.Time
	Success   bool
	Duration  time.Duration
}

type Config struct {
	FailureThreshold int
	Timeout          time.Duration
}

type StateChangeCallback func(serviceName string, fromState, toState State)

type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	config           *Config
	failureCount     int
	lastFailureTime  time.Time
	lastStateChange  time.Time
	callRecords      []CallRecord
	callback         StateChangeCallback
	halfOpenExecuted bool
}

type Manager struct {
	mu         sync.RWMutex
	breakers   map[string]*CircuitBreaker
	defaultCfg *Config
}

func NewManager(defaultFailureThreshold int, defaultTimeout time.Duration) *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
		defaultCfg: &Config{
			FailureThreshold: defaultFailureThreshold,
			Timeout:          defaultTimeout,
		},
	}
}

func (m *Manager) GetOrCreate(serviceName string) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[serviceName]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	breaker, exists = m.breakers[serviceName]
	if exists {
		return breaker
	}

	breaker = &CircuitBreaker{
		state:            StateClosed,
		config:           &Config{FailureThreshold: m.defaultCfg.FailureThreshold, Timeout: m.defaultCfg.Timeout},
		callRecords:      make([]CallRecord, 0, 10),
		lastStateChange:  time.Now(),
		halfOpenExecuted: false,
	}
	m.breakers[serviceName] = breaker
	return breaker
}

func (m *Manager) SetConfig(serviceName string, cfg Config) {
	breaker := m.GetOrCreate(serviceName)
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.config = &Config{
		FailureThreshold: cfg.FailureThreshold,
		Timeout:          cfg.Timeout,
	}
}

func (m *Manager) GetConfig(serviceName string) (*Config, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	breaker, exists := m.breakers[serviceName]
	if !exists {
		return nil, false
	}
	return breaker.config, true
}

func (m *Manager) SetCallback(serviceName string, callback StateChangeCallback) {
	breaker := m.GetOrCreate(serviceName)
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.callback = callback
}

func (m *Manager) GetStatus(serviceName string) (State, []CallRecord) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	breaker, exists := m.breakers[serviceName]
	if !exists {
		return StateUnmonitored, nil
	}

	breaker.mu.RLock()
	defer breaker.mu.RUnlock()

	return breaker.state, append([]CallRecord{}, breaker.callRecords...)
}

func (m *Manager) ListServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]string, 0, len(m.breakers))
	for service := range m.breakers {
		services = append(services, service)
	}
	return services
}

func (cb *CircuitBreaker) changeState(newState State) {
	oldState := cb.state
	if oldState == newState {
		return
	}

	cb.state = newState
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.halfOpenExecuted = false

	if cb.callback != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			cb.callback("", oldState, newState)
		}()
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		if now.Sub(cb.lastStateChange) >= cb.config.Timeout {
			cb.changeState(StateHalfOpen)
			cb.halfOpenExecuted = false
			return true
		}
		return false

	case StateHalfOpen:
		if cb.halfOpenExecuted {
			return false
		}
		cb.halfOpenExecuted = true
		return true

	default:
		return true
	}
}

func (cb *CircuitBreaker) Record(success bool, duration time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	record := CallRecord{
		Timestamp: time.Now(),
		Success:   success,
		Duration:  duration,
	}

	cb.callRecords = append(cb.callRecords, record)
	if len(cb.callRecords) > 10 {
		cb.callRecords = cb.callRecords[len(cb.callRecords)-10:]
	}

	switch cb.state {
	case StateClosed:
		if !success {
			cb.failureCount++
			if cb.failureCount >= cb.config.FailureThreshold {
				cb.changeState(StateOpen)
			}
		} else {
			cb.failureCount = 0
		}

	case StateHalfOpen:
		if success {
			cb.changeState(StateClosed)
		} else {
			cb.changeState(StateOpen)
		}
	}
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) GetConfig() Config {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return *cb.config
}

func (cb *CircuitBreaker) GetCallRecords() []CallRecord {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return append([]CallRecord{}, cb.callRecords...)
}

func (cb *CircuitBreaker) UpdateConfig(cfg Config) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config = &Config{
		FailureThreshold: cfg.FailureThreshold,
		Timeout:          cfg.Timeout,
	}
}

func (cb *CircuitBreaker) SetCallback(callback StateChangeCallback) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.callback = callback
}
