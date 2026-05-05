package ratelimiter

import (
	"errors"
)

// NewLimiter creates a new rate limiter based on the configuration mode.
// It returns the appropriate Limiter implementation for the specified mode.
func NewLimiter(config Config) (Limiter, error) {
	switch config.Mode {
	case ModeFixedWindow:
		return NewFixedWindowLimiter(config)
	case ModeSlidingWindow:
		return NewSlidingWindowLimiter(config)
	case ModeSmoothWindow:
		return NewSmoothWindowLimiter(config)
	case ModeTokenBucket:
		return NewTokenBucketLimiter(config)
	case ModeLeakyBucket:
		return NewLeakyBucketLimiter(config)
	default:
		return nil, errors.New("unknown rate limit mode")
	}
}

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
