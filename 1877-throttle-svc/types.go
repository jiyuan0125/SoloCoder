package main

import (
	"sync"
	"time"
)

type Dimension string

const (
	DimensionIP     Dimension = "ip"
	DimensionAPIKey Dimension = "apikey"
)

type Mode string

const (
	ModeFixed   Mode = "fixed"
	ModeSliding Mode = "sliding"
)

type Rule struct {
	ID           string    `json:"id"`
	Dimension    Dimension `json:"dimension"`
	Mode         Mode      `json:"mode"`
	WindowSeconds int      `json:"window_seconds"`
	MaxRequests  int       `json:"max_requests"`
	CreatedAt    time.Time `json:"created_at"`
}

type RuleRequest struct {
	Dimension    Dimension `json:"dimension"`
	Mode         Mode      `json:"mode"`
	WindowSeconds int      `json:"window_seconds"`
	MaxRequests  int       `json:"max_requests"`
}

type Stats struct {
	RuleID          string    `json:"rule_id"`
	Dimension       Dimension `json:"dimension"`
	TriggerCount    int64     `json:"trigger_count"`
	LastTriggerTime time.Time `json:"last_trigger_time"`
}

type StatsKey struct {
	RuleID string
	Key    string
}

type Counter struct {
	FixedWindows map[int64]int64
	SlidingLogs  []time.Time
}

type RuleStore struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

type StatsStore struct {
	mu      sync.RWMutex
	stats   map[StatsKey]*Stats
	counter map[StatsKey]*Counter
}
