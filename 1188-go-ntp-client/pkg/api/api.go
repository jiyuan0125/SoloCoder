package api

import "time"

type QueryRequest struct {
	Server  string        `json:"server"`
	Port    int           `json:"port,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

type QueryResponse struct {
	Server      string        `json:"server"`
	ServerTime  time.Time     `json:"server_time"`
	Offset      time.Duration `json:"offset_ns"`
	Delay       time.Duration `json:"delay_ns"`
	Stratum     uint8         `json:"stratum"`
	PollSec     float64       `json:"poll_sec"`
	PrecisionSec float64      `json:"precision_sec"`
}

type SyncRequest struct {
	Server  string        `json:"server"`
	Port    int           `json:"port,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

type SyncResponse struct {
	Server      string        `json:"server"`
	ServerTime  time.Time     `json:"server_time"`
	LocalTime   time.Time     `json:"local_time"`
	Offset      time.Duration `json:"offset_ns"`
	Delay       time.Duration `json:"delay_ns"`
	Adjusted    bool          `json:"adjusted"`
	Message     string        `json:"message"`
}

type HistoryEntry struct {
	Timestamp   time.Time     `json:"timestamp"`
	Server      string        `json:"server"`
	Offset      time.Duration `json:"offset_ns"`
	Delay       time.Duration `json:"delay_ns"`
	Stratum     uint8         `json:"stratum"`
}

type HistoryResponse struct {
	Entries []HistoryEntry `json:"entries"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
