// Package ratelimiter provides rate limiting primitives for microservice gateways
// and API open platforms. It supports multiple rate limiting strategies including
// fixed window, sliding window, token bucket, and leaky bucket.
package ratelimiter

import (
	"time"
)

// LimitMode represents the rate limiting mode.
type LimitMode int

const (
	// ModeFixedWindow is the fixed window rate limiting mode.
	// It counts requests in fixed time intervals (e.g., per second, per minute).
	// Note: This mode may allow doubled traffic at window boundaries.
	ModeFixedWindow LimitMode = iota

	// ModeSlidingWindow is the sliding window rate limiting mode.
	// It maintains a sliding window of request timestamps to provide
	// more accurate rate limiting than fixed window.
	ModeSlidingWindow

	// ModeSmoothWindow is the smooth window rate limiting mode.
	// It divides a window into multiple small buckets to smooth out
	// boundary bursts that occur in fixed window mode.
	ModeSmoothWindow

	// ModeTokenBucket is the token bucket rate limiting mode.
	// Tokens are added to the bucket at a fixed rate. Requests consume
	// tokens and are rejected if the bucket is empty.
	ModeTokenBucket

	// ModeLeakyBucket is the leaky bucket rate limiting mode.
	// Requests are queued and processed at a constant rate, regardless
	// of incoming request burstiness.
	ModeLeakyBucket
)

// Result represents the result of a rate limit check.
type Result struct {
	// Allowed indicates whether the request is allowed to proceed.
	Allowed bool

	// Limited indicates whether the request was rate limited.
	Limited bool

	// Remaining is the number of remaining requests/tokens allowed in the current window.
	Remaining int64

	// WaitTime is the suggested time to wait before retrying.
	WaitTime time.Duration

	// ResetAt is the time when the current window resets.
	ResetAt time.Time
}

// Stats represents statistics for a rate limiter.
type Stats struct {
	// Key is the identifier of the rate limiter.
	Key string

	// Mode is the rate limiting mode.
	Mode LimitMode

	// Used is the number of requests/tokens used in the current window.
	Used int64

	// Limit is the maximum allowed requests/tokens per window.
	Limit int64

	// Remaining is the number of remaining requests/tokens.
	Remaining int64

	// ResetAt is the time when the current window resets.
	ResetAt time.Time
}

// Limiter is the interface that all rate limiters must implement.
type Limiter interface {
	// Allow checks if a request with the given key is allowed to proceed.
	// It returns a Result indicating whether the request is allowed.
	Allow(key string) Result

	// AllowN checks if n requests with the given key are allowed to proceed.
	// It returns a Result indicating whether the requests are allowed.
	AllowN(key string, n int64) Result

	// Stats returns statistics for the given key.
	Stats(key string) Stats

	// UpdateConfig updates the rate limiter configuration dynamically.
	// The update should take effect immediately without discarding existing counts.
	UpdateConfig(config Config) error

	// GetConfig returns the current configuration of the rate limiter.
	GetConfig() Config

	// Reset resets the rate limiter for the given key.
	Reset(key string)

	// ResetAll resets all rate limiters.
	ResetAll()
}

// Config represents the configuration for a rate limiter.
type Config struct {
	// Mode specifies the rate limiting mode.
	Mode LimitMode

	// Limit is the maximum number of requests/tokens allowed per window.
	Limit int64

	// Window is the duration of the rate limiting window.
	Window time.Duration

	// BucketCount is the number of small buckets for smooth window mode.
	// It must be at least 2.
	BucketCount int

	// Burst is the maximum burst size for token bucket mode.
	// This is the maximum number of tokens that can be accumulated.
	Burst int64

	// Rate is the token generation rate per second for token bucket mode,
	// or the processing rate per second for leaky bucket mode.
	Rate float64

	// Capacity is the maximum queue size for leaky bucket mode.
	Capacity int64
}

// DefaultConfig returns a default configuration for rate limiters.
func DefaultConfig() Config {
	return Config{
		Mode:        ModeFixedWindow,
		Limit:       100,
		Window:      time.Second,
		BucketCount: 10,
		Burst:       100,
		Rate:        100.0,
		Capacity:    100,
	}
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if c.Limit <= 0 {
		return ErrInvalidLimit
	}
	if c.Window <= 0 {
		return ErrInvalidWindow
	}
	if c.Mode == ModeSmoothWindow && c.BucketCount < 2 {
		return ErrInvalidBucketCount
	}
	if c.Mode == ModeTokenBucket {
		if c.Rate <= 0 {
			return ErrInvalidRate
		}
		if c.Burst <= 0 {
			return ErrInvalidBurst
		}
	}
	if c.Mode == ModeLeakyBucket {
		if c.Rate <= 0 {
			return ErrInvalidRate
		}
		if c.Capacity <= 0 {
			return ErrInvalidCapacity
		}
	}
	return nil
}
