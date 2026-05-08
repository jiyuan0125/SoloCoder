package limiter

import (
	"sync"
	"time"
)

type LimiterMode string

const (
	ModeGrid    LimiterMode = "grid"
	ModePrecise LimiterMode = "precise"
)

type LimitRule struct {
	Name       string
	Limit      int
	Window     time.Duration
	GridSize   time.Duration
	Mode       LimiterMode
	Priority   int
}

type Result struct {
	Allowed    bool
	Remaining  int
	Limit      int
	RetryAfter time.Duration
	WindowEnd  time.Time
}

type Limiter interface {
	Allow() Result
	AllowN(n int) Result
	Reset()
	Stats() map[string]interface{}
	GetRule() LimitRule
}

type baseLimiter struct {
	mu   sync.RWMutex
	rule LimitRule
}

func (b *baseLimiter) GetRule() LimitRule {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.rule
}

func NewLimiter(rule LimitRule) Limiter {
	if rule.GridSize <= 0 {
		rule.GridSize = time.Second
	}
	if rule.Window <= 0 {
		rule.Window = time.Minute
	}
	if rule.Limit <= 0 {
		rule.Limit = 100
	}
	if rule.Mode == "" {
		rule.Mode = ModeGrid
	}

	switch rule.Mode {
	case ModePrecise:
		return newPreciseLimiter(rule)
	default:
		return newGridLimiter(rule)
	}
}
