package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusSuccess   TaskStatus = "success"
	StatusFailed    TaskStatus = "failed"
)

type IntervalStrategy string

const (
	FixedInterval     IntervalStrategy = "fixed"
	ExponentialBackoff IntervalStrategy = "exponential"
)

type ExecutionRecord struct {
	Attempt   int       `json:"attempt"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  string    `json:"duration"`
	StatusCode int     `json:"status_code,omitempty"`
	Error      string    `json:"error,omitempty"`
	Body       string    `json:"body,omitempty"`
}

type Task struct {
	ID                 string           `json:"id"`
	IdempotencyKey    string           `json:"idempotency_key"`
	URL                string           `json:"url"`
	Method             string           `json:"method"`
	RequestBody        string           `json:"request_body"`
	MaxRetries         int              `json:"max_retries"`
	IntervalStrategy   IntervalStrategy `json:"interval_strategy"`
	BaseIntervalSec       int              `json:"base_interval_sec"`
	Status             TaskStatus       `json:"status"`
	Attempt            int              `json:"attempt"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Executions         []ExecutionRecord `json:"executions"`
	ResultStatusCode    int              `json:"result_status_code"`
	ResultBody        string           `json:"result_body"`
}

type CreateTaskRequest struct {
	IdempotencyKey  string           `json:"idempotency_key" binding:"required"`
	URL              string           `json:"url" binding:"required"`
	Method           string           `json:"method" binding:"required"`
	RequestBody      string           `json:"request_body"`
	MaxRetries       int              `json:"max_retries"`
	IntervalStrategy IntervalStrategy `json:"interval_strategy"`
	BaseIntervalSec  int              `json:"base_interval_sec"`
}

type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (r *RateLimiter) Check(key string) bool {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	history, exists := r.requests[key]
	if !exists {
		r.requests[key] = []time.Time{now}
		return true
	}

	cutoff := now.Add(-r.window)
	newHistory := []time.Time{}
	for _, t := range history {
		if t.After(cutoff) {
			newHistory = append(newHistory, t)
		}
	}

	if len(newHistory) >= r.limit {
		r.requests[key] = newHistory
		return false
	}

	newHistory = append(newHistory, now)
	r.requests[key] = newHistory
	return true
}

type TaskStore struct {
	mu         sync.RWMutex
	tasks       map[string]*Task
	idempotency map[string]*Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:       make(map[string]*Task),
		idempotency: make(map[string]*Task),
	}
}

func (s *TaskStore) GetTask(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

func (s *TaskStore) GetByIdempotencyKey(key string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.idempotency[key]
}

func (s *TaskStore) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		all = append(all, t)
	}
	return all
}

func (s *TaskStore) AddTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	s.idempotency[task.IdempotencyKey] = task
}

func (s *TaskStore) UpdateTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.UpdatedAt = time.Now()
}

type Engine struct {
	store    *TaskStore
	limiter  *RateLimiter
	httpCli  *http.Client
	nextID   uint64
}

func NewEngine() *Engine {
	return &Engine{
		store:    NewTaskStore(),
		limiter:  NewRateLimiter(10, time.Minute),
		httpCli:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *Engine) generateID() string {
	id := atomic.AddUint64(&e.nextID, 1)
	return fmt.Sprintf("task-%d", id)
}

func (e *Engine) CreateTask(req CreateTaskRequest) *Task {
	existing := e.store.GetByIdempotencyKey(req.IdempotencyKey)
	if existing != nil {
		return existing
	}

	if req.MaxRetries <= 0 {
		req.MaxRetries = 3
	}
	if req.IntervalStrategy == "" {
		req.IntervalStrategy = ExponentialBackoff
	}
	if req.BaseIntervalSec <= 0 {
		req.BaseIntervalSec = 1
	}

	now := time.Now()
	task := &Task{
		ID:               e.generateID(),
		IdempotencyKey:   req.IdempotencyKey,
		URL:              req.URL,
		Method:           req.Method,
		RequestBody:      req.RequestBody,
		MaxRetries:       req.MaxRetries,
		IntervalStrategy: req.IntervalStrategy,
		BaseIntervalSec: req.BaseIntervalSec,
		Status:           StatusPending,
		Attempt:          0,
		CreatedAt:        now,
		UpdatedAt:        now,
		Executions:       []ExecutionRecord{},
	}

	e.store.AddTask(task)

	go e.runTask(task)

	return task
}

func (e *Engine) calculateWaitTime(task *Task, attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	base := time.Duration(task.BaseIntervalSec) * time.Second

	if task.IntervalStrategy == FixedInterval {
		return base
	}

	backoff := base * time.Duration(1<<uint(attempt-1))
	jitter := time.Duration(rand.Int63n(int64(base) + 1))
	return backoff + jitter
}

func (e *Engine) runTask(task *Task) {
	task.Status = StatusRunning
	e.store.UpdateTask(task)

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		waitTime := e.calculateWaitTime(task, attempt)
		if waitTime > 0 {
			time.Sleep(waitTime)
		}

		if !e.limiter.Check(task.URL) {
			continue
		}

		start := time.Now()
		record := ExecutionRecord{
			Attempt:   attempt + 1,
			StartTime: start,
		}

		statusCode, respBody, err := e.executeRequest(task)
		end := time.Now()

		record.EndTime = end
		record.Duration = end.Sub(start).String()
		record.StatusCode = statusCode
		if err != nil {
			record.Error = err.Error()
		}
		record.Body = respBody

		task.Attempt = attempt + 1
		task.Executions = append(task.Executions, record)

		if err == nil && statusCode >= 200 && statusCode < 300 {
			task.Status = StatusSuccess
			task.ResultStatusCode = statusCode
			task.ResultBody = respBody
			e.store.UpdateTask(task)
			return
		}

		e.store.UpdateTask(task)
	}

	task.Status = StatusFailed
	e.store.UpdateTask(task)
}

func (e *Engine) executeRequest(task *Task) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body io.Reader
	if task.RequestBody != "" {
		body = bytes.NewBufferString(task.RequestBody)
	}

	req, err := http.NewRequestWithContext(ctx, task.Method, task.URL, body)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpCli.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody), nil
}

func main() {
	engine := NewEngine()

	r := gin.Default()

	r.POST("/tasks", func(c *gin.Context) {
		var req CreateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		task := engine.CreateTask(req)
		c.JSON(http.StatusCreated, gin.H{"task_id": task.ID})
	})

	r.GET("/tasks", func(c *gin.Context) {
		statusFilter := c.Query("status")
		all := engine.store.ListTasks()
		result := []*Task{}
		for _, t := range all {
			if statusFilter == "" || t.Status == TaskStatus(statusFilter) {
				result = append(result, t)
			}
		}
		c.JSON(http.StatusOK, result)
	})

	r.GET("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		task := engine.store.GetTask(id)
		if task == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	r.GET("/tasks/:id/result", func(c *gin.Context) {
		id := c.Param("id")
		task := engine.store.GetTask(id)
		if task == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		if task.Status != StatusSuccess {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not completed successfully"})
			return
		}

		if task.ResultBody != "" {
			var jsonBody interface{}
			if err := json.Unmarshal([]byte(task.ResultBody), &jsonBody); err == nil {
				c.JSON(task.ResultStatusCode, jsonBody)
				return
			}
		}

		c.Data(task.ResultStatusCode, "text/plain", []byte(task.ResultBody))
	})

	r.GET("/result-by-idempotency", func(c *gin.Context) {
		key := c.Query("idempotency_key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "idempotency_key is required"})
			return
		}

		task := engine.store.GetByIdempotencyKey(key)
		if task == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		if task.Status != StatusSuccess {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not completed successfully"})
			return
		}

		if task.ResultBody != "" {
			var jsonBody interface{}
			if err := json.Unmarshal([]byte(task.ResultBody), &jsonBody); err == nil {
				c.JSON(task.ResultStatusCode, jsonBody)
				return
			}
		}

		c.Data(task.ResultStatusCode, "text/plain", []byte(task.ResultBody))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
