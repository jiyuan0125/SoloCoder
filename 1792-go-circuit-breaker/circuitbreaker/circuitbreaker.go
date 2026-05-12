package circuitbreaker

import (
	"log"
	"sync"
	"time"
)

type State string

const (
	StateClosed   State = "Closed"
	StateOpen     State = "Open"
	StateHalfOpen State = "HalfOpen"
)

type Config struct {
	ServiceName          string
	FailureThreshold     int
	OpenDuration         time.Duration
	HalfOpenTimeout      time.Duration
	OnStateChange        func(serviceName string, from, to State)
	OnFailure            func(serviceName string)
	OnSuccess            func(serviceName string)
}

type CallRecord struct {
	Timestamp time.Time
	Success   bool
	Duration  time.Duration
}

type CircuitBreaker struct {
	mu                 sync.RWMutex
	serviceName        string
	state              State
	failureCount       int
	lastFailureTime    time.Time
	openUntil          time.Time
	halfOpenInFlight   bool
	callRecords        []*CallRecord
	config             *Config
}

func NewCircuitBreaker(config *Config) *CircuitBreaker {
	return &CircuitBreaker{
		serviceName:      config.ServiceName,
		state:            StateClosed,
		config:           config,
		callRecords:      make([]*CallRecord, 0, 20),
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) ServiceName() string {
	return cb.serviceName
}

func (cb *CircuitBreaker) Config() *Config {
	return cb.config
}

func (cb *CircuitBreaker) CallRecords() []*CallRecord {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	records := make([]*CallRecord, len(cb.callRecords))
	for i, r := range cb.callRecords {
		records[i] = &CallRecord{
			Timestamp: r.Timestamp,
			Success:   r.Success,
			Duration:  r.Duration,
		}
	}
	return records
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Now().After(cb.openUntil) {
			cb.transitionState(StateHalfOpen)
			cb.halfOpenInFlight = true
			return true
		}
		return false
	case StateHalfOpen:
		if !cb.halfOpenInFlight {
			cb.halfOpenInFlight = true
			return true
		}
		return false
	}
	return false
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	from := cb.state
	cb.state = StateClosed
	cb.failureCount = 0
	cb.halfOpenInFlight = false
	cb.lastFailureTime = time.Time{}
	cb.openUntil = time.Time{}
	if from != StateClosed {
		cb.triggerOnStateChange(from, StateClosed)
	}
}

func (cb *CircuitBreaker) Record(success bool, duration time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	record := &CallRecord{
		Timestamp: time.Now(),
		Success:   success,
		Duration:  duration,
	}
	if len(cb.callRecords) >= 20 {
		cb.callRecords = cb.callRecords[1:]
	}
	cb.callRecords = append(cb.callRecords, record)

	switch cb.state {
	case StateClosed:
		if success {
			cb.failureCount = 0
			if cb.config.OnSuccess != nil {
				cb.config.OnSuccess(cb.serviceName)
			}
		} else {
			cb.failureCount++
			cb.lastFailureTime = time.Now()
			if cb.config.OnFailure != nil {
				cb.config.OnFailure(cb.serviceName)
			}
			threshold := cb.config.FailureThreshold
			if threshold <= 0 {
				threshold = 5
			}
			if cb.failureCount >= threshold {
				cb.transitionState(StateOpen)
				openDuration := cb.config.OpenDuration
				if openDuration <= 0 {
					openDuration = 30 * time.Second
				}
				cb.openUntil = time.Now().Add(openDuration)
			}
		}
	case StateHalfOpen:
		cb.halfOpenInFlight = false
		if success {
			cb.transitionState(StateClosed)
			cb.failureCount = 0
			if cb.config.OnSuccess != nil {
				cb.config.OnSuccess(cb.serviceName)
			}
		} else {
			cb.transitionState(StateOpen)
			openDuration := cb.config.OpenDuration
			if openDuration <= 0 {
				openDuration = 30 * time.Second
			}
			cb.openUntil = time.Now().Add(openDuration)
			if cb.config.OnFailure != nil {
				cb.config.OnFailure(cb.serviceName)
			}
		}
	}
}

func (cb *CircuitBreaker) transitionState(to State) {
	from := cb.state
	cb.state = to
	cb.triggerOnStateChange(from, to)
}

func (cb *CircuitBreaker) triggerOnStateChange(from, to State) {
	if cb.config.OnStateChange != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[WARN] State change callback panicked for service %s: %v", cb.serviceName, r)
				}
			}()
			cb.config.OnStateChange(cb.serviceName, from, to)
		}()
	}
}
