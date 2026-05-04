package circuitbreaker

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type circuitBreaker struct {
	name           string
	config         Config
	state          State
	mu             sync.RWMutex
	window         *SlidingWindow
	totalRequests  int64
	successCount   int64
	failureCount   int64
	consecutiveSuccesses int
	openAt         time.Time
	lastStateChange time.Time
	halfOpenRequests int
	halfOpenSuccesses int
	halfOpenFailures int
	callbacks      []StateChangeCallback
}

func New(name string, config ...Config) (CircuitBreaker, error) {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	cb := &circuitBreaker{
		name:   name,
		config: cfg,
		state:  StateClosed,
		window: NewSlidingWindow(cfg.WindowSize),
	}

	return cb, nil
}

func NewFromFile(name, path string, config ...Config) (CircuitBreaker, error) {
	cb, err := New(name, config...)
	if err != nil {
		return nil, err
	}

	if err := cb.LoadFromFile(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return cb, nil
}

func (cb *circuitBreaker) Name() string {
	return cb.name
}

func (cb *circuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateOpen {
		if time.Since(cb.openAt) >= cb.config.OpenTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == StateOpen && time.Since(cb.openAt) >= cb.config.OpenTimeout {
				cb.transitionTo(StateHalfOpen)
			}
			cb.mu.Unlock()
			cb.mu.RLock()
		}
	}

	return cb.state
}

func (cb *circuitBreaker) Allow() (bool, error) {
	state := cb.State()

	switch state {
	case StateOpen:
		cb.mu.RLock()
		openAt := cb.openAt
		cb.mu.RUnlock()
		return false, &CircuitOpenError{
			CircuitName: cb.name,
			OpenSince:   openAt,
		}
	case StateHalfOpen:
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if cb.halfOpenRequests >= cb.config.HalfOpenMaxRequests {
			return false, &CircuitOpenError{
				CircuitName: cb.name,
				OpenSince:   cb.openAt,
			}
		}
		cb.halfOpenRequests++
		return true, nil
	default:
		return true, nil
	}
}

func (cb *circuitBreaker) MarkSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++
	cb.successCount++
	cb.consecutiveSuccesses++

	if cb.state == StateClosed {
		cb.window.Add(ResultSuccess)
		if cb.consecutiveSuccesses >= cb.config.SuccessThreshold {
			cb.window.Reset()
			cb.consecutiveSuccesses = 0
		}
	} else if cb.state == StateHalfOpen {
		cb.halfOpenSuccesses++
		if cb.halfOpenRequests >= cb.config.HalfOpenMaxRequests {
			if cb.halfOpenFailures == 0 {
				cb.transitionTo(StateClosed)
			}
		}
	}
}

func (cb *circuitBreaker) MarkFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++
	cb.failureCount++
	cb.consecutiveSuccesses = 0

	if cb.state == StateClosed {
		cb.window.Add(ResultFailure)
		failures := cb.window.Failures()
		total := cb.window.Total()

		if total >= cb.config.FailureThreshold && failures >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
		}
	} else if cb.state == StateHalfOpen {
		cb.halfOpenFailures++
		if cb.halfOpenRequests >= cb.config.HalfOpenMaxRequests || cb.halfOpenFailures > 0 {
			cb.transitionTo(StateOpen)
		}
	}
}

func (cb *circuitBreaker) Execute(fn func() error) error {
	allowed, err := cb.Allow()
	if !allowed {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			cb.MarkFailure()
			panic(r)
		}
	}()

	err = fn()
	if err != nil {
		cb.MarkFailure()
		return err
	}

	cb.MarkSuccess()
	return nil
}

func (cb *circuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state != StateClosed {
		cb.transitionTo(StateClosed)
	}
}

func (cb *circuitBreaker) ForceState(state State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state != state {
		cb.transitionTo(state)
	}
}

func (cb *circuitBreaker) GetStatistics() Statistics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Statistics{
		TotalRequests:   cb.totalRequests,
		SuccessCount:    cb.successCount,
		FailureCount:    cb.failureCount,
		RecentFailures:  cb.window.Failures(),
		RecentSuccesses: cb.window.Successes(),
	}
}

func (cb *circuitBreaker) OnStateChange(callback StateChangeCallback) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.callbacks = append(cb.callbacks, callback)
}

func (cb *circuitBreaker) GetConfig() Config {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.config
}

func (cb *circuitBreaker) GetOpenAt() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.openAt
}

func (cb *circuitBreaker) GetLastStateChange() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastStateChange
}

func (cb *circuitBreaker) transitionTo(newState State) {
	oldState := cb.state
	if oldState == newState {
		return
	}

	cb.state = newState
	cb.lastStateChange = time.Now()

	switch newState {
	case StateOpen:
		cb.openAt = time.Now()
	case StateHalfOpen:
		cb.halfOpenRequests = 0
		cb.halfOpenSuccesses = 0
		cb.halfOpenFailures = 0
	case StateClosed:
		cb.window.Reset()
		cb.halfOpenRequests = 0
		cb.halfOpenSuccesses = 0
		cb.halfOpenFailures = 0
		cb.consecutiveSuccesses = 0
	}

	if len(cb.callbacks) > 0 {
		event := StateChangeEvent{
			Name:      cb.name,
			From:      oldState,
			To:        newState,
			Timestamp: time.Now(),
		}
		for _, cbk := range cb.callbacks {
			cbk(event)
		}
	}
}

type persistedState struct {
	Name              string
	Config            Config
	State             State
	TotalRequests     int64
	SuccessCount      int64
	FailureCount      int64
	ConsecutiveSuccesses int
	OpenAt            time.Time
	LastStateChange   time.Time
	HalfOpenRequests  int
	HalfOpenSuccesses int
	HalfOpenFailures  int
	WindowRecords     []WindowRecord
}

func (cb *circuitBreaker) SaveToFile(path string) error {
	cb.mu.RLock()
	state := persistedState{
		Name:                 cb.name,
		Config:               cb.config,
		State:                cb.state,
		TotalRequests:        cb.totalRequests,
		SuccessCount:         cb.successCount,
		FailureCount:         cb.failureCount,
		ConsecutiveSuccesses: cb.consecutiveSuccesses,
		OpenAt:               cb.openAt,
		LastStateChange:      cb.lastStateChange,
		HalfOpenRequests:     cb.halfOpenRequests,
		HalfOpenSuccesses:    cb.halfOpenSuccesses,
		HalfOpenFailures:     cb.halfOpenFailures,
		WindowRecords:        cb.window.GetRecords(),
	}
	cb.mu.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (cb *circuitBreaker) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if state.Name != cb.name {
		return fmt.Errorf("circuit name mismatch: expected %q, got %q", cb.name, state.Name)
	}

	cb.config = state.Config
	cb.state = state.State
	cb.totalRequests = state.TotalRequests
	cb.successCount = state.SuccessCount
	cb.failureCount = state.FailureCount
	cb.consecutiveSuccesses = state.ConsecutiveSuccesses
	cb.openAt = state.OpenAt
	cb.lastStateChange = state.LastStateChange
	cb.halfOpenRequests = state.HalfOpenRequests
	cb.halfOpenSuccesses = state.HalfOpenSuccesses
	cb.halfOpenFailures = state.HalfOpenFailures

	cb.window = NewSlidingWindow(state.Config.WindowSize)
	for _, record := range state.WindowRecords {
		cb.window.Add(record.Result)
	}

	return nil
}
