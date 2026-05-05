package protocol

import (
	"encoding/json"
	"errors"
	"time"
)

type MessageType string

const (
	MsgTypeSubmitTask     MessageType = "submit_task"
	MsgTypeTaskSubmitted  MessageType = "task_submitted"
	MsgTypeGetStatus      MessageType = "get_status"
	MsgTypeStatusResponse MessageType = "status_response"
	MsgTypeError          MessageType = "error"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type SubmitTaskRequest struct {
	InputFiles   []string `json:"input_files"`
	OutputFile   string   `json:"output_file"`
	TimeFormat   string   `json:"time_format,omitempty"`
	Resume       bool     `json:"resume"`
}

type SubmitTaskResponse struct {
	TaskID    string     `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Message   string     `json:"message,omitempty"`
}

type GetStatusRequest struct {
	TaskID string `json:"task_id"`
}

type GetStatusResponse struct {
	TaskID          string     `json:"task_id"`
	Status          TaskStatus `json:"status"`
	TotalFiles      int        `json:"total_files"`
	ProcessedLines  int64      `json:"processed_lines"`
	CurrentProgress float64    `json:"current_progress"`
	OutputFile      string     `json:"output_file,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

const DefaultTimeFormat = "2006-01-02 15:04:05"

func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	msg := &Message{Type: msgType}
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		msg.Payload = data
	}
	return msg, nil
}

func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func DecodeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func GenerateTaskID() string {
	return time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[int(time.Now().UnixNano())%len(letters)]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidPayload     = errors.New("invalid payload")
)
