package common

import "time"

type MessageInfo struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Size      int       `json:"size"`
}

type StatsResponse struct {
	TotalMessages   int64  `json:"total_messages"`
	RecentMessages  []MessageInfo `json:"recent_messages"`
	HasTruncation   bool   `json:"has_truncation"`
	TruncationCount int64  `json:"truncation_count"`
	FifoPath        string `json:"fifo_path"`
	ServerUptime    string `json:"server_uptime"`
}

type WriteRequest struct {
	Content string `json:"content"`
}

type WriteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type BatchWriteRequest struct {
	Messages []string `json:"messages"`
}

type BatchWriteResponse struct {
	Success   bool   `json:"success"`
	Count     int    `json:"count"`
	Failed    int    `json:"failed"`
	Message   string `json:"message,omitempty"`
}

type FlushTestRequest struct {
	Count       int    `json:"count"`
	Concurrency int    `json:"concurrency"`
}

type FlushTestResponse struct {
	Success     bool   `json:"success"`
	SentCount   int    `json:"sent_count"`
	FailedCount int    `json:"failed_count"`
	Duration    string `json:"duration"`
	Message     string `json:"message,omitempty"`
}
