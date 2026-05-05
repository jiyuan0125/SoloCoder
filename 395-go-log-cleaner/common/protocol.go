package common

import (
	"encoding/json"
	"time"
)

type RequestType string

const (
	RequestTypeSubmitTask  RequestType = "submit_task"
	RequestTypeGetHistory  RequestType = "get_history"
	RequestTypeGetTask     RequestType = "get_task"
	RequestTypeListTasks   RequestType = "list_tasks"
	RequestTypeHealthCheck RequestType = "health_check"
)

type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusError   ResponseStatus = "error"
	ResponseStatusPending ResponseStatus = "pending"
)

type CleanMode int

const (
	CleanModeTime CleanMode = 1 << iota
	CleanModeCapacity
)

type CleanTaskRequest struct {
	TargetDir       string        `json:"target_dir"`
	Mode            CleanMode     `json:"mode"`
	RetainDays      int           `json:"retain_days"`
	CapacityThreshold float64    `json:"capacity_threshold"`
	SafeWaterLevel  float64       `json:"safe_water_level"`
	FileExtensions  []string      `json:"file_extensions"`
	Recursive       bool          `json:"recursive"`
	DryRun          bool          `json:"dry_run"`
	Confirm         bool          `json:"confirm"`
	MaxDelete       int           `json:"max_delete"`
}

type FileInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	DiskUsage    int64     `json:"disk_usage"`
	ModTime      time.Time `json:"mod_time"`
	IsInUse      bool      `json:"is_in_use"`
}

type CleanTaskResult struct {
	TaskID          string     `json:"task_id"`
	Status          TaskStatus `json:"status"`
	FilesToDelete   []FileInfo `json:"files_to_delete"`
	FilesDeleted    []FileInfo `json:"files_deleted"`
	TotalSizeToDelete int64    `json:"total_size_to_delete"`
	TotalSizeDeleted int64    `json:"total_size_deleted"`
	FilesSkipped    []FileInfo `json:"files_skipped"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	Error           string     `json:"error,omitempty"`
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type CleanHistoryRecord struct {
	TaskID      string    `json:"task_id"`
	TargetDir   string    `json:"target_dir"`
	Mode        CleanMode `json:"mode"`
	Status      TaskStatus `json:"status"`
	FilesDeleted int       `json:"files_deleted"`
	SizeDeleted int64     `json:"size_deleted"`
	ExecTime    time.Time `json:"exec_time"`
}

type Request struct {
	Type    RequestType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Response struct {
	Status  ResponseStatus `json:"status"`
	Message string         `json:"message,omitempty"`
	Payload any            `json:"payload,omitempty"`
}

type SubmitTaskResponse struct {
	TaskID string `json:"task_id"`
}

type GetHistoryResponse struct {
	Records []CleanHistoryRecord `json:"records"`
}

type GetTaskResponse struct {
	Result CleanTaskResult `json:"result"`
}

type ListTasksResponse struct {
	Tasks []CleanTaskResult `json:"tasks"`
}
