package common

import "time"

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type ProbeType string

const (
	ProbeTypeHTTP ProbeType = "http"
	ProbeTypeTCP  ProbeType = "tcp"
)

type ProbeResult struct {
	Success   bool          `json:"success"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
}

type ServiceStatus struct {
	Name               string        `json:"name"`
	Group              string        `json:"group,omitempty"`
	ProbeType          ProbeType     `json:"probe_type"`
	Target             string        `json:"target"`
	Healthy            bool          `json:"healthy"`
	CurrentStatus      HealthStatus  `json:"current_status"`
	ConsecutiveSuccess int           `json:"consecutive_success"`
	ConsecutiveFail    int           `json:"consecutive_fail"`
	LastProbe          *ProbeResult  `json:"last_probe,omitempty"`
}

type ProbeHistory struct {
	Total   int            `json:"total"`
	Results []*ProbeResult `json:"results"`
}

type StateChangeEvent struct {
	Name        string       `json:"name"`
	From        HealthStatus `json:"from"`
	To          HealthStatus `json:"to"`
	Timestamp   time.Time    `json:"timestamp"`
}

type AggregateStatus struct {
	OverallStatus    HealthStatus         `json:"overall_status"`
	TotalServices    int                  `json:"total_services"`
	HealthyCount     int                  `json:"healthy_count"`
	UnhealthyCount   int                  `json:"unhealthy_count"`
	ServiceStatuses  []*ServiceStatus     `json:"service_statuses,omitempty"`
	GroupAggregates  map[string]*GroupAggregate `json:"group_aggregates,omitempty"`
}

type GroupAggregate struct {
	GroupName      string       `json:"group_name"`
	OverallStatus  HealthStatus `json:"overall_status"`
	TotalServices  int          `json:"total_services"`
	HealthyCount   int          `json:"healthy_count"`
	UnhealthyCount int          `json:"unhealthy_count"`
}

type ServiceConfig struct {
	Name               string        `json:"name" yaml:"name"`
	Group              string        `json:"group,omitempty" yaml:"group,omitempty"`
	ProbeType          ProbeType     `json:"probe_type" yaml:"probe_type"`
	Target             string        `json:"target" yaml:"target"`
	Interval           time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout            time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	FailureThreshold   int           `json:"failure_threshold,omitempty" yaml:"failure_threshold,omitempty"`
	RecoveryThreshold  int           `json:"recovery_threshold,omitempty" yaml:"recovery_threshold,omitempty"`
	ExpectedStatusCodes []int         `json:"expected_status_codes,omitempty" yaml:"expected_status_codes,omitempty"`
}

type AddServiceRequest struct {
	ServiceConfig
}

type AddServiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RemoveServiceRequest struct {
	Name string `json:"name"`
}

type RemoveServiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ServerHealthResponse struct {
	Version       string         `json:"version"`
	Status        HealthStatus   `json:"status"`
	Timestamp     time.Time      `json:"timestamp"`
	Aggregate     AggregateStatus `json:"aggregate"`
}
