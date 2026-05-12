package models

import "time"

type ComponentStatus string

const (
	StatusUnknown   ComponentStatus = "UNKNOWN"
	StatusHealthy   ComponentStatus = "HEALTHY"
	StatusUnhealthy ComponentStatus = "UNHEALTHY"
)

type CheckType string

const (
	CheckTypeHTTP CheckType = "HTTP"
	CheckTypeTCP  CheckType = "TCP"
)

type CheckResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Status    ComponentStatus `json:"status"`
	Message   string          `json:"message,omitempty"`
}

type Component struct {
	Name             string            `json:"name"`
	CheckType        CheckType         `json:"checkType"`
	Address          string            `json:"address"`
	Timeout          time.Duration     `json:"timeout"`
	Interval         time.Duration     `json:"interval"`
	Status           ComponentStatus   `json:"status"`
	RegisterTime     time.Time         `json:"registerTime"`
	Deregistered     bool              `json:"-"`
	DeregisterTime   time.Time         `json:"-"`
	CheckResults     []CheckResult     `json:"checkResults"`
	LastCheckTime    time.Time         `json:"lastCheckTime,omitempty"`
	ConsecutiveGood  int               `json:"-"`
	ConsecutiveBad   int               `json:"-"`
	StopChan         chan struct{}     `json:"-"`
}

type RegisterRequest struct {
	Name      string        `json:"name" binding:"required"`
	CheckType CheckType     `json:"checkType" binding:"required"`
	Address   string        `json:"address" binding:"required"`
	Timeout   time.Duration `json:"timeout"`
	Interval  time.Duration `json:"interval"`
}

type ComponentDetail struct {
	Name         string        `json:"name"`
	CheckType    CheckType     `json:"checkType"`
	Address      string        `json:"address"`
	Timeout      time.Duration `json:"timeout"`
	Interval     time.Duration `json:"interval"`
	Status       ComponentStatus `json:"status"`
	RegisterTime time.Time     `json:"registerTime"`
	LastCheckTime time.Time    `json:"lastCheckTime,omitempty"`
	CheckResults []CheckResult `json:"checkResults"`
}

type HealthSummary struct {
	Status     ComponentStatus   `json:"status"`
	Components []*ComponentDetail `json:"components"`
}
