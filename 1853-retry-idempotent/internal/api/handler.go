package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"idempotent-retry/internal/models"
	"idempotent-retry/internal/query"
	"idempotent-retry/internal/scheduler"
	"idempotent-retry/internal/store"
)

type Handler struct {
	store     *store.Store
	scheduler *scheduler.Scheduler
	query     *query.Service
}

func NewHandler(s *store.Store, sch *scheduler.Scheduler, q *query.Service) *Handler {
	return &Handler{
		store:     s,
		scheduler: sch,
		query:     q,
	}
}

type SubmitTaskRequest struct {
	URL            string        `json:"url"`
	Method         string        `json:"method"`
	RequestBody    []byte        `json:"request_body"`
	MaxRetries     int           `json:"max_retries"`
	RetryStrategy  string        `json:"retry_strategy"`
	BaseIntervalMs int           `json:"base_interval_ms"`
	IdempotencyKey string        `json:"idempotency_key"`
}

type SubmitTaskResponse struct {
	TaskID          string `json:"task_id"`
	IdempotencyKey  string `json:"idempotency_key,omitempty"`
	IsNew           bool   `json:"is_new"`
	ExistingStatus  string `json:"existing_status,omitempty"`
}

type TaskResponse struct {
	TaskID          string           `json:"task_id"`
	IdempotencyKey  string           `json:"idempotency_key"`
	URL             string           `json:"url"`
	Method          string           `json:"method"`
	MaxRetries      int              `json:"max_retries"`
	RetryStrategy   string           `json:"retry_strategy"`
	BaseIntervalMs  int64            `json:"base_interval_ms"`
	Attempts        int              `json:"attempts"`
	Status          string           `json:"status"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	Result          *ResultResponse  `json:"result,omitempty"`
}

type ResultResponse struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body,omitempty"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

func (h *Handler) SubmitTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req SubmitTaskRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	if req.Method == "" {
		req.Method = http.MethodPost
	}

	if req.MaxRetries <= 0 {
		req.MaxRetries = 3
	}

	if req.BaseIntervalMs <= 0 {
		req.BaseIntervalMs = 1000
	}

	strategy := models.StrategyFixed
	if req.RetryStrategy == "exponential" {
		strategy = models.StrategyExponential
	}

	taskID, err := generateTaskID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task := &models.Task{
		ID:             taskID,
		IdempotencyKey: req.IdempotencyKey,
		URL:            req.URL,
		Method:         req.Method,
		RequestBody:    req.RequestBody,
		MaxRetries:     req.MaxRetries,
		RetryStrategy:  strategy,
		BaseInterval:   time.Duration(req.BaseIntervalMs) * time.Millisecond,
	}

	savedTask, isNew, err := h.store.CreateTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isNew {
		h.scheduler.Enqueue(savedTask.ID)
	}

	resp := SubmitTaskResponse{
		TaskID:         savedTask.ID,
		IdempotencyKey: savedTask.IdempotencyKey,
		IsNew:          isNew,
		ExistingStatus: string(savedTask.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Path[len("/tasks/"):]
	if taskID == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	task, exists := h.query.GetTask(taskID)
	if !exists {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	resp := buildTaskResponse(task)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetTaskByIdempotencyKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	task, exists := h.query.GetTaskByIdempotencyKey(key)
	if !exists {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	resp := buildTaskResponse(task)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) SubscribeCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Path[len("/tasks/"):]
	taskID = taskID[:len(taskID)-len("/subscribe")]
	if taskID == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		CallbackURL string `json:"callback_url"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CallbackURL == "" {
		http.Error(w, "callback_url is required", http.StatusBadRequest)
		return
	}

	_, exists := h.query.GetTask(taskID)
	if !exists {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	h.store.AddCallback(taskID, req.CallbackURL)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func buildTaskResponse(task *models.Task) TaskResponse {
	resp := TaskResponse{
		TaskID:         task.ID,
		IdempotencyKey: task.IdempotencyKey,
		URL:            task.URL,
		Method:         task.Method,
		MaxRetries:     task.MaxRetries,
		RetryStrategy:  string(task.RetryStrategy),
		BaseIntervalMs: task.BaseInterval.Milliseconds(),
		Attempts:       task.Attempts,
		Status:         string(task.Status),
		CreatedAt:      task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      task.UpdatedAt.Format(time.RFC3339),
	}

	if task.Result != nil {
		resp.Result = &ResultResponse{
			StatusCode: task.Result.StatusCode,
			Body:       string(task.Result.Body),
			Success:    task.Result.Success,
			Error:      task.Result.Error,
		}
	}

	return resp
}

func generateTaskID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
