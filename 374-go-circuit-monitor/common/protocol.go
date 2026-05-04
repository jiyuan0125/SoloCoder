package common

import "time"

type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half-open"
)

type CircuitInfo struct {
	Name           string       `json:"name"`
	State          CircuitState `json:"state"`
	TotalRequests  int64        `json:"total_requests"`
	SuccessCount   int64        `json:"success_count"`
	FailureCount   int64        `json:"failure_count"`
	RecentFailures int          `json:"recent_failures"`
	RecentSuccesses int         `json:"recent_successes"`
	OpenAt         *time.Time   `json:"open_at,omitempty"`
	LastStateChange *time.Time `json:"last_state_change,omitempty"`
}

type CircuitConfig struct {
	WindowSize          int           `json:"window_size"`
	FailureThreshold    int           `json:"failure_threshold"`
	OpenTimeout         time.Duration `json:"open_timeout"`
	HalfOpenMaxRequests int           `json:"half_open_max_requests"`
	SuccessThreshold    int           `json:"success_threshold"`
}

type ListCircuitsResponse struct {
	Circuits []string `json:"circuits"`
}

type GetCircuitResponse struct {
	CircuitInfo CircuitInfo `json:"circuit_info"`
}

type CreateCircuitRequest struct {
	Name   string         `json:"name"`
	Config *CircuitConfig `json:"config,omitempty"`
}

type CreateCircuitResponse struct {
	Success bool `json:"success"`
}

type ResetCircuitResponse struct {
	Success bool `json:"success"`
}

type ForceStateRequest struct {
	State CircuitState `json:"state"`
}

type ForceStateResponse struct {
	Success bool `json:"success"`
}

type GetConfigResponse struct {
	Config CircuitConfig `json:"config"`
}

type UpdateConfigRequest struct {
	Config CircuitConfig `json:"config"`
}

type UpdateConfigResponse struct {
	Success bool `json:"success"`
}

type SaveStateResponse struct {
	Success bool `json:"success"`
}

type LoadStateResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
