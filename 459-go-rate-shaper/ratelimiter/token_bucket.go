package ratelimiter

import (
	"sync"
	"time"
)

// TokenBucketLimiter implements rate limiting using a token bucket algorithm.
// Tokens are added to the bucket at a fixed rate. Requests consume tokens
// and are rejected if the bucket is empty. This allows for controlled bursts.
type TokenBucketLimiter struct {
	mu     sync.RWMutex
	config Config
	store  Storage
}

type tokenBucketState struct {
	tokens     float64
	lastUpdate time.Time
}

// NewTokenBucketLimiter creates a new TokenBucketLimiter with the given configuration.
func NewTokenBucketLimiter(config Config) (*TokenBucketLimiter, error) {
	config.Mode = ModeTokenBucket
	if err := config.Validate(); err != nil {
		return nil, err
	}

	store := NewMemoryStorage()

	return &TokenBucketLimiter{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a request is allowed.
func (l *TokenBucketLimiter) Allow(key string) Result {
	return l.AllowN(key, 1)
}

// AllowN checks if n requests are allowed.
func (l *TokenBucketLimiter) AllowN(key string, n int64) Result {
	l.mu.RLock()
	rate := l.config.Rate
	burst := l.config.Burst
	l.mu.RUnlock()

	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	stateAny, exists := l.store.Get(key)
	var state tokenBucketState

	if exists {
		state = stateAny.(tokenBucketState)
		elapsed := now.Sub(state.lastUpdate).Seconds()
		newTokens := elapsed * rate
		state.tokens += newTokens
		if state.tokens > float64(burst) {
			state.tokens = float64(burst)
		}
	} else {
		state = tokenBucketState{
			tokens:     float64(burst),
			lastUpdate: now,
		}
	}

	state.lastUpdate = now
	allowed := state.tokens >= float64(n)

	if allowed {
		state.tokens -= float64(n)
	}

	l.store.Set(key, state)

	remaining := int64(state.tokens)
	if remaining < 0 {
		remaining = 0
	}

	waitTime := time.Duration(0)
	resetAt := now

	if !allowed {
		tokensNeeded := float64(n) - state.tokens
		waitSeconds := tokensNeeded / rate
		waitTime = time.Duration(waitSeconds * float64(time.Second))
		resetAt = now.Add(waitTime)
	}

	return Result{
		Allowed:   allowed,
		Limited:   !allowed,
		Remaining: remaining,
		WaitTime:  waitTime,
		ResetAt:   resetAt,
	}
}

// Stats returns statistics for the given key.
func (l *TokenBucketLimiter) Stats(key string) Stats {
	l.mu.RLock()
	rate := l.config.Rate
	burst := l.config.Burst
	l.mu.RUnlock()

	now := time.Now()

	l.mu.RLock()
	defer l.mu.RUnlock()

	stateAny, exists := l.store.Get(key)
	var tokens float64 = float64(burst)

	if exists {
		state := stateAny.(tokenBucketState)
		elapsed := now.Sub(state.lastUpdate).Seconds()
		newTokens := elapsed * rate
		tokens = state.tokens + newTokens
		if tokens > float64(burst) {
			tokens = float64(burst)
		}
	}

	remaining := int64(tokens)
	if remaining < 0 {
		remaining = 0
	}

	return Stats{
		Key:       key,
		Mode:      ModeTokenBucket,
		Used:      burst - remaining,
		Limit:     burst,
		Remaining: remaining,
		ResetAt:   now.Add(time.Duration(float64(burst)/rate) * time.Second),
	}
}

// UpdateConfig updates the configuration.
func (l *TokenBucketLimiter) UpdateConfig(config Config) error {
	if config.Mode != ModeTokenBucket {
		return ErrModeMismatch
	}
	if err := config.Validate(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.config.Rate = config.Rate
	l.config.Burst = config.Burst

	return nil
}

// GetConfig returns the current configuration.
func (l *TokenBucketLimiter) GetConfig() Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// Reset resets the rate limiter for the given key.
func (l *TokenBucketLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Delete(key)
}

// ResetAll resets all rate limiters.
func (l *TokenBucketLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store.Clear()
}
