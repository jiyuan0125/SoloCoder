package circuit

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"circuit-breaker/models"
)

var ErrBreakerOpen = errors.New("circuit breaker is open")

type Breaker struct {
	mu          sync.RWMutex
	name        string
	state       models.State
	config      models.BreakerConfig
	windowStats *models.WindowStats
	
	openUntil   time.Time
	lastTriggeredAt *time.Time
	probeSuccess int
	
	eventListeners []func(event models.Event)
}

type BreakerManager struct {
	mu        sync.RWMutex
	breakers  map[string]*Breaker
}

func NewBreakerManager() *BreakerManager {
	return &BreakerManager{
		breakers: make(map[string]*Breaker),
	}
}

func (bm *BreakerManager) Create(name string, config models.BreakerConfig) (*Breaker, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if _, exists := bm.breakers[name]; exists {
		return nil, fmt.Errorf("breaker with name %q already exists", name)
	}
	
	b := &Breaker{
		name:        name,
		state:       models.Closed,
		config:      config,
		windowStats: models.NewWindowStats(config.WindowSize),
	}
	
	bm.breakers[name] = b
	return b, nil
}

func (bm *BreakerManager) Get(name string) (*Breaker, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	b, ok := bm.breakers[name]
	return b, ok
}

func (bm *BreakerManager) UpdateConfig(name string, config models.BreakerConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	b, ok := bm.breakers[name]
	if !ok {
		return fmt.Errorf("breaker with name %q not found", name)
	}
	
	b.mu.Lock()
	b.config = config
	b.windowStats = models.NewWindowStats(config.WindowSize)
	b.mu.Unlock()
	
	return nil
}

func (bm *BreakerManager) List() []*Breaker {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	list := make([]*Breaker, 0, len(bm.breakers))
	for _, b := range bm.breakers {
		list = append(list, b)
	}
	return list
}

func (b *Breaker) Name() string {
	return b.name
}

func (b *Breaker) State() models.State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

func (b *Breaker) Config() models.BreakerConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

func (b *Breaker) AddEventListener(listener func(event models.Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.eventListeners = append(b.eventListeners, listener)
}

func (b *Breaker) RemoveEventListener(listener func(event models.Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for i, l := range b.eventListeners {
		if fmt.Sprintf("%p", l) == fmt.Sprintf("%p", listener) {
			b.eventListeners = append(b.eventListeners[:i], b.eventListeners[i+1:]...)
			return
		}
	}
}

func (b *Breaker) notifyListeners(event models.Event) {
	b.mu.RLock()
	listeners := make([]func(event models.Event), len(b.eventListeners))
	copy(listeners, b.eventListeners)
	b.mu.RUnlock()
	
	for _, listener := range listeners {
		go listener(event)
	}
}

func (b *Breaker) transitionTo(newState models.State, reason string) {
	b.mu.Lock()
	oldState := b.state
	b.state = newState
	
	if newState == models.Open {
		now := time.Now()
		b.openUntil = now.Add(b.config.CoolDownPeriod)
		b.lastTriggeredAt = &now
	} else if newState == models.HalfOpen {
		b.probeSuccess = 0
		b.windowStats.Clear()
	} else if newState == models.Closed {
		b.windowStats.Clear()
	}
	b.mu.Unlock()
	
	if oldState != newState {
		event := models.Event{
			BreakerName: b.name,
			FromState:   oldState,
			ToState:     newState,
			Timestamp:   time.Now(),
			Reason:      reason,
		}
		b.notifyListeners(event)
	}
}

func (b *Breaker) checkThresholds() bool {
	metrics := b.windowStats.Current()
	
	if metrics.Total < int64(b.config.MinRequests) {
		return false
	}
	
	failureRate := float64(metrics.Failures) / float64(metrics.Total)
	slowRatio := float64(metrics.SlowCalls) / float64(metrics.Total)
	
	return failureRate >= b.config.FailureThreshold || slowRatio >= b.config.SlowRatio
}

func (b *Breaker) Allow() (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	switch b.state {
	case models.Closed:
		return true, nil
		
	case models.Open:
		if time.Now().After(b.openUntil) {
			b.state = models.HalfOpen
			b.probeSuccess = 0
			return true, nil
		}
		return false, ErrBreakerOpen
		
	case models.HalfOpen:
		if b.probeSuccess >= b.config.ProbeCount {
			return false, ErrBreakerOpen
		}
		return true, nil
		
	default:
		return true, nil
	}
}

func (b *Breaker) Success(duration time.Duration) {
	now := time.Now()
	isSlow := duration >= b.config.SlowThreshold
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.state == models.HalfOpen {
		b.probeSuccess++
		if b.probeSuccess >= b.config.ProbeCount {
			b.state = models.Closed
			b.windowStats.Clear()
			
			event := models.Event{
				BreakerName: b.name,
				FromState:   models.HalfOpen,
				ToState:     models.Closed,
				Timestamp:   time.Now(),
				Reason:      "all probes succeeded",
			}
			b.notifyListeners(event)
		}
		return
	}
	
	b.windowStats.Record(now, true, isSlow)
	
	if b.state == models.Closed && b.checkThresholds() {
		b.state = models.Open
		now := time.Now()
		b.openUntil = now.Add(b.config.CoolDownPeriod)
		b.lastTriggeredAt = &now
		
		event := models.Event{
			BreakerName: b.name,
			FromState:   models.Closed,
			ToState:     models.Open,
			Timestamp:   time.Now(),
			Reason:      "threshold exceeded",
		}
		b.notifyListeners(event)
	}
}

func (b *Breaker) Failure(duration time.Duration) {
	now := time.Now()
	isSlow := duration >= b.config.SlowThreshold
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.state == models.HalfOpen {
		b.state = models.Open
		now := time.Now()
		b.openUntil = now.Add(b.config.CoolDownPeriod)
		b.lastTriggeredAt = &now
		
		event := models.Event{
			BreakerName: b.name,
			FromState:   models.HalfOpen,
			ToState:     models.Open,
			Timestamp:   time.Now(),
			Reason:      "probe failed",
		}
		b.notifyListeners(event)
		return
	}
	
	b.windowStats.Record(now, false, isSlow)
	
	if b.state == models.Closed && b.checkThresholds() {
		b.state = models.Open
		now := time.Now()
		b.openUntil = now.Add(b.config.CoolDownPeriod)
		b.lastTriggeredAt = &now
		
		event := models.Event{
			BreakerName: b.name,
			FromState:   models.Closed,
			ToState:     models.Open,
			Timestamp:   time.Now(),
			Reason:      "threshold exceeded",
		}
		b.notifyListeners(event)
	}
}

func (b *Breaker) GetStatus() models.BreakerState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	return models.BreakerState{
		Name:            b.name,
		State:           b.state,
		Config:          b.config,
		Metrics:         b.windowStats.Current(),
		LastTriggeredAt: b.lastTriggeredAt,
		ProbeSuccess:    b.probeSuccess,
		OpenUntil:       b.openUntil,
	}
}
