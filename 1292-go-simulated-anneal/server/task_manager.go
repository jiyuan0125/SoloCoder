package main

import (
	"sync"

	"github.com/tsp-simulated-anneal/annealing"
	"github.com/tsp-simulated-anneal/common"
)

type TaskStatus string

const (
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID     string
	Status TaskStatus
	Cities []common.City
	Config common.SolverConfig
	Solver *annealing.Solver
	Result *annealing.SolverResult
	Error  error
	Mu     sync.RWMutex
}

type TaskManager struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
	}
}

func (tm *TaskManager) CreateTask(id string, cities []common.City, config common.SolverConfig) *Task {
	task := &Task{
		ID:     id,
		Status: TaskStatusRunning,
		Cities: cities,
		Config: config,
	}
	tm.mu.Lock()
	tm.tasks[id] = task
	tm.mu.Unlock()
	return task
}

func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mu.RLock()
	task, exists := tm.tasks[id]
	tm.mu.RUnlock()
	return task, exists
}

func (tm *TaskManager) UpdateTaskStatus(id string, status TaskStatus) {
	tm.mu.RLock()
	task, exists := tm.tasks[id]
	tm.mu.RUnlock()
	if exists {
		task.Mu.Lock()
		task.Status = status
		task.Mu.Unlock()
	}
}

func (tm *TaskManager) SetTaskResult(id string, result *annealing.SolverResult) {
	tm.mu.RLock()
	task, exists := tm.tasks[id]
	tm.mu.RUnlock()
	if exists {
		task.Mu.Lock()
		task.Result = result
		task.Status = TaskStatusCompleted
		task.Mu.Unlock()
	}
}

func (tm *TaskManager) SetTaskError(id string, err error) {
	tm.mu.RLock()
	task, exists := tm.tasks[id]
	tm.mu.RUnlock()
	if exists {
		task.Mu.Lock()
		task.Error = err
		task.Status = TaskStatusFailed
		task.Mu.Unlock()
	}
}
