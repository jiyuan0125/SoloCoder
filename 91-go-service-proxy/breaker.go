package main

import (
	"sync"
	"time"
)

type BreakerState int

const (
	StateClosed BreakerState = iota
	StateOpen
	StateHalfOpen
)

const (
	FailureThreshold = 3
	Timeout          = 30 * time.Second
)

type Breaker struct {
	state           BreakerState
	failureCount    int
	lastFailureTime time.Time
	mu              sync.Mutex
}

func NewBreaker() *Breaker {
	return &Breaker{
		state: StateClosed,
	}
}

func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastFailureTime) >= Timeout {
			b.state = StateHalfOpen
			b.failureCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return true
	}
}

func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failureCount = 0
	if b.state == StateHalfOpen {
		b.state = StateClosed
	}
}

func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failureCount++
	b.lastFailureTime = time.Now()

	if b.state == StateClosed {
		if b.failureCount >= FailureThreshold {
			b.state = StateOpen
		}
	} else if b.state == StateHalfOpen {
		b.state = StateOpen
	}
}

func (b *Breaker) State() BreakerState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
