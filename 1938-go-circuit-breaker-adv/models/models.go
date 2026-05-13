package models

import (
	"sync"
	"time"
)

type State string

const (
	Closed   State = "closed"
	Open     State = "open"
	HalfOpen State = "half-open"
)

type BreakerConfig struct {
	WindowSize       time.Duration `json:"window_size"`
	FailureThreshold float64       `json:"failure_threshold"`
	SlowThreshold    time.Duration `json:"slow_threshold"`
	SlowRatio        float64       `json:"slow_ratio"`
	CoolDownPeriod   time.Duration `json:"cooldown_period"`
	ProbeCount       int           `json:"probe_count"`
	MinRequests      int           `json:"min_requests"`
}

func DefaultConfig() BreakerConfig {
	return BreakerConfig{
		WindowSize:       60 * time.Second,
		FailureThreshold: 0.5,
		SlowThreshold:    3 * time.Second,
		SlowRatio:        0.5,
		CoolDownPeriod:   30 * time.Second,
		ProbeCount:       3,
		MinRequests:      10,
	}
}

type RequestMetrics struct {
	Total     int64
	Failures  int64
	SlowCalls int64
}

type WindowStats struct {
	sync.RWMutex
	Buckets map[int64]*RequestMetrics
	Size    time.Duration
}

func NewWindowStats(size time.Duration) *WindowStats {
	return &WindowStats{
		Buckets: make(map[int64]*RequestMetrics),
		Size:    size,
	}
}

func (w *WindowStats) bucket(now time.Time) int64 {
	return now.Unix() % int64(w.Size.Seconds())
}

func (w *WindowStats) Record(now time.Time, success bool, slow bool) {
	w.Lock()
	defer w.Unlock()
	
	b := w.bucket(now)
	if _, ok := w.Buckets[b]; !ok {
		w.Buckets[b] = &RequestMetrics{}
	}
	
	w.Buckets[b].Total++
	if !success {
		w.Buckets[b].Failures++
	}
	if slow {
		w.Buckets[b].SlowCalls++
	}
}

func (w *WindowStats) Clear() {
	w.Lock()
	defer w.Unlock()
	w.Buckets = make(map[int64]*RequestMetrics)
}

func (w *WindowStats) Current() RequestMetrics {
	w.RLock()
	defer w.RUnlock()
	
	var total RequestMetrics
	for _, m := range w.Buckets {
		total.Total += m.Total
		total.Failures += m.Failures
		total.SlowCalls += m.SlowCalls
	}
	return total
}

type BreakerState struct {
	Name             string
	State            State
	Config           BreakerConfig
	Metrics          RequestMetrics
	LastTriggeredAt  *time.Time
	ProbeSuccess     int
	OpenUntil        time.Time
}

type Subscriber struct {
	ID      string
	URL     string
	Created time.Time
}

type Event struct {
	BreakerName string    `json:"breaker_name"`
	FromState   State     `json:"from_state"`
	ToState     State     `json:"to_state"`
	Timestamp   time.Time `json:"timestamp"`
	Reason      string    `json:"reason,omitempty"`
}
