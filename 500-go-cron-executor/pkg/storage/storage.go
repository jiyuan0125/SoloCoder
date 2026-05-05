package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cron-executor/pkg/model"
)

type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

func NewStorage(dataDir string) *Storage {
	return &Storage{
		dataDir: dataDir,
	}
}

func (s *Storage) ensureDir() error {
	if s.dataDir == "" {
		return nil
	}
	return os.MkdirAll(s.dataDir, 0755)
}

func (s *Storage) tasksFile() string {
	return filepath.Join(s.dataDir, "tasks.json")
}

func (s *Storage) executionsFile() string {
	return filepath.Join(s.dataDir, "executions.json")
}

type persistedTask struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Status   model.TaskStatus `json:"status"`
	LastRun  int64            `json:"last_run"`
	NextRun  int64            `json:"next_run"`
}

type persistedExecution struct {
	ID              string                `json:"id"`
	TaskID          string                `json:"task_id"`
	TaskName        string                `json:"task_name"`
	StartTime       int64                 `json:"start_time"`
	EndTime         int64                 `json:"end_time"`
	Duration        int64                 `json:"duration_ns"`
	ExitCode        int                   `json:"exit_code"`
	Status          model.ExecutionStatus `json:"status"`
	RetryCount      int                   `json:"retry_count"`
	IsTimeout       bool                  `json:"is_timeout"`
	IsManualTrigger bool                  `json:"is_manual_trigger"`
}

func (s *Storage) SaveTasks(tasks []*model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureDir(); err != nil {
		return err
	}

	ptasks := make([]persistedTask, 0, len(tasks))
	for _, t := range tasks {
		pt := persistedTask{
			ID:     t.ID,
			Name:   t.Name,
			Status: t.Status,
		}
		if !t.LastRun.IsZero() {
			pt.LastRun = t.LastRun.UnixNano()
		}
		if !t.NextRun.IsZero() {
			pt.NextRun = t.NextRun.UnixNano()
		}
		ptasks = append(ptasks, pt)
	}

	data, err := json.MarshalIndent(ptasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.tasksFile(), data, 0644)
}

func (s *Storage) LoadTasks() ([]*model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.tasksFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ptasks []persistedTask
	if err := json.Unmarshal(data, &ptasks); err != nil {
		return nil, err
	}

	tasks := make([]*model.Task, 0, len(ptasks))
	for _, pt := range ptasks {
		t := &model.Task{
			ID:     pt.ID,
			Name:   pt.Name,
			Status: pt.Status,
		}
		if pt.LastRun > 0 {
			t.LastRun = timeFromUnixNano(pt.LastRun)
		}
		if pt.NextRun > 0 {
			t.NextRun = timeFromUnixNano(pt.NextRun)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (s *Storage) SaveExecution(exec *model.Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureDir(); err != nil {
		return err
	}

	pexecs, err := s.loadPersistedExecutions()
	if err != nil {
		return err
	}

	pe := persistedExecution{
		ID:              exec.ID,
		TaskID:          exec.TaskID,
		TaskName:        exec.TaskName,
		Status:          exec.Status,
		ExitCode:        exec.ExitCode,
		RetryCount:      exec.RetryCount,
		IsTimeout:       exec.IsTimeout,
		IsManualTrigger: exec.IsManualTrigger,
	}
	if !exec.StartTime.IsZero() {
		pe.StartTime = exec.StartTime.UnixNano()
	}
	if !exec.EndTime.IsZero() {
		pe.EndTime = exec.EndTime.UnixNano()
	}
	pe.Duration = exec.Duration.Nanoseconds()

	pexecs = append(pexecs, pe)
	if len(pexecs) > 10000 {
		pexecs = pexecs[len(pexecs)-10000:]
	}

	data, err := json.MarshalIndent(pexecs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.executionsFile(), data, 0644)
}

func (s *Storage) LoadExecutions() ([]*model.Execution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pexecs, err := s.loadPersistedExecutions()
	if err != nil {
		return nil, err
	}

	execs := make([]*model.Execution, 0, len(pexecs))
	for _, pe := range pexecs {
		exec := &model.Execution{
			ID:              pe.ID,
			TaskID:          pe.TaskID,
			TaskName:        pe.TaskName,
			Status:          pe.Status,
			ExitCode:        pe.ExitCode,
			RetryCount:      pe.RetryCount,
			IsTimeout:       pe.IsTimeout,
			IsManualTrigger: pe.IsManualTrigger,
		}
		if pe.StartTime > 0 {
			exec.StartTime = timeFromUnixNano(pe.StartTime)
		}
		if pe.EndTime > 0 {
			exec.EndTime = timeFromUnixNano(pe.EndTime)
		}
		exec.Duration = time.Duration(pe.Duration)
		execs = append(execs, exec)
	}

	return execs, nil
}

func (s *Storage) loadPersistedExecutions() ([]persistedExecution, error) {
	data, err := os.ReadFile(s.executionsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var pexecs []persistedExecution
	if err := json.Unmarshal(data, &pexecs); err != nil {
		return nil, err
	}

	return pexecs, nil
}

func timeFromUnixNano(n int64) time.Time {
	return time.Unix(0, n)
}
