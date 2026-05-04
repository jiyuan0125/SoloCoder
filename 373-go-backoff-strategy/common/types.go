package common

import (
	"errors"
	"fmt"
	"time"
)

type StrategyType string

const (
	StrategyFixedInterval     StrategyType = "fixed"
	StrategyLinear            StrategyType = "linear"
	StrategyExponential       StrategyType = "exponential"
	StrategyExponentialJitter StrategyType = "exponential-jitter"
)

var ErrUnknownStrategyType = errors.New("unknown strategy type")

func (s StrategyType) Validate() error {
	switch s {
	case StrategyFixedInterval, StrategyLinear, StrategyExponential, StrategyExponentialJitter:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownStrategyType, s)
	}
}

type StrategyConfig struct {
	Type         StrategyType  `json:"type"`
	MaxRetries   int           `json:"max_retries"`
	MaxWait      time.Duration `json:"max_wait"`
	InitialWait  time.Duration `json:"initial_wait"`
	Increment    time.Duration `json:"increment,omitempty"`
	Multiplier   float64       `json:"multiplier,omitempty"`
	JitterFactor float64       `json:"jitter_factor,omitempty"`
}

type ExecuteRequest struct {
	TaskID         string         `json:"task_id"`
	StrategyConfig StrategyConfig `json:"strategy_config"`
	TaskType       string         `json:"task_type"`
	TaskPayload    map[string]any `json:"task_payload"`
}

type RetryEvent struct {
	Attempt int           `json:"attempt"`
	Delay   time.Duration `json:"delay"`
	Error   string        `json:"error"`
}

type ExecuteResponse struct {
	TaskID     string       `json:"task_id"`
	Success    bool         `json:"success"`
	Attempts   int          `json:"attempts"`
	RetryEvents []RetryEvent `json:"retry_events,omitempty"`
	Result     map[string]any `json:"result,omitempty"`
	Error      string       `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	TaskTypeMockSuccess    = "mock_success"
	TaskTypeMockFailure    = "mock_failure"
	TaskTypeMockNonRetryable = "mock_non_retryable"
	TaskTypeMockFlaky      = "mock_flaky"
)

func IsValidTaskType(taskType string) bool {
	switch taskType {
	case TaskTypeMockSuccess, TaskTypeMockFailure, TaskTypeMockNonRetryable, TaskTypeMockFlaky:
		return true
	default:
		return false
	}
}
