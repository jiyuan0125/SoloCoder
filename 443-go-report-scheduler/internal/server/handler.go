package server

import (
	"encoding/json"
	"go-report-scheduler/pkg/common"
	"net/http"
	"strings"
)

type Handler struct {
	store     *Store
	scheduler *Scheduler
}

func NewHandler(store *Store, scheduler *Scheduler) *Handler {
	return &Handler{
		store:     store,
		scheduler: scheduler,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/tasks", h.handleTasks)
	mux.HandleFunc("/api/tasks/", h.handleTaskByID)
	mux.HandleFunc("/api/executions", h.handleExecutions)
	mux.HandleFunc("/api/executions/", h.handleExecutionByID)
}

func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodPut:
		h.updateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if id == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, id)
	case http.MethodDelete:
		h.deleteTask(w, id)
	case http.MethodPost:
		if strings.HasSuffix(r.URL.Path, "/pause") {
			h.pauseTask(w, strings.TrimSuffix(id, "/pause"))
		} else if strings.HasSuffix(r.URL.Path, "/resume") {
			h.resumeTask(w, strings.TrimSuffix(id, "/resume"))
		} else if strings.HasSuffix(r.URL.Path, "/trigger") {
			h.triggerTask(w, strings.TrimSuffix(id, "/trigger"))
		} else {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.listExecutions(w, r)
}

func (h *Handler) handleExecutionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	id := strings.TrimPrefix(r.URL.Path, "/api/executions/")
	if id == "" {
		http.Error(w, "Execution ID required", http.StatusBadRequest)
		return
	}
	h.getExecution(w, id)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req common.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.store.CreateTask(&req)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(common.CreateTaskResponse{Task: task})
}

func (h *Handler) getTask(w http.ResponseWriter, id string) {
	task, err := h.store.GetTask(id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetTaskResponse{Task: task})
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	var status *common.TaskStatus
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		ts := common.TaskStatus(statusStr)
		status = &ts
	}

	tasks := h.store.ListTasks(status)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ListTasksResponse{Tasks: tasks})
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	var req common.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.store.UpdateTask(&req)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.UpdateTaskResponse{Task: task})
}

func (h *Handler) deleteTask(w http.ResponseWriter, id string) {
	err := h.store.DeleteTask(id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.SuccessResponse{Message: "Task deleted successfully"})
}

func (h *Handler) pauseTask(w http.ResponseWriter, id string) {
	task, err := h.store.PauseTask(id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.PauseTaskResponse{Task: task})
}

func (h *Handler) resumeTask(w http.ResponseWriter, id string) {
	task, err := h.store.ResumeTask(id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ResumeTaskResponse{Task: task})
}

func (h *Handler) triggerTask(w http.ResponseWriter, id string) {
	execID, err := h.scheduler.ManualTrigger(id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(common.ManualTriggerResponse{ExecutionID: execID})
}

func (h *Handler) getExecution(w http.ResponseWriter, id string) {
	exec, err := h.store.GetExecution(id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetExecutionResponse{Execution: exec})
}

func (h *Handler) listExecutions(w http.ResponseWriter, r *http.Request) {
	req := &common.ListExecutionsRequest{}
	
	if taskID := r.URL.Query().Get("task_id"); taskID != "" {
		req.TaskID = &taskID
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		es := common.ExecutionStatus(statusStr)
		req.Status = &es
	}
	if triggerTypeStr := r.URL.Query().Get("trigger_type"); triggerTypeStr != "" {
		tt := common.TriggerType(triggerTypeStr)
		req.TriggerType = &tt
	}

	executions := h.store.ListExecutions(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ListExecutionsResponse{Executions: executions})
}

func (h *Handler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(common.ErrorResponse{Error: message})
}
