package model

import "time"

type PingRequest struct {
	Target   string        `json:"target"`
	Count    int           `json:"count"`
	Interval time.Duration `json:"interval"`
	TTL      int           `json:"ttl"`
}

type PingResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type TaskStatus struct {
	ID              string        `json:"id"`
	Target          string        `json:"target"`
	Count           int           `json:"count"`
	Interval        time.Duration `json:"interval"`
	TTL             int           `json:"ttl"`
	PacketsSent     int           `json:"packets_sent"`
	PacketsReceived int           `json:"packets_received"`
	PacketLoss      float64       `json:"packet_loss"`
	MinRTT          int64         `json:"min_rtt_ns"`
	MaxRTT          int64         `json:"max_rtt_ns"`
	AvgRTT          int64         `json:"avg_rtt_ns"`
	StdDevRTT       int64         `json:"stddev_rtt_ns"`
	IsRunning       bool          `json:"is_running"`
	CreatedAt       time.Time     `json:"created_at"`
	FinishedAt      *time.Time    `json:"finished_at,omitempty"`
	TimeExceeded    []*TimeExceededInfo `json:"time_exceeded,omitempty"`
}

type TimeExceededInfo struct {
	RouterIP   string `json:"router_ip"`
	TargetIP   string `json:"target_ip"`
	Sequence   uint16 `json:"sequence"`
}

type StopResponse struct {
	Success bool `json:"success"`
}
