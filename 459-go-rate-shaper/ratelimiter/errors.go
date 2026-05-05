package ratelimiter

import "errors"

var (
	// ErrInvalidLimit is returned when the limit is not positive.
	ErrInvalidLimit = errors.New("limit must be positive")

	// ErrInvalidWindow is returned when the window is not positive.
	ErrInvalidWindow = errors.New("window must be positive")

	// ErrInvalidBucketCount is returned when the bucket count is less than 2 for smooth window mode.
	ErrInvalidBucketCount = errors.New("bucket count must be at least 2 for smooth window mode")

	// ErrInvalidRate is returned when the rate is not positive.
	ErrInvalidRate = errors.New("rate must be positive")

	// ErrInvalidBurst is returned when the burst is not positive.
	ErrInvalidBurst = errors.New("burst must be positive")

	// ErrInvalidCapacity is returned when the capacity is not positive.
	ErrInvalidCapacity = errors.New("capacity must be positive")

	// ErrModeMismatch is returned when trying to update config with a different mode.
	ErrModeMismatch = errors.New("cannot change mode of existing limiter")

	// ErrUnsupportedMode is returned when the mode is not supported.
	ErrUnsupportedMode = errors.New("unsupported mode")
)
