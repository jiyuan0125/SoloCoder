package webapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cron-executor/pkg/model"
	"cron-executor/pkg/server"
)

type WebAPI struct {
	port   int
	server *server.Server
	srv    *http.Server
}

func NewWebAPI(port int, srv *server.Server) *WebAPI {
	return &WebAPI{
		port:   port,
		server: srv,
	}
}

func (wa *WebAPI) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", wa.handleListTasks)
	mux.HandleFunc("/tasks/", wa.handleTask)
	mux.HandleFunc("/trigger/", wa.handleTrigger)
	mux.HandleFunc("/executions", wa.handleExecutions)
	mux.HandleFunc("/stats", wa.handleStats)

	addr := fmt.Sprintf(":%d", wa.port)
	wa.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		wa.srv.ListenAndServe()
	}()

	return nil
}

func (wa *WebAPI) Stop() {
	if wa.srv != nil {
		wa.srv.Close()
	}
}

type taskResponse struct {
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
	ActiveExecID string        `json:"active_exec_id,omitempty"`
}

func (wa *WebAPI) handleListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := wa.server.ListTasks()

	var responses []taskResponse
	for _, t := range tasks {
		tr := taskResponse{
			ID:           t.ID,
			Name:         t.Name,
			CronExpr:     t.CronExpr,
			Command:      t.Command,
			Timeout:      t.Timeout,
			MaxRetries:   t.MaxRetries,
			Status:       string(t.Status),
			Dependencies: t.Dependencies,
			ActiveExecID: t.ActiveExecID,
		}
		if !t.LastRun.IsZero() {
			tr.LastRun = t.LastRun.Format(time.RFC3339)
		}
		if !t.NextRun.IsZero() {
			tr.NextRun = t.NextRun.Format(time.RFC3339)
		}
		responses = append(responses, tr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (wa *WebAPI) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if path == "" {
		http.Error(w, "task name required", http.StatusBadRequest)
		return
	}

	task := wa.server.GetTaskByName(path)
	if task == nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	tr := taskResponse{
		ID:           task.ID,
		Name:         task.Name,
		CronExpr:     task.CronExpr,
		Command:      task.Command,
		Timeout:      task.Timeout,
		MaxRetries:   task.MaxRetries,
		Status:       string(task.Status),
		Dependencies: task.Dependencies,
		ActiveExecID: task.ActiveExecID,
	}
	if !task.LastRun.IsZero() {
		tr.LastRun = task.LastRun.Format(time.RFC3339)
	}
	if !task.NextRun.IsZero() {
		tr.NextRun = task.NextRun.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tr)
}

type triggerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ExecID  string `json:"exec_id,omitempty"`
}

func (wa *WebAPI) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/trigger/")
	if path == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(triggerResponse{
			Success: false,
			Message: "task name required",
		})
		return
	}

	execID, err := wa.server.TriggerTaskManually(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(triggerResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggerResponse{
		Success: true,
		Message: "task triggered",
		ExecID:  execID,
	})
}

type executionResponse struct {
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

func (wa *WebAPI) handleExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskName := r.URL.Query().Get("task")
	limit := 20

	executions := wa.server.GetExecutions(taskName, limit)

	var responses []executionResponse
	for _, e := range executions {
		er := executionResponse{
			ID:              e.ID,
			TaskID:          e.TaskID,
			TaskName:        e.TaskName,
			StartTime:       e.StartTime.Format(time.RFC3339),
			ExitCode:        e.ExitCode,
			Status:          string(e.Status),
			RetryCount:      e.RetryCount,
			IsTimeout:       e.IsTimeout,
			IsManualTrigger: e.IsManualTrigger,
		}
		if !e.EndTime.IsZero() {
			er.EndTime = e.EndTime.Format(time.RFC3339)
			er.Duration = e.Duration.String()
		}
		responses = append(responses, er)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

type statsResponse struct {
	TotalTasks       int                  `json:"total_tasks"`
	RunningTasks     int                  `json:"running_tasks"`
	TotalExecutions  int64                `json:"total_executions"`
	SuccessRate      float64              `json:"success_rate"`
	AvgDuration      time.Duration        `json:"avg_duration"`
	RecentFailures   []executionResponse  `json:"recent_failures"`
}

func (wa *WebAPI) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := wa.server.GetStats()

	var recentFailures []executionResponse
	for _, e := range stats.RecentFailures {
		er := executionResponse{
			ID:              e.ID,
			TaskID:          e.TaskID,
			TaskName:        e.TaskName,
			StartTime:       e.StartTime.Format(time.RFC3339),
			ExitCode:        e.ExitCode,
			Status:          string(e.Status),
			RetryCount:      e.RetryCount,
			IsTimeout:       e.IsTimeout,
			IsManualTrigger: e.IsManualTrigger,
		}
		if !e.EndTime.IsZero() {
			er.EndTime = e.EndTime.Format(time.RFC3339)
		}
		recentFailures = append(recentFailures, er)
	}

	runningTasks := 0
	for _, t := range wa.server.ListTasks() {
		if t.Status == model.TaskStatusRunning {
			runningTasks++
		}
	}

	resp := statsResponse{
		TotalTasks:       len(wa.server.ListTasks()),
		RunningTasks:     runningTasks,
		TotalExecutions:  stats.TotalExecutions,
		SuccessRate:      stats.SuccessRate,
		AvgDuration:      stats.AvgDuration,
		RecentFailures:   recentFailures,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
