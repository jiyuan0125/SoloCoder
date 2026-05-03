package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

const (
	StateClosed   = "Closed"
	StateOpen     = "Open"
	StateHalfOpen = "HalfOpen"
)

const (
	DefaultFailureThreshold = 5
	DefaultOpenDuration     = 30 * time.Second
	DefaultWaitQueueSize    = 100
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

type CircuitBreaker struct {
	mu               sync.RWMutex
	state            string
	failureCount     int
	failureThreshold int
	openDuration     time.Duration
	lastOpenTime     time.Time
	nextProbeTime    time.Time
	lastError        error
	probeInProgress  bool
	waitQueue        chan struct{}
	onStateChange    func(from, to string)
	onRequest        func(allowed bool, duration time.Duration)
}

type Option func(*CircuitBreaker)

func WithFailureThreshold(threshold int) Option {
	return func(cb *CircuitBreaker) {
		if threshold > 0 {
			cb.failureThreshold = threshold
		}
	}
}

func WithOpenDuration(duration time.Duration) Option {
	return func(cb *CircuitBreaker) {
		if duration > 0 {
			cb.openDuration = duration
		}
	}
}

func New(options ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: DefaultFailureThreshold,
		openDuration:     DefaultOpenDuration,
		waitQueue:        make(chan struct{}, DefaultWaitQueueSize),
	}

	for _, opt := range options {
		opt(cb)
	}

	return cb
}

func (cb *CircuitBreaker) State() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Allow() (bool, error) {
	start := time.Now()
	allowed, err := cb.allowInternal()

	cb.mu.RLock()
	callback := cb.onRequest
	cb.mu.RUnlock()

	if callback != nil {
		duration := time.Since(start)
		callback(allowed, duration)
	}

	return allowed, err
}

func (cb *CircuitBreaker) allowInternal() (bool, error) {
	cb.mu.Lock()

	if cb.state == StateOpen {
		if time.Since(cb.lastOpenTime) >= cb.openDuration {
			oldState := cb.state
			newState := StateHalfOpen
			cb.state = newState
			now := time.Now()
			cb.nextProbeTime = now.Add(cb.openDuration)
			cb.probeInProgress = true
			cb.failureCount = 0
			cb.mu.Unlock()

			cb.mu.RLock()
			stateCallback := cb.onStateChange
			cb.mu.RUnlock()
			if stateCallback != nil {
				stateCallback(oldState, newState)
			}
			return true, nil
		}
		lastErr := cb.lastError
		cb.mu.Unlock()
		return false, lastErr
	}

	if cb.state == StateHalfOpen {
		if !cb.probeInProgress {
			if time.Now().Before(cb.nextProbeTime) {
				cb.mu.Unlock()
				return cb.waitInQueue()
			}
			cb.probeInProgress = true
			cb.mu.Unlock()
			return true, nil
		}
		cb.mu.Unlock()
		return cb.waitInQueue()
	}

	cb.mu.Unlock()
	return true, nil
}

func (cb *CircuitBreaker) waitInQueue() (bool, error) {
	select {
	case cb.waitQueue <- struct{}{}:
		defer func() { <-cb.waitQueue }()

		for {
			cb.mu.Lock()

			if cb.state == StateClosed {
				cb.mu.Unlock()
				return true, nil
			}

			if cb.state == StateOpen {
				if time.Since(cb.lastOpenTime) >= cb.openDuration {
					oldState := cb.state
					newState := StateHalfOpen
					cb.state = newState
					now := time.Now()
					cb.nextProbeTime = now.Add(cb.openDuration)
					cb.probeInProgress = true
					cb.failureCount = 0
					cb.mu.Unlock()

					cb.mu.RLock()
					stateCallback := cb.onStateChange
					cb.mu.RUnlock()
					if stateCallback != nil {
						stateCallback(oldState, newState)
					}
					return true, nil
				}
				lastErr := cb.lastError
				cb.mu.Unlock()
				return false, lastErr
			}

			if cb.state == StateHalfOpen {
				if !cb.probeInProgress {
					if time.Now().Before(cb.nextProbeTime) {
						cb.mu.Unlock()
						time.Sleep(10 * time.Millisecond)
						continue
					}
					cb.probeInProgress = true
					cb.mu.Unlock()
					return true, nil
				}
				cb.mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}

			cb.mu.Unlock()
			return true, nil
		}
	default:
		return false, ErrCircuitOpen
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()

	cb.failureCount = 0
	cb.lastError = nil

	if cb.state == StateHalfOpen {
		cb.probeInProgress = false
		oldState := cb.state
		newState := StateClosed
		cb.state = newState
		cb.mu.Unlock()

		cb.mu.RLock()
		stateCallback := cb.onStateChange
		cb.mu.RUnlock()
		if stateCallback != nil {
			stateCallback(oldState, newState)
		}
		return
	}

	cb.mu.Unlock()
}

func (cb *CircuitBreaker) RecordFailure(err error) {
	cb.mu.Lock()

	cb.lastError = err

	if cb.state == StateClosed {
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			now := time.Now()
			cb.lastOpenTime = now
			oldState := cb.state
			newState := StateOpen
			cb.state = newState
			cb.mu.Unlock()

			cb.mu.RLock()
			stateCallback := cb.onStateChange
			cb.mu.RUnlock()
			if stateCallback != nil {
				stateCallback(oldState, newState)
			}
			return
		}
		cb.mu.Unlock()
		return
	}

	if cb.state == StateHalfOpen {
		cb.probeInProgress = false
		cb.nextProbeTime = cb.nextProbeTime.Add(cb.openDuration)
	}

	cb.mu.Unlock()
}

func (cb *CircuitBreaker) OnStateChange(callback func(from, to string)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = callback
}

func (cb *CircuitBreaker) OnRequest(callback func(allowed bool, duration time.Duration)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onRequest = callback
}
