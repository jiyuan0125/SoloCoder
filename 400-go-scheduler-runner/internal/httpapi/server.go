package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"scheduler/internal/cron"
	"scheduler/internal/protocol"
	"scheduler/internal/scheduler"
)

type HTTPServer struct {
	sched     *scheduler.Scheduler
	server    *http.Server
	addr      string
}

type apiResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type addTaskRequest struct {
	Name          string        `json:"name"`
	Command       string        `json:"command"`
	CronExpr      string        `json:"cron_expr"`
	Timeout       time.Duration `json:"timeout"`
	MaxRetry      int           `json:"max_retry"`
	RetryInterval time.Duration `json:"retry_interval"`
}

func NewHTTPServer(addr string, s *scheduler.Scheduler) *HTTPServer {
	return &HTTPServer{
		sched: s,
		addr:  addr,
	}
}

func (h *HTTPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", h.handleTasks)
	mux.HandleFunc("/tasks/", h.handleTask)
	mux.HandleFunc("/executions", h.handleExecutions)
	mux.HandleFunc("/executions/", h.handleTaskExecutions)

	h.server = &http.Server{
		Addr:    h.addr,
		Handler: mux,
	}

	go func() {
		h.server.ListenAndServe()
	}()

	return nil
}

func (h *HTTPServer) Stop(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HTTPServer) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.addTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPServer) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.sched.ListTaskStatuses()

	resp := apiResponse{
		Success: true,
		Data:    tasks,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPServer) addTask(w http.ResponseWriter, r *http.Request) {
	var req addTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := apiResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Name == "" {
		resp := apiResponse{
			Success: false,
			Error:   "Task name is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Command == "" {
		resp := apiResponse{
			Success: false,
			Error:   "Command is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.CronExpr == "" {
		resp := apiResponse{
			Success: false,
			Error:   "Cron expression is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if _, err := cron.ParseCron(req.CronExpr); err != nil {
		resp := apiResponse{
			Success: false,
			Error:   "Invalid cron expression: " + err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	cfg := protocol.TaskConfig{
		Name:          req.Name,
		Command:       req.Command,
		CronExpr:      req.CronExpr,
		Timeout:       req.Timeout,
		MaxRetry:      req.MaxRetry,
		RetryInterval: req.RetryInterval,
		Disabled:      false,
	}

	if err := h.sched.AddTask(cfg); err != nil {
		if os.IsExist(err) {
			resp := apiResponse{
				Success: false,
				Error:   "Task already exists",
			}
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := apiResponse{
				Success: false,
				Error:   "Failed to add task: " + err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
		}
		return
	}

	resp := apiResponse{
		Success: true,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPServer) handleTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Path[len("/tasks/"):]
	if name == "" {
		resp := apiResponse{
			Success: false,
			Error:   "Task name is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, name)
	case http.MethodDelete:
		h.deleteTask(w, r, name)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPServer) getTask(w http.ResponseWriter, r *http.Request, name string) {
	status, err := h.sched.GetTaskStatus(name)
	if err != nil {
		if os.IsNotExist(err) {
			resp := apiResponse{
				Success: false,
				Error:   "Task not found",
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := apiResponse{
				Success: false,
				Error:   "Failed to get task: " + err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
		}
		return
	}

	resp := apiResponse{
		Success: true,
		Data:    status,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPServer) deleteTask(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.sched.DeleteTask(name); err != nil {
		if os.IsNotExist(err) {
			resp := apiResponse{
				Success: false,
				Error:   "Task not found",
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := apiResponse{
				Success: false,
				Error:   "Failed to delete task: " + err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
		}
		return
	}

	resp := apiResponse{
		Success: true,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPServer) handleExecutions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	executions := h.sched.GetAllExecutions()

	resp := apiResponse{
		Success: true,
		Data:    executions,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPServer) handleTaskExecutions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Path[len("/executions/"):]
	if name == "" {
		resp := apiResponse{
			Success: false,
			Error:   "Task name is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	executions := h.sched.GetExecutions(name)

	resp := apiResponse{
		Success: true,
		Data:    executions,
	}
	json.NewEncoder(w).Encode(resp)
}
