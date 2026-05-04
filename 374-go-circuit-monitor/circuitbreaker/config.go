package circuitbreaker

import "time"

const (
	DefaultWindowSize          = 10
	DefaultFailureThreshold    = 6
	DefaultOpenTimeout         = 30 * time.Second
	DefaultHalfOpenMaxRequests = 3
	DefaultSuccessThreshold    = 10
)

type Config struct {
	WindowSize          int
	FailureThreshold    int
	OpenTimeout         time.Duration
	HalfOpenMaxRequests int
	SuccessThreshold    int
}

func DefaultConfig() Config {
	return Config{
		WindowSize:          DefaultWindowSize,
		FailureThreshold:    DefaultFailureThreshold,
		OpenTimeout:         DefaultOpenTimeout,
		HalfOpenMaxRequests: DefaultHalfOpenMaxRequests,
		SuccessThreshold:    DefaultSuccessThreshold,
	}
}

func (c Config) Validate() error {
	if c.WindowSize <= 0 {
		return &InvalidConfigError{Field: "WindowSize", Reason: "must be greater than 0"}
	}
	if c.FailureThreshold <= 0 || c.FailureThreshold > c.WindowSize {
		return &InvalidConfigError{Field: "FailureThreshold", Reason: "must be between 1 and WindowSize"}
	}
	if c.OpenTimeout <= 0 {
		return &InvalidConfigError{Field: "OpenTimeout", Reason: "must be greater than 0"}
	}
	if c.HalfOpenMaxRequests <= 0 {
		return &InvalidConfigError{Field: "HalfOpenMaxRequests", Reason: "must be greater than 0"}
	}
	if c.SuccessThreshold <= 0 {
		return &InvalidConfigError{Field: "SuccessThreshold", Reason: "must be greater than 0"}
	}
	return nil
}

type InvalidConfigError struct {
	Field  string
	Reason string
}

func (e *InvalidConfigError) Error() string {
	return "invalid config: " + e.Field + " - " + e.Reason
}
