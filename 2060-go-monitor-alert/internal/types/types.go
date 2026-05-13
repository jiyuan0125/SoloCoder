package types

import (
	"time"
)

type MetricType string

const (
	MetricCPU    MetricType = "cpu"
	MetricMemory MetricType = "memory"
	MetricDisk   MetricType = "disk"
	MetricNetwork MetricType = "network"
)

type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

type Condition struct {
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

type AlertRule struct {
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	Condition      string          `json:"condition"`
	ParsedConditions []Condition   `json:"-"`
	LogicalOp      string          `json:"logical_op"`
	Duration       time.Duration   `json:"duration"`
	Severity       Severity        `json:"severity"`
	ReminderInterval time.Duration `json:"reminder_interval"`
	Channels       []string        `json:"channels"`
}

type Alert struct {
	RuleName       string        `json:"rule_name"`
	Severity       Severity      `json:"severity"`
	Status         AlertStatus   `json:"status"`
	StartedAt      time.Time     `json:"started_at"`
	ResolvedAt     *time.Time    `json:"resolved_at,omitempty"`
	LastNotifiedAt time.Time     `json:"last_notified_at"`
	Message        string        `json:"message"`
}

type ChannelType string

const (
	ChannelStdout  ChannelType = "stdout"
	ChannelEmail   ChannelType = "email"
	ChannelWebhook ChannelType = "webhook"
)

type NotificationChannel struct {
	Name     string      `json:"name"`
	Type     ChannelType `json:"type"`
	Config   interface{} `json:"config"`
}

type EmailConfig struct {
	SMTPHost string   `json:"smtp_host"`
	SMTPPort int      `json:"smtp_port"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
}

type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout time.Duration     `json:"timeout"`
}

type StatsInfo struct {
	TotalRecords   int64            `json:"total_records"`
	MetricsCount   map[string]int64 `json:"metrics_count"`
	AlertsCount    int64            `json:"alerts_count"`
	LastUpdated    time.Time        `json:"last_updated"`
}
