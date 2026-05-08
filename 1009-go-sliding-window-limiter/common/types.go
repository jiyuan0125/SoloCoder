package common

import (
	"time"
)

type LimitRuleDTO struct {
	Name     string `json:"name"`
	Limit    int    `json:"limit"`
	WindowMs int64  `json:"window_ms"`
	GridMs   int64  `json:"grid_ms"`
	Mode     string `json:"mode"`
	Priority int    `json:"priority"`
}

type AddRuleRequest struct {
	Endpoint string         `json:"endpoint"`
	Rule     LimitRuleDTO   `json:"rule"`
}

type RemoveRuleRequest struct {
	Endpoint string `json:"endpoint"`
	RuleName string `json:"rule_name"`
}

type RateLimitResponse struct {
	Allowed    bool   `json:"allowed"`
	Remaining  int    `json:"remaining"`
	Limit      int    `json:"limit"`
	RetryAfter int64  `json:"retry_after_ms"`
	RuleName   string `json:"rule_name,omitempty"`
	Message    string `json:"message,omitempty"`
}

type EndpointStatsDTO struct {
	Endpoint string                            `json:"endpoint"`
	Rules    []LimitRuleDTO                    `json:"rules"`
	Stats    map[string]map[string]interface{} `json:"stats"`
}

type AdminStatusResponse struct {
	Endpoints map[string]EndpointStatsDTO `json:"endpoints"`
	Timestamp time.Time                   `json:"timestamp"`
}

type TestRequest struct {
	Endpoint string `json:"endpoint"`
}

type TestResponse struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	RetryAfter int64  `json:"retry_after_ms,omitempty"`
	Message    string `json:"message,omitempty"`
}
