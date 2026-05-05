package ratelimiter

import (
	"time"
)

// CompositeLimiter combines multiple limiters into a single limiter.
// All limiters must allow the request for it to be allowed.
// If any limiter rejects the request, the request is rejected.
// This can be used to apply multiple constraints (e.g., per-user and per-API limits).
type CompositeLimiter struct {
	limiters []Limiter
}

// CompositeResult is returned by CompositeLimiter to indicate which limiter rejected the request.
type CompositeResult struct {
	Result

	// RejectingLimiterIndex is the index of the limiter that rejected the request.
	// If the request was allowed, this is -1.
	RejectingLimiterIndex int

	// IndividualResults contains the results from each individual limiter.
	IndividualResults []Result
}

// NewCompositeLimiter creates a new CompositeLimiter with the given limiters.
func NewCompositeLimiter(limiters ...Limiter) *CompositeLimiter {
	return &CompositeLimiter{
		limiters: limiters,
	}
}

// AddLimiter adds a limiter to the composite.
func (c *CompositeLimiter) AddLimiter(limiter Limiter) {
	c.limiters = append(c.limiters, limiter)
}

// Limiters returns the list of limiters in the composite.
func (c *CompositeLimiter) Limiters() []Limiter {
	return c.limiters
}

// Allow checks if a request is allowed by all limiters.
// It returns a CompositeResult with detailed information.
func (c *CompositeLimiter) Allow(key string) Result {
	result := c.AllowN(key, 1)
	return result
}

// AllowN checks if n requests are allowed by all limiters.
// It returns a CompositeResult with detailed information.
func (c *CompositeLimiter) AllowN(key string, n int64) Result {
	results := make([]Result, len(c.limiters))
	allowed := true
	maxWaitTime := time.Duration(0)
	earliestReset := time.Now().Add(24 * time.Hour)
	minRemaining := int64(-1)

	for i, limiter := range c.limiters {
		results[i] = limiter.AllowN(key, n)
		if !results[i].Allowed {
			allowed = false
			if results[i].WaitTime > maxWaitTime {
				maxWaitTime = results[i].WaitTime
			}
		}
		if results[i].ResetAt.Before(earliestReset) && !results[i].ResetAt.IsZero() {
			earliestReset = results[i].ResetAt
		}
		if minRemaining < 0 || results[i].Remaining < minRemaining {
			minRemaining = results[i].Remaining
		}
	}

	return Result{
		Allowed:   allowed,
		Limited:   !allowed,
		Remaining: minRemaining,
		WaitTime:  maxWaitTime,
		ResetAt:   earliestReset,
	}
}

// Stats returns combined statistics from all limiters.
// It returns the stats with the smallest remaining count.
func (c *CompositeLimiter) Stats(key string) Stats {
	var minStats Stats
	found := false

	for _, limiter := range c.limiters {
		stats := limiter.Stats(key)
		if !found || stats.Remaining < minStats.Remaining {
			minStats = stats
			found = true
		}
	}

	return minStats
}

// UpdateConfig is not supported for CompositeLimiter.
// Use UpdateConfig on individual limiters instead.
func (c *CompositeLimiter) UpdateConfig(config Config) error {
	return ErrUnsupportedMode
}

// GetConfig is not supported for CompositeLimiter.
// Use GetConfig on individual limiters instead.
func (c *CompositeLimiter) GetConfig() Config {
	if len(c.limiters) > 0 {
		return c.limiters[0].GetConfig()
	}
	return DefaultConfig()
}

// Reset resets all limiters for the given key.
func (c *CompositeLimiter) Reset(key string) {
	for _, limiter := range c.limiters {
		limiter.Reset(key)
	}
}

// ResetAll resets all limiters for all keys.
func (c *CompositeLimiter) ResetAll() {
	for _, limiter := range c.limiters {
		limiter.ResetAll()
	}
}
