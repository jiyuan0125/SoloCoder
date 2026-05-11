package main

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pingtool/pkg/ping"
)

type TaskInfo struct {
	ID        string
	Config    *ping.Config
	Pinger    *ping.Pinger
	Cancel    context.CancelFunc
	ResultChan chan *ping.Result
	IsRunning bool
	CreatedAt time.Time
	FinishedAt *time.Time
	mu        sync.RWMutex
}

type TaskManager struct {
	tasks map[string]*TaskInfo
	mu    sync.RWMutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*TaskInfo),
	}
}

func (tm *TaskManager) CreateTask(config *ping.Config) (*TaskInfo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	taskID := uuid.New().String()
	pinger := ping.NewPinger(config)
	ctx, cancel := context.WithCancel(context.Background())
	resultChan := make(chan *ping.Result, 100)

	task := &TaskInfo{
		ID:         taskID,
		Config:     config,
		Pinger:     pinger,
		Cancel:     cancel,
		ResultChan: resultChan,
		IsRunning:  true,
		CreatedAt:  time.Now(),
	}

	tm.tasks[taskID] = task

	go func() {
		err := pinger.Run(ctx, resultChan)
		task.mu.Lock()
		task.IsRunning = false
		now := time.Now()
		task.FinishedAt = &now
		task.mu.Unlock()
		close(resultChan)
		if err != nil && err != context.Canceled {
			_ = err
		}
	}()

	return task, nil
}

func (tm *TaskManager) GetTask(taskID string) (*TaskInfo, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, exists := tm.tasks[taskID]
	return task, exists
}

func (tm *TaskManager) StopTask(taskID string) bool {
	tm.mu.RLock()
	task, exists := tm.tasks[taskID]
	tm.mu.RUnlock()

	if !exists {
		return false
	}

	task.mu.Lock()
	if task.IsRunning {
		task.Cancel()
	}
	task.mu.Unlock()

	return true
}
