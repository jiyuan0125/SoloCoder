package breaker

import (
	"sync"
	"time"
)

type CircuitBreaker struct {
	mu            sync.RWMutex
	config        *CircuitBreakerConfig
	state         State
	openedAt      time.Time
	halfOpenCount int
	window        *slidingWindow
}

func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{}
	}
	if config.Metrics == nil {
		config.Metrics = make(map[MetricType]*MetricConfig)
	}
	if config.CustomMetrics == nil {
		config.CustomMetrics = make(map[string]*MetricConfig)
	}
	if config.WindowDuration <= 0 {
		config.WindowDuration = time.Second * 10
	}
	if config.CooldownPeriod <= 0 {
		config.CooldownPeriod = time.Second * 30
	}
	if config.HalfOpenLimit <= 0 {
		config.HalfOpenLimit = 5
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		window: newSlidingWindow(config.WindowDuration),
	}
}

func (cb *CircuitBreaker) Name() string {
	return cb.config.Name
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) getTimeoutLimit() time.Duration {
	if cfg, ok := cb.config.Metrics[MetricTimeoutRate]; ok {
		return cfg.TimeoutLimit
	}
	return 0
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.config.CooldownPeriod {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			cb.window.Reset()
			return true
		}
		return false
	case StateHalfOpen:
		return cb.halfOpenCount < cb.config.HalfOpenLimit
	}
	return true
}

func (cb *CircuitBreaker) Record(start, end time.Time, statusCode int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.window.Record(start, end, statusCode, cb.getTimeoutLimit())

	switch cb.state {
	case StateClosed:
		if cb.shouldOpen() {
			cb.state = StateOpen
			cb.openedAt = time.Now()
		}
	case StateHalfOpen:
		cb.halfOpenCount++
		if statusCode >= 500 {
			cb.state = StateOpen
			cb.openedAt = time.Now()
			cb.window.Reset()
		} else if cb.halfOpenCount >= cb.config.HalfOpenLimit {
			if cb.shouldOpen() {
				cb.state = StateOpen
				cb.openedAt = time.Now()
			} else {
				cb.state = StateClosed
			}
		}
	}
}

func (cb *CircuitBreaker) RecordCustom(name string, value float64) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.window.RecordCustom(name, value)

	if cb.state == StateClosed && cb.shouldOpen() {
		cb.state = StateOpen
		cb.openedAt = time.Now()
	}
}

func (cb *CircuitBreaker) shouldOpen() bool {
	configuredCount := len(cb.config.Metrics) + len(cb.config.CustomMetrics)
	if configuredCount == 0 {
		return false
	}

	snap := cb.window.Snapshot()

	for mType, cfg := range cb.config.Metrics {
		if !cb.checkMetric(mType, cfg, snap) {
			return false
		}
	}

	for name, cfg := range cb.config.CustomMetrics {
		if !cb.checkCustomMetric(name, cfg, snap) {
			return false
		}
	}

	return true
}

func (cb *CircuitBreaker) checkMetric(mType MetricType, cfg *MetricConfig, snap *Snapshot) bool {
	switch mType {
	case MetricErrorRate:
		if snap.TotalRequests == 0 {
			return false
		}
		rate := float64(snap.ErrorRequests) / float64(snap.TotalRequests)
		return rate >= cfg.Threshold
	case MetricAvgResponseTime:
		if snap.TotalRequests == 0 {
			return false
		}
		avgMs := float64(snap.TotalDuration.Milliseconds()) / float64(snap.TotalRequests)
		return avgMs >= cfg.Threshold
	case MetricTimeoutRate:
		if snap.TotalRequests == 0 {
			return false
		}
		rate := float64(snap.TimeoutRequests) / float64(snap.TotalRequests)
		return rate >= cfg.Threshold
	}
	return false
}

func (cb *CircuitBreaker) checkCustomMetric(name string, cfg *MetricConfig, snap *Snapshot) bool {
	val, ok := snap.CustomValues[name]
	if !ok {
		return false
	}
	return val >= cfg.Threshold
}

func (cb *CircuitBreaker) UpdateMetrics(metrics []*MetricConfig) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config.Metrics = make(map[MetricType]*MetricConfig)
	cb.config.CustomMetrics = make(map[string]*MetricConfig)

	for _, m := range metrics {
		if m.Type == MetricCustom {
			if m.CustomName != "" {
				cb.config.CustomMetrics[m.CustomName] = m
			}
		} else {
			cb.config.Metrics[m.Type] = m
		}
	}
}

func (cb *CircuitBreaker) Status() *CircuitBreakerStatus {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	snap := cb.window.Snapshot()
	metrics := make([]*MetricsStatus, 0)

	allTypes := []MetricType{MetricErrorRate, MetricAvgResponseTime, MetricTimeoutRate}
	for _, mType := range allTypes {
		cfg, configured := cb.config.Metrics[mType]
		ms := &MetricsStatus{
			Type:       mType,
			Configured: configured,
		}

		if snap.TotalRequests > 0 {
			switch mType {
			case MetricErrorRate:
				ms.Current = float64(snap.ErrorRequests) / float64(snap.TotalRequests)
			case MetricAvgResponseTime:
				ms.Current = float64(snap.TotalDuration.Milliseconds()) / float64(snap.TotalRequests)
			case MetricTimeoutRate:
				ms.Current = float64(snap.TimeoutRequests) / float64(snap.TotalRequests)
			}
		}

		if configured {
			ms.Threshold = cfg.Threshold
			ms.Exceeded = ms.Current >= cfg.Threshold
		}

		metrics = append(metrics, ms)
	}

	for name, cfg := range cb.config.CustomMetrics {
		ms := &MetricsStatus{
			Type:       MetricCustom,
			Name:       name,
			Configured: true,
			Threshold:  cfg.Threshold,
		}
		if val, ok := snap.CustomValues[name]; ok {
			ms.Current = val
			ms.Exceeded = val >= cfg.Threshold
		}
		metrics = append(metrics, ms)
	}

	return &CircuitBreakerStatus{
		Name:      cb.config.Name,
		State:     cb.state,
		Metrics:   metrics,
		Timestamp: time.Now(),
	}
}

func (cb *CircuitBreaker) GetFallback() (int, interface{}) {
	if cb.config.FallbackHandler != nil {
		return cb.config.FallbackHandler()
	}
	return 503, map[string]string{"error": "service unavailable"}
}
