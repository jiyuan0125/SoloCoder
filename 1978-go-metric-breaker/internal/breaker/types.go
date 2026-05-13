package breaker

import "time"

type State string

const (
	StateClosed   State = "Closed"
	StateOpen     State = "Open"
	StateHalfOpen State = "Half-Open"
)

type MetricType string

const (
	MetricErrorRate       MetricType = "error_rate"
	MetricAvgResponseTime MetricType = "avg_response_time"
	MetricTimeoutRate     MetricType = "timeout_rate"
	MetricCustom          MetricType = "custom"
)

type MetricConfig struct {
	Type         MetricType
	Threshold    float64
	TimeoutLimit time.Duration
	CustomName   string
}

type CircuitBreakerConfig struct {
	Name            string
	Metrics         map[MetricType]*MetricConfig
	CustomMetrics   map[string]*MetricConfig
	WindowDuration  time.Duration
	CooldownPeriod  time.Duration
	HalfOpenLimit   int
	FallbackHandler func() (int, interface{})
}

type Snapshot struct {
	TotalRequests   int64
	ErrorRequests   int64
	TimeoutRequests int64
	TotalDuration   time.Duration
	CustomValues    map[string]float64
}

type MetricsStatus struct {
	Type        MetricType
	Name        string
	Current     float64
	Threshold   float64
	Exceeded    bool
	Configured  bool
}

type CircuitBreakerStatus struct {
	Name      string
	State     State
	Metrics   []*MetricsStatus
	Timestamp time.Time
}
