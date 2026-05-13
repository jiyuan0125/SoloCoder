package models

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusRunning TaskStatus = "running"
	TaskStatusPaused  TaskStatus = "paused"
)

type ExecutionStatus string

const (
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusTimeout ExecutionStatus = "timeout"
)

type TaskType string

const (
	TaskTypeCommand TaskType = "command"
	TaskTypeHTTP    TaskType = "http"
)

type Task struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	CronExpr    string     `json:"cron_expr"`
	Type        TaskType   `json:"type"`
	Command     string     `json:"command,omitempty"`
	CallbackURL string     `json:"callback_url,omitempty"`
	TimeoutSec  int        `json:"timeout_sec"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`

	mu         sync.Mutex
	nextRun    time.Time
	executions []*ExecutionLog
}

type ExecutionLog struct {
	ID         string          `json:"id"`
	TaskID     string          `json:"task_id"`
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt time.Time       `json:"finished_at"`
	DurationMS int64           `json:"duration_ms"`
	Status     ExecutionStatus `json:"status"`
	Result     string          `json:"result"`
}

type CreateTaskRequest struct {
	Name        string   `json:"name"`
	CronExpr    string   `json:"cron_expr"`
	Type        TaskType `json:"type"`
	Command     string   `json:"command"`
	CallbackURL string   `json:"callback_url"`
	TimeoutSec  int      `json:"timeout_sec"`
}

type TaskResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	CronExpr    string     `json:"cron_expr"`
	Type        TaskType   `json:"type"`
	Command     string     `json:"command,omitempty"`
	CallbackURL string     `json:"callback_url,omitempty"`
	TimeoutSec  int        `json:"timeout_sec"`
	Status      TaskStatus `json:"status"`
	NextRun     string     `json:"next_run,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TaskDetailResponse struct {
	TaskResponse
	RecentExecutions []*ExecutionLog `json:"recent_executions"`
}

func NewTask(req *CreateTaskRequest) *Task {
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 30
	}
	return &Task{
		ID:          uuid.NewString(),
		Name:        req.Name,
		CronExpr:    req.CronExpr,
		Type:        req.Type,
		Command:     req.Command,
		CallbackURL: req.CallbackURL,
		TimeoutSec:  req.TimeoutSec,
		Status:      TaskStatusRunning,
		CreatedAt:   time.Now(),
		executions:  make([]*ExecutionLog, 0, 100),
	}
}

func (t *Task) SetNextRun(next time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextRun = next
}

func (t *Task) GetNextRun() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.nextRun
}

func (t *Task) AddExecution(log *ExecutionLog) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executions = append(t.executions, log)
	if len(t.executions) > 100 {
		t.executions = t.executions[len(t.executions)-100:]
	}
}

func (t *Task) GetExecutions() []*ExecutionLog {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]*ExecutionLog, len(t.executions))
	copy(result, t.executions)
	return result
}

func (t *Task) GetRecentExecutions(n int) []*ExecutionLog {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n > len(t.executions) {
		n = len(t.executions)
	}
	result := make([]*ExecutionLog, n)
	for i := 0; i < n; i++ {
		result[i] = t.executions[len(t.executions)-1-i]
	}
	return result
}

func (t *Task) ToResponse() *TaskResponse {
	resp := &TaskResponse{
		ID:          t.ID,
		Name:        t.Name,
		CronExpr:    t.CronExpr,
		Type:        t.Type,
		TimeoutSec:  t.TimeoutSec,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
	}
	if t.Type == TaskTypeCommand {
		resp.Command = t.Command
	} else {
		resp.CallbackURL = t.CallbackURL
	}
	if !t.nextRun.IsZero() {
		resp.NextRun = t.nextRun.Format(time.RFC3339)
	}
	return resp
}

func (t *Task) ToDetailResponse() *TaskDetailResponse {
	return &TaskDetailResponse{
		TaskResponse:     *t.ToResponse(),
		RecentExecutions: t.GetRecentExecutions(10),
	}
}
