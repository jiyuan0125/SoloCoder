package protocol

import "encoding/json"

type MessageType string

const (
	MessageTypeAddWatch        MessageType = "ADD_WATCH"
	MessageTypeRemoveWatch     MessageType = "REMOVE_WATCH"
	MessageTypeListWatches     MessageType = "LIST_WATCHES"
	MessageTypeGetStatus       MessageType = "GET_STATUS"
	MessageTypeGetForwardLogs  MessageType = "GET_FORWARD_LOGS"
	MessageTypeResponse        MessageType = "RESPONSE"
)

type Request struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Response struct {
	Type    MessageType     `json:"type"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type AddWatchRequest struct {
	FilePath string `json:"file_path"`
}

type RemoveWatchRequest struct {
	FilePath string `json:"file_path"`
}

type ListWatchesResponse struct {
	Files []WatchedFileInfo `json:"files"`
}

type WatchedFileInfo struct {
	FilePath     string `json:"file_path"`
	CurrentOffset int64 `json:"current_offset"`
	AlertsParsed int64  `json:"alerts_parsed"`
}

type StatusResponse struct {
	ServerStatus string             `json:"server_status"`
	Uptime       string             `json:"uptime"`
	TotalAlerts  int64              `json:"total_alerts"`
	ChannelStats map[string]int64   `json:"channel_stats"`
}

type ForwardLog struct {
	Timestamp   string `json:"timestamp"`
	Level       string `json:"level"`
	Message     string `json:"message"`
	Channel     string `json:"channel"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

type ForwardLogsResponse struct {
	Logs []ForwardLog `json:"logs"`
}
