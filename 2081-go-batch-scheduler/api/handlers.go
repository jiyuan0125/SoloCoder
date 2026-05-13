package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-batch-scheduler/db"
	"go-batch-scheduler/models"
	"go-batch-scheduler/scheduler"

	"github.com/google/uuid"
)

type Handler struct {
	db    *db.Database
	sched *scheduler.Scheduler
}

func NewHandler(database *db.Database, s *scheduler.Scheduler) *Handler {
	return &Handler{db: database, sched: s}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var res models.Resource
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if res.ID == "" {
		res.ID = uuid.New().String()
	}
	if res.Name == "" || res.Type == "" {
		h.writeError(w, http.StatusBadRequest, "name and type are required")
		return
	}

	if err := h.db.CreateResource(&res); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) GetResources(w http.ResponseWriter, r *http.Request) {
	resources, err := h.db.ListResources()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, resources)
}

func (h *Handler) CreateResourceRelation(w http.ResponseWriter, r *http.Request) {
	var rel models.ResourceRelation
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if rel.FromID == "" || rel.ToID == "" || rel.Relation == "" {
		h.writeError(w, http.StatusBadRequest, "from_id, to_id, and relation are required")
		return
	}

	if err := h.db.CreateResourceRelation(rel.FromID, rel.ToID, rel.Relation); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, rel)
}

func (h *Handler) SubmitTask(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitTaskRequestJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	priority, ok := models.PriorityFromString(req.Priority)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid priority, must be high/medium/low")
		return
	}

	if req.TimeoutSec <= 0 {
		h.writeError(w, http.StatusBadRequest, "timeout must be at least 1 second")
		return
	}

	if req.Name == "" || req.ResourceID == "" {
		h.writeError(w, http.StatusBadRequest, "name and resource_id are required")
		return
	}

	res, err := h.db.GetResource(req.ResourceID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if res == nil {
		h.writeError(w, http.StatusBadRequest, "resource not found")
		return
	}

	taskID := req.ID
	if taskID == "" {
		taskID = uuid.New().String()
	}

	for _, depID := range req.Dependencies {
		if depID == taskID {
			h.writeError(w, http.StatusBadRequest, "cyclic dependency detected: cannot depend on self")
			return
		}
	}

	if h.db.CheckCyclicDependency(taskID, req.Dependencies) {
		h.writeError(w, http.StatusBadRequest, "cyclic dependency detected")
		return
	}

	for _, depID := range req.Dependencies {
		exists, err := h.db.GetTask(depID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if exists == nil {
			h.writeError(w, http.StatusBadRequest, fmt.Sprintf("dependent task %s not found", depID))
			return
		}
	}

	task := &models.Task{
		ID:           taskID,
		Name:         req.Name,
		Type:         req.Type,
		Priority:     priority,
		Timeout:      time.Duration(req.TimeoutSec) * time.Second,
		MaxRetries:   req.MaxRetries,
		RetryCount:   0,
		Status:       models.StatusQueued,
		FlowStatus:   models.FlowToSubmit,
		ResourceID:   req.ResourceID,
		CreatedAt:    time.Now(),
		IsFinalFailure: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := h.db.CreateTask(ctx, task); err != nil {
		cancel()
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cancel()

	for _, depID := range req.Dependencies {
		if err := h.db.AddDependency(taskID, depID); err != nil {
			h.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           task.ID,
		"name":         task.Name,
		"status":       task.Status,
		"flow_status":  task.FlowStatus,
		"created_at":   task.CreatedAt,
		"dependencies": req.Dependencies,
	})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if path == "" {
		h.writeError(w, http.StatusBadRequest, "task id required")
		return
	}

	task, err := h.db.GetTask(path)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		h.writeError(w, http.StatusNotFound, "task not found")
		return
	}

	h.writeJSON(w, http.StatusOK, task)
}

func (h *Handler) RetryTask(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/retry")
	if path == "" {
		h.writeError(w, http.StatusBadRequest, "task id required")
		return
	}

	if err := h.sched.RetryTask(path); err != nil {
		if err.Error() == "task not found" {
			h.writeError(w, http.StatusNotFound, "task not found")
		} else {
			h.writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "task scheduled for retry"})
}

func (h *Handler) AdvanceFlow(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/flow")
	if path == "" {
		h.writeError(w, http.StatusBadRequest, "task id required")
		return
	}

	var body struct {
		Action   string `json:"action"`
		Reason   string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.db.GetTask(path)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		h.writeError(w, http.StatusNotFound, "task not found")
		return
	}

	switch task.FlowStatus {
	case models.FlowToSubmit:
		if body.Action != "submit" {
			h.writeError(w, http.StatusBadRequest, "invalid action for current flow status")
			return
		}
		task.FlowStatus = models.FlowReviewing
	case models.FlowReviewing:
		if body.Action == "approve" {
			task.FlowStatus = models.FlowApproved
			h.sched.AddTask(task.ID, task.Priority)
		} else if body.Action == "reject" {
			task.FlowStatus = models.FlowRejected
			task.FailedReason = body.Reason
		} else {
			h.writeError(w, http.StatusBadRequest, "invalid action")
			return
		}
	case models.FlowRejected:
		if body.Action == "resubmit" {
			task.FlowStatus = models.FlowReviewing
			task.FailedReason = ""
		} else {
			h.writeError(w, http.StatusBadRequest, "invalid action")
			return
		}
	case models.FlowApproved, models.FlowExecuting, models.FlowCompleted:
		h.writeError(w, http.StatusBadRequest, "flow already advanced past this point")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := h.db.UpdateTask(ctx, task); err != nil {
		cancel()
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cancel()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          task.ID,
		"flow_status": task.FlowStatus,
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetStats()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetResourceSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.db.GetResourceSummaries()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, summaries)
}

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	ids, err := h.db.GetAllTaskIDs()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var tasks []*models.Task
	for _, id := range ids {
		task, err := h.db.GetTask(id)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	h.writeJSON(w, http.StatusOK, tasks)
}
