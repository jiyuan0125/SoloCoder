package scheduler

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"scheduler/internal/cron"
	"scheduler/internal/protocol"
)

const (
	defaultMaxRetry     = 3
	defaultRetryInterval = 60 * time.Second
	defaultTimeout      = 60 * time.Second
	maxRecords          = 100
)

type Task struct {
	Config     protocol.TaskConfig
	CronField  *cron.CronField
	NextRun    time.Time
	IsRunning  bool
	LastExec   *protocol.ExecutionRecord
	Executions []protocol.ExecutionRecord
	Deleted    bool
	mu         sync.Mutex
}

type Scheduler struct {
	tasks      map[string]*Task
	executions []protocol.ExecutionRecord
	taskMu     sync.RWMutex
	execMu     sync.RWMutex
	maxRecords int
	configFile string
	stopChan   chan struct{}
	running    bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:      make(map[string]*Task),
		executions: make([]protocol.ExecutionRecord, 0),
		maxRecords: maxRecords,
		stopChan:   make(chan struct{}),
	}
}

func (s *Scheduler) SetConfigFile(path string) {
	s.configFile = path
}

func (s *Scheduler) SetMaxRecords(max int) {
	s.maxRecords = max
}

func (s *Scheduler) LoadConfig() error {
	if s.configFile == "" {
		return nil
	}

	file, err := os.Open(s.configFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var configs []protocol.TaskConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	for _, cfg := range configs {
		task := &Task{
			Config: cfg,
		}

		cf, err := cron.ParseCron(cfg.CronExpr)
		if err != nil {
			task.Config.Disabled = true
		} else {
			task.CronField = cf
			task.NextRun = cf.Next(time.Now())
		}

		if task.Config.MaxRetry <= 0 {
			task.Config.MaxRetry = defaultMaxRetry
		}
		if task.Config.RetryInterval <= 0 {
			task.Config.RetryInterval = defaultRetryInterval
		}
		if task.Config.Timeout <= 0 {
			task.Config.Timeout = defaultTimeout
		}

		s.taskMu.Lock()
		s.tasks[task.Config.Name] = task
		s.taskMu.Unlock()
	}

	return nil
}

func (s *Scheduler) SaveConfig() error {
	if s.configFile == "" {
		return nil
	}

	s.taskMu.RLock()
	configs := make([]protocol.TaskConfig, 0, len(s.tasks))
	for _, task := range s.tasks {
		if !task.Deleted {
			configs = append(configs, task.Config)
		}
	}
	s.taskMu.RUnlock()

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configFile, data, 0644)
}

func (s *Scheduler) AddTask(cfg protocol.TaskConfig) error {
	cf, err := cron.ParseCron(cfg.CronExpr)
	if err != nil {
		return err
	}

	if cfg.MaxRetry <= 0 {
		cfg.MaxRetry = defaultMaxRetry
	}
	if cfg.RetryInterval <= 0 {
		cfg.RetryInterval = defaultRetryInterval
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	task := &Task{
		Config:    cfg,
		CronField: cf,
		NextRun:   cf.Next(time.Now()),
	}

	s.taskMu.Lock()
	defer s.taskMu.Unlock()

	if existing, ok := s.tasks[cfg.Name]; ok && !existing.Deleted {
		return os.ErrExist
	}

	s.tasks[cfg.Name] = task

	if err := s.SaveConfig(); err != nil {
		delete(s.tasks, cfg.Name)
		return err
	}

	return nil
}

func (s *Scheduler) DeleteTask(name string) error {
	s.taskMu.Lock()
	defer s.taskMu.Unlock()

	task, ok := s.tasks[name]
	if !ok || task.Deleted {
		return os.ErrNotExist
	}

	task.Deleted = true

	if err := s.SaveConfig(); err != nil {
		task.Deleted = false
		return err
	}

	return nil
}

func (s *Scheduler) GetTask(name string) (*Task, error) {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	task, ok := s.tasks[name]
	if !ok || task.Deleted {
		return nil, os.ErrNotExist
	}
	return task, nil
}

func (s *Scheduler) ListTasks() []*Task {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	result := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if !task.Deleted {
			result = append(result, task)
		}
	}
	return result
}

func (s *Scheduler) GetTaskStatus(name string) (*protocol.TaskStatus, error) {
	task, err := s.GetTask(name)
	if err != nil {
		return nil, err
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	status := &protocol.TaskStatus{
		TaskConfig:    task.Config,
		NextRun:       task.NextRun,
		LastExecution: task.LastExec,
		IsRunning:     task.IsRunning,
	}
	return status, nil
}

func (s *Scheduler) ListTaskStatuses() []protocol.TaskStatus {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	result := make([]protocol.TaskStatus, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.Deleted {
			continue
		}
		task.mu.Lock()
		status := protocol.TaskStatus{
			TaskConfig:    task.Config,
			NextRun:       task.NextRun,
			LastExecution: task.LastExec,
			IsRunning:     task.IsRunning,
		}
		task.mu.Unlock()
		result = append(result, status)
	}
	return result
}

func (s *Scheduler) AddExecution(rec protocol.ExecutionRecord) {
	s.execMu.Lock()
	defer s.execMu.Unlock()

	s.executions = append(s.executions, rec)
	if len(s.executions) > s.maxRecords {
		s.executions = s.executions[len(s.executions)-s.maxRecords:]
	}
}

func (s *Scheduler) GetExecutions(taskName string) []protocol.ExecutionRecord {
	s.execMu.RLock()
	defer s.execMu.RUnlock()

	result := make([]protocol.ExecutionRecord, 0)
	for _, rec := range s.executions {
		if rec.TaskName == taskName {
			result = append(result, rec)
		}
	}
	return result
}

func (s *Scheduler) GetAllExecutions() []protocol.ExecutionRecord {
	s.execMu.RLock()
	defer s.execMu.RUnlock()

	result := make([]protocol.ExecutionRecord, len(s.executions))
	copy(result, s.executions)
	return result
}

func (s *Scheduler) Start(ctx context.Context) {
	s.running = true
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.running = false
			return
		case <-s.stopChan:
			s.running = false
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

func (s *Scheduler) checkAndRunTasks() {
	now := time.Now()
	s.taskMu.RLock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	s.taskMu.RUnlock()

	for _, task := range tasks {
		task.mu.Lock()
		if task.Deleted || task.Config.Disabled || task.IsRunning {
			task.mu.Unlock()
			continue
		}

		if now.After(task.NextRun) || now.Equal(task.NextRun) {
			task.IsRunning = true
			task.mu.Unlock()
			go s.executeTask(task)
		} else {
			task.mu.Unlock()
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) IsRunning() bool {
	return s.running
}
