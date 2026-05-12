package query

import (
	"idempotent-retry/internal/models"
	"idempotent-retry/internal/store"
)

type Service struct {
	store *store.Store
}

func New(s *store.Store) *Service {
	return &Service{
		store: s,
	}
}

func (q *Service) GetTaskStatus(taskID string) (models.TaskStatus, bool) {
	task, exists := q.store.GetTask(taskID)
	if !exists {
		return "", false
	}
	return task.Status, true
}

func (q *Service) GetTask(taskID string) (*models.Task, bool) {
	return q.store.GetTask(taskID)
}

func (q *Service) GetTaskByIdempotencyKey(key string) (*models.Task, bool) {
	return q.store.GetTaskByIdempotencyKey(key)
}

func (q *Service) GetTaskHistory(taskID string) ([]*models.ExecutionRecord, bool) {
	task, exists := q.store.GetTask(taskID)
	if !exists {
		return nil, false
	}
	return task.History, true
}

func (q *Service) GetTaskResult(taskID string) (*models.TaskResult, bool) {
	task, exists := q.store.GetTask(taskID)
	if !exists {
		return nil, false
	}
	return task.Result, true
}

func (q *Service) GetTaskResultByIdempotencyKey(key string) (*models.TaskResult, bool) {
	task, exists := q.store.GetTaskByIdempotencyKey(key)
	if !exists {
		return nil, false
	}
	return task.Result, true
}
