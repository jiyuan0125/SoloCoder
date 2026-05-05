// Package protocol defines the communication protocol between the rate limiter
// server and client. It contains request and response structures for HTTP API.
package protocol

import (
	"time"
)

// LimitMode represents the rate limiting mode.
type LimitMode int

const (
	// ModeFixedWindow is the fixed window rate limiting mode.
	ModeFixedWindow LimitMode = iota

	// ModeSlidingWindow is the sliding window rate limiting mode.
	ModeSlidingWindow

	// ModeSmoothWindow is the smooth window rate limiting mode.
	ModeSmoothWindow

	// ModeTokenBucket is the token bucket rate limiting mode.
	ModeTokenBucket

	// ModeLeakyBucket is the leaky bucket rate limiting mode.
	ModeLeakyBucket
)

// ModeToString converts a LimitMode to its string representation.
func ModeToString(mode LimitMode) string {
	switch mode {
	case ModeFixedWindow:
		return "fixed_window"
	case ModeSlidingWindow:
		return "sliding_window"
	case ModeSmoothWindow:
		return "smooth_window"
	case ModeTokenBucket:
		return "token_bucket"
	case ModeLeakyBucket:
		return "leaky_bucket"
	default:
		return "unknown"
	}
}

// StringToMode converts a string to a LimitMode.
func StringToMode(s string) LimitMode {
	switch s {
	case "fixed_window":
		return ModeFixedWindow
	case "sliding_window":
		return ModeSlidingWindow
	case "smooth_window":
		return ModeSmoothWindow
	case "token_bucket":
		return ModeTokenBucket
	case "leaky_bucket":
		return ModeLeakyBucket
	default:
		return ModeFixedWindow
	}
}

// AllowRequest is the request structure for checking if a request is allowed.
type AllowRequest struct {
	// Key is the identifier for rate limiting (e.g., user ID, IP address).
	Key string `json:"key"`

	// Count is the number of requests to check (default: 1).
	Count int64 `json:"count,omitempty"`
}

// AllowResponse is the response structure for an allow request.
type AllowResponse struct {
	// Allowed indicates whether the request is allowed to proceed.
	Allowed bool `json:"allowed"`

	// Limited indicates whether the request was rate limited.
	Limited bool `json:"limited"`

	// Remaining is the number of remaining requests/tokens allowed.
	Remaining int64 `json:"remaining"`

	// WaitTimeMs is the suggested wait time in milliseconds before retrying.
	WaitTimeMs int64 `json:"wait_time_ms"`

	// ResetAt is the time when the current window resets (RFC3339 format).
	ResetAt string `json:"reset_at"`

	// Error contains the error message if any.
	Error string `json:"error,omitempty"`
}

// StatsRequest is the request structure for getting rate limiter statistics.
type StatsRequest struct {
	// Key is the identifier for rate limiting.
	Key string `json:"key"`
}

// StatsResponse is the response structure for statistics.
type StatsResponse struct {
	// Key is the identifier of the rate limiter.
	Key string `json:"key"`

	// Mode is the rate limiting mode.
	Mode string `json:"mode"`

	// Used is the number of requests/tokens used.
	Used int64 `json:"used"`

	// Limit is the maximum allowed requests/tokens.
	Limit int64 `json:"limit"`

	// Remaining is the number of remaining requests/tokens.
	Remaining int64 `json:"remaining"`

	// ResetAt is the time when the current window resets (RFC3339 format).
	ResetAt string `json:"reset_at"`

	// Error contains the error message if any.
	Error string `json:"error,omitempty"`
}

// ConfigRequest is the request structure for updating rate limiter configuration.
type ConfigRequest struct {
	// Mode specifies the rate limiting mode.
	Mode string `json:"mode"`

	// Limit is the maximum number of requests/tokens allowed per window.
	Limit int64 `json:"limit"`

	// WindowMs is the duration of the rate limiting window in milliseconds.
	WindowMs int64 `json:"window_ms"`

	// BucketCount is the number of small buckets for smooth window mode.
	BucketCount int `json:"bucket_count,omitempty"`

	// Burst is the maximum burst size for token bucket mode.
	Burst int64 `json:"burst,omitempty"`

	// Rate is the token generation rate per second for token bucket mode,
	// or the processing rate per second for leaky bucket mode.
	Rate float64 `json:"rate,omitempty"`

	// Capacity is the maximum queue size for leaky bucket mode.
	Capacity int64 `json:"capacity,omitempty"`
}

// ConfigResponse is the response structure for configuration updates.
type ConfigResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// WhitelistRequest is the request structure for whitelist operations.
type WhitelistRequest struct {
	// Key is the identifier to add/remove from the whitelist.
	Key string `json:"key"`
}

// WhitelistResponse is the response structure for whitelist operations.
type WhitelistResponse struct {
	Success bool     `json:"success"`
	Keys    []string `json:"keys,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// ResetRequest is the request structure for reset operations.
type ResetRequest struct {
	// Key is the identifier to reset. If empty, reset all.
	Key string `json:"key,omitempty"`
}

// ResetResponse is the response structure for reset operations.
type ResetResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ResultToAllowResponse converts a ratelimiter.Result to AllowResponse.
func ResultToAllowResponse(result AllowResult) AllowResponse {
	resetAt := ""
	if !result.ResetAt.IsZero() {
		resetAt = result.ResetAt.Format(time.RFC3339)
	}

	return AllowResponse{
		Allowed:    result.Allowed,
		Limited:    result.Limited,
		Remaining:  result.Remaining,
		WaitTimeMs: result.WaitTime.Milliseconds(),
		ResetAt:    resetAt,
	}
}

// AllowResult is a simplified version of ratelimiter.Result for protocol use.
type AllowResult struct {
	Allowed   bool
	Limited   bool
	Remaining int64
	WaitTime  time.Duration
	ResetAt   time.Time
}
