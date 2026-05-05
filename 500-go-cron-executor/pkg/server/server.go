package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cron-executor/pkg/executor"
	"cron-executor/pkg/model"
	"cron-executor/pkg/scheduler"
	"cron-executor/pkg/stats"
	"cron-executor/pkg/storage"
)

type Server struct {
	config     *model.Config
	scheduler  *scheduler.Scheduler
	executor   *executor.Executor
	stats      *stats.StatsManager
	storage    *storage.Storage
	tasks      map[string]*model.Task
	tasksMu    sync.RWMutex
	executions map[string]*model.Execution
	execsMu    sync.RWMutex
	dataDir    string
}

func NewServer(cfg *model.Config) *Server {
	s := &Server{
		config:     cfg,
		tasks:      make(map[string]*model.Task),
		executions: make(map[string]*model.Execution),
		dataDir:    cfg.DataDir,
	}

	s.stats = stats.NewStatsManager()
	s.storage = storage.NewStorage(cfg.DataDir)
	s.executor = executor.NewExecutor(cfg.LogDir, s.onExecutionComplete)
	s.scheduler = scheduler.NewScheduler(s.onTaskTrigger)

	return s
}

func (s *Server) LoadTasksFromConfig() error {
	for _, tc := range s.config.Tasks {
		task := &model.Task{
			ID:            generateID(),
			Name:          tc.Name,
			CronExpr:      tc.CronExpr,
			Command:       tc.Command,
			Timeout:       tc.Timeout,
			MaxRetries:    tc.MaxRetries,
			RetryInterval: tc.RetryInterval,
			Dependencies:  tc.Dependencies,
			Status:        model.TaskStatusIdle,
		}

		if task.Timeout <= 0 {
			task.Timeout = 5 * time.Minute
		}
		if task.MaxRetries < 0 {
			task.MaxRetries = 0
		}
		if task.RetryInterval <= 0 {
			task.RetryInterval = 5 * time.Second
		}

		s.tasksMu.Lock()
		s.tasks[task.ID] = task
		s.tasksMu.Unlock()

		if err := s.scheduler.AddTask(task); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) LoadPersistedData() error {
	if s.storage == nil {
		return nil
	}

	persistedTasks, err := s.storage.LoadTasks()
	if err != nil {
		return err
	}

	for _, pt := range persistedTasks {
		s.tasksMu.Lock()
		if existingTask, exists := s.tasks[pt.ID]; exists {
			existingTask.Status = pt.Status
			existingTask.LastRun = pt.LastRun
			existingTask.NextRun = pt.NextRun
		}
		s.tasksMu.Unlock()
	}

	executions, err := s.storage.LoadExecutions()
	if err != nil {
		return err
	}

	for _, exec := range executions {
		s.execsMu.Lock()
		s.executions[exec.ID] = exec
		s.execsMu.Unlock()
		s.stats.RecordExecution(exec)
	}

	return nil
}

func (s *Server) Start() {
	s.scheduler.Start()
}

func (s *Server) Stop() {
	s.scheduler.Stop()
	s.PersistData()
}

func (s *Server) onTaskTrigger(taskID string, isManual bool) {
	s.tasksMu.RLock()
	task, exists := s.tasks[taskID]
	s.tasksMu.RUnlock()

	if !exists {
		return
	}

	if s.executor.IsTaskRunning(taskID) {
		exec := &model.Execution{
			ID:              generateID(),
			TaskID:          task.ID,
			TaskName:        task.Name,
			StartTime:       time.Now(),
			EndTime:         time.Now(),
			Status:          model.ExecStatusSkipped,
			IsManualTrigger: isManual,
			Stdout:          "",
			Stderr:          "Task already running, skipped",
			ExitCode:        -1,
		}
		s.onExecutionComplete(exec)
		return
	}

	if !s.checkDependenciesReady(task) {
		exec := &model.Execution{
			ID:              generateID(),
			TaskID:          task.ID,
			TaskName:        task.Name,
			StartTime:       time.Now(),
			EndTime:         time.Now(),
			Status:          model.ExecStatusSkipped,
			IsManualTrigger: isManual,
			Stdout:          "",
			Stderr:          "Dependencies not ready",
			ExitCode:        -1,
		}
		s.onExecutionComplete(exec)
		return
	}

	execID := generateID()
	s.executor.Execute(task, execID, isManual)

	s.tasksMu.Lock()
	task.Status = model.TaskStatusRunning
	task.ActiveExecID = execID
	task.LastRun = time.Now()
	s.tasksMu.Unlock()
}

func (s *Server) checkDependenciesReady(task *model.Task) bool {
	if len(task.Dependencies) == 0 {
		return true
	}

	for _, depName := range task.Dependencies {
		depTask := s.scheduler.GetTaskByName(depName)
		if depTask == nil {
			return false
		}

		s.execsMu.RLock()
		foundSuccess := false
		for _, exec := range s.executions {
			if exec.TaskID == depTask.ID && exec.Status == model.ExecStatusSuccess {
				if exec.EndTime.After(task.LastRun) {
					foundSuccess = true
					break
				}
			}
		}
		s.execsMu.RUnlock()

		if !foundSuccess && !task.LastRun.IsZero() {
			return false
		}
	}

	return true
}

func (s *Server) onExecutionComplete(exec *model.Execution) {
	s.execsMu.Lock()
	s.executions[exec.ID] = exec
	s.execsMu.Unlock()

	s.tasksMu.Lock()
	if task, exists := s.tasks[exec.TaskID]; exists {
		if exec.Status == model.ExecStatusSuccess {
			task.Status = model.TaskStatusSuccess
		} else if exec.Status == model.ExecStatusFailed || exec.Status == model.ExecStatusTimeout {
			task.Status = model.TaskStatusFailed
		} else {
			task.Status = model.TaskStatusIdle
		}
		task.ActiveExecID = ""
	}
	s.tasksMu.Unlock()

	s.stats.RecordExecution(exec)

	if exec.Status == model.ExecStatusSuccess {
		s.triggerDependentTasks(exec.TaskID)
	}

	if s.storage != nil {
		s.storage.SaveExecution(exec)
	}
}

func (s *Server) triggerDependentTasks(completedTaskID string) {
	s.tasksMu.RLock()
	completedTask := s.tasks[completedTaskID]
	s.tasksMu.RUnlock()

	if completedTask == nil {
		return
	}

	s.tasksMu.RLock()
	for _, task := range s.tasks {
		for _, depName := range task.Dependencies {
			if depName == completedTask.Name {
				if s.checkDependenciesReady(task) {
					go func(t *model.Task) {
						s.onTaskTrigger(t.ID, false)
					}(task)
				}
			}
		}
	}
	s.tasksMu.RUnlock()
}

func (s *Server) PersistData() {
	if s.storage == nil {
		return
	}

	s.tasksMu.RLock()
	tasks := make([]*model.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	s.tasksMu.RUnlock()

	s.storage.SaveTasks(tasks)
}

func (s *Server) GetTaskByName(name string) *model.Task {
	return s.scheduler.GetTaskByName(name)
}

func (s *Server) ListTasks() []*model.Task {
	return s.scheduler.ListTasks()
}

func (s *Server) TriggerTaskManually(taskName string) (string, error) {
	task := s.scheduler.GetTaskByName(taskName)
	if task == nil {
		return "", fmt.Errorf("task not found: %s", taskName)
	}

	if s.executor.IsTaskRunning(task.ID) {
		return "", fmt.Errorf("task is already running")
	}

	execID := generateID()
	s.executor.Execute(task, execID, true)

	s.tasksMu.Lock()
	task.Status = model.TaskStatusRunning
	task.ActiveExecID = execID
	task.LastRun = time.Now()
	s.tasksMu.Unlock()

	return execID, nil
}

func (s *Server) GetStats() *stats.Stats {
	return s.stats.GetStats()
}

func (s *Server) GetExecutions(taskName string, limit int) []*model.Execution {
	s.execsMu.RLock()
	defer s.execsMu.RUnlock()

	var result []*model.Execution

	for _, exec := range s.executions {
		if taskName == "" || exec.TaskName == taskName {
			result = append(result, exec)
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[len(result)-limit:]
	}

	return result
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func LoadConfig(configPath string) (*model.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.TCPPort <= 0 {
		cfg.TCPPort = 8765
	}
	if cfg.WebAPIPort <= 0 {
		cfg.WebAPIPort = 8080
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "./logs"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}

	return &cfg, nil
}

func DefaultConfigPath() string {
	if envPath := os.Getenv("CRON_EXECUTOR_CONFIG"); envPath != "" {
		return envPath
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "config.json")
	}

	return "./config.json"
}
