package common

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MsgTypePing          MessageType = "ping"
	MsgTypePong          MessageType = "pong"
	MsgTypeListTasks     MessageType = "list_tasks"
	MsgTypeTaskList      MessageType = "task_list"
	MsgTypeTriggerTask   MessageType = "trigger_task"
	MsgTypeTriggerResult MessageType = "trigger_result"
	MsgTypeGetStatus     MessageType = "get_status"
	MsgTypeTaskStatus    MessageType = "task_status"
	MsgTypeGetExecutions MessageType = "get_executions"
	MsgTypeExecutionList MessageType = "execution_list"
	MsgTypeGetStats      MessageType = "get_stats"
	MsgTypeStatsResult   MessageType = "stats_result"
	MsgTypeError         MessageType = "error"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ListTasksRequest struct {
}

type TaskInfo struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	CronExpr     string        `json:"cron_expr"`
	Command      string        `json:"command"`
	Timeout      time.Duration `json:"timeout"`
	MaxRetries   int           `json:"max_retries"`
	Status       string        `json:"status"`
	LastRun      string        `json:"last_run,omitempty"`
	NextRun      string        `json:"next_run,omitempty"`
	Dependencies []string      `json:"dependencies,omitempty"`
}

type TaskListResponse struct {
	Tasks []TaskInfo `json:"tasks"`
}

type TriggerTaskRequest struct {
	TaskName string `json:"task_name"`
}

type TriggerResultResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ExecID    string `json:"exec_id,omitempty"`
}

type GetStatusRequest struct {
	TaskName string `json:"task_name"`
}

type TaskStatusResponse struct {
	TaskInfo
	ActiveExecID string `json:"active_exec_id,omitempty"`
}

type GetExecutionsRequest struct {
	TaskName  string `json:"task_name,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type ExecutionInfo struct {
	ID              string `json:"id"`
	TaskID          string `json:"task_id"`
	TaskName        string `json:"task_name"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time,omitempty"`
	Duration        string `json:"duration,omitempty"`
	ExitCode        int    `json:"exit_code"`
	Status          string `json:"status"`
	RetryCount      int    `json:"retry_count"`
	IsTimeout       bool   `json:"is_timeout"`
	IsManualTrigger bool   `json:"is_manual_trigger"`
}

type ExecutionListResponse struct {
	Executions []ExecutionInfo `json:"executions"`
}

type GetStatsRequest struct {
}

type StatsResponse struct {
	TotalTasks        int           `json:"total_tasks"`
	RunningTasks      int           `json:"running_tasks"`
	TotalExecutions   int64         `json:"total_executions"`
	SuccessRate       float64       `json:"success_rate"`
	AvgDuration       time.Duration `json:"avg_duration"`
	RecentFailures    []ExecutionInfo `json:"recent_failures"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
