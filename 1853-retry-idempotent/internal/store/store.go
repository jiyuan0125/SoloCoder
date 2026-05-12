package store

import (
	"sync"
	"time"

	"idempotent-retry/internal/models"
)

type Store struct {
	mu             sync.RWMutex
	tasks          map[string]*models.Task
	idempotencyMap map[string]string
}

func New() *Store {
	return &Store{
		tasks:          make(map[string]*models.Task),
		idempotencyMap: make(map[string]string),
	}
}

func (s *Store) CreateTask(task *models.Task) (*models.Task, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.IdempotencyKey != "" {
		if existingTaskID, exists := s.idempotencyMap[task.IdempotencyKey]; exists {
			return s.tasks[existingTaskID], false, nil
		}
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Status = models.StatusPending
	task.Attempts = 0
	task.History = make([]*models.ExecutionRecord, 0)
	task.Callbacks = make([]string, 0)

	s.tasks[task.ID] = task

	if task.IdempotencyKey != "" {
		s.idempotencyMap[task.IdempotencyKey] = task.ID
	}

	return task, true, nil
}

func (s *Store) GetTask(taskID string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	return task, exists
}

func (s *Store) GetTaskByIdempotencyKey(key string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if taskID, exists := s.idempotencyMap[key]; exists {
		return s.tasks[taskID], true
	}
	return nil, false
}

func (s *Store) UpdateTask(task *models.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.UpdatedAt = time.Now()
	s.tasks[task.ID] = task
}

func (s *Store) AddExecutionRecord(taskID string, record *models.ExecutionRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		task.History = append(task.History, record)
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) SetTaskResult(taskID string, result *models.TaskResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		task.Result = result
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) SetTaskStatus(taskID string, status models.TaskStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		task.Status = status
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) IncrementAttempt(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		task.Attempts++
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) AddCallback(taskID string, callbackURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		task.Callbacks = append(task.Callbacks, callbackURL)
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) GetPendingTasks() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []*models.Task
	for _, task := range s.tasks {
		if task.Status == models.StatusPending {
			pending = append(pending, task)
		}
	}
	return pending
}
