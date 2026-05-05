package common

import (
	"time"
)

type MessageType string

const (
	MsgTypeSubscribe         MessageType = "subscribe"
	MsgTypeUnsubscribe       MessageType = "unsubscribe"
	MsgTypeMetricUpdate      MessageType = "metric_update"
	MsgTypeAlert             MessageType = "alert"
	MsgTypeReconnect         MessageType = "reconnect"
	MsgTypeReconnectResponse MessageType = "reconnect_response"
	MsgTypeError             MessageType = "error"
	MsgTypeHeartbeat         MessageType = "heartbeat"
)

type MetricMetadata struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
	Formula     string `json:"formula"`
}

type AlertThreshold struct {
	MaxValue *float64 `json:"max_value,omitempty"`
	MinValue *float64 `json:"min_value,omitempty"`
}

type Metric struct {
	Key          string          `json:"key"`
	Metadata     MetricMetadata  `json:"metadata"`
	Value        float64         `json:"value"`
	Timestamp    time.Time       `json:"timestamp"`
	IsExpired    bool            `json:"is_expired"`
	AlertThreshold *AlertThreshold `json:"alert_threshold,omitempty"`
}

type MetricDataPoint struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type TrendDataPoint struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Period    string    `json:"period"`
}

type ComparisonData struct {
	CurrentValue  float64 `json:"current_value"`
	CompareValue  float64 `json:"compare_value"`
	YoYPercent    float64 `json:"yoy_percent"`
	MoMPercent    float64 `json:"mom_percent"`
}

type Alert struct {
	ID          string    `json:"id"`
	MetricKey   string    `json:"metric_key"`
	Message     string    `json:"message"`
	Threshold   float64   `json:"threshold"`
	ActualValue float64   `json:"actual_value"`
	AlertType   string    `json:"alert_type"`
	Timestamp   time.Time `json:"timestamp"`
}

type SubscribeRequest struct {
	ClientID     string   `json:"client_id"`
	MetricKeys   []string `json:"metric_keys"`
	RefreshRate  int      `json:"refresh_rate"`
	LastSeenTime *time.Time `json:"last_seen_time,omitempty"`
}

type SubscribeResponse struct {
	ClientID     string       `json:"client_id"`
	Subscribed   []string     `json:"subscribed"`
	QueuePosition *int         `json:"queue_position,omitempty"`
	Error        string       `json:"error,omitempty"`
}

type UnsubscribeRequest struct {
	ClientID   string   `json:"client_id"`
	MetricKeys []string `json:"metric_keys"`
}

type ReconnectRequest struct {
	ClientID       string    `json:"client_id"`
	DisconnectTime time.Time `json:"disconnect_time"`
}

type ReconnectResponse struct {
	ClientID     string          `json:"client_id"`
	MissedData   []Metric        `json:"missed_data,omitempty"`
	LatestOnly   bool            `json:"latest_only"`
	Subscriptions []string        `json:"subscriptions"`
}

type ClientInfo struct {
	ClientID        string    `json:"client_id"`
	ConnectedAt     time.Time `json:"connected_at"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
	SubscribedMetrics []string `json:"subscribed_metrics"`
	RefreshRate     int       `json:"refresh_rate"`
}

type DashboardLayout struct {
	ClientID string   `json:"client_id"`
	Metrics  []string `json:"metrics"`
	Order    []int    `json:"order"`
}

type DelayConfig struct {
	WarningThresholdMS int `json:"warning_threshold_ms"`
	CriticalThresholdMS int `json:"critical_threshold_ms"`
}

type WebSocketMessage struct {
	Type    MessageType `json:"type"`
	Payload []byte      `json:"payload"`
}

type MetricLatestValue struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Timestamp  time.Time `json:"timestamp"`
	IsExpired  bool      `json:"is_expired"`
}

type ServerStatus struct {
	OnlineClients    int                     `json:"online_clients"`
	MaxClients       int                     `json:"max_clients"`
	QueueSize        int                     `json:"queue_size"`
	MetricsCount     int                     `json:"metrics_count"`
	LatestValues     []MetricLatestValue     `json:"latest_values,omitempty"`
	Clients          []ClientInfo            `json:"clients,omitempty"`
}
