package task

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"gocron/internal/config"
	"gocron/internal/logger"
)

type Status string

const (
	StatusIdle     Status = "idle"
	StatusRunning  Status = "running"
	StatusPaused   Status = "paused"
	StatusFailed   Status = "failed"
	StatusSuccess  Status = "success"
)

type Task struct {
	Name         string
	Config       *config.TaskConfig
	CronID       int
	Status       Status
	Running      bool
	StartTime    time.Time
	LastEndTime  time.Time
	LastExitCode int
	mu           sync.Mutex
}

type DetailRecord struct {
	ID        string
	TaskName  string
	Action    string
	Detail    string
	Time      time.Time
	ExitCode  int
	Output    string
	RetryNum  int
	IsManual  bool
}

type Manager struct {
	tasks   map[string]*Task
	logger  *logger.Logger
	mu      sync.RWMutex
	details []*DetailRecord
}

func NewManager(l *logger.Logger) *Manager {
	return &Manager{
		tasks:   make(map[string]*Task),
		logger:  l,
		details: make([]*DetailRecord, 0),
	}
}

func (m *Manager) LoadFromConfig(cfg *config.Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tc := range cfg.Tasks {
		if _, exists := m.tasks[tc.Name]; !exists {
			m.tasks[tc.Name] = &Task{
				Name:   tc.Name,
				Config: tc,
				Status: StatusIdle,
			}
		}
	}
}

func (m *Manager) Add(tc *config.TaskConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[tc.Name]; exists {
		return fmt.Errorf("任务 %s 已存在", tc.Name)
	}

	m.tasks[tc.Name] = &Task{
		Name:   tc.Name,
		Config: tc,
		Status: StatusIdle,
	}

	m.addDetail(tc.Name, "ADD", fmt.Sprintf("添加任务: %s", tc.Name), 0, "", 0, false)
	return nil
}

func (m *Manager) Get(name string) (*Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[name]
	return t, ok
}

func (m *Manager) List() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		list = append(list, t)
	}
	return list
}

func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[name]
	if !ok {
		return fmt.Errorf("任务 %s 不存在", name)
	}

	t.mu.Lock()
	if t.Running {
		t.mu.Unlock()
		return fmt.Errorf("任务 %s 正在执行中，请等待执行完成后再删除", name)
	}
	t.mu.Unlock()

	delete(m.tasks, name)
	m.addDetail(name, "DELETE", fmt.Sprintf("删除任务: %s", name), 0, "", 0, false)
	return nil
}

func (m *Manager) Pause(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[name]
	if !ok {
		return fmt.Errorf("任务 %s 不存在", name)
	}

	t.mu.Lock()
	if t.Config.Paused {
		t.mu.Unlock()
		return fmt.Errorf("任务 %s 已经是暂停状态", name)
	}
	t.Config.Paused = true
	t.Status = StatusPaused
	t.mu.Unlock()

	m.addDetail(name, "PAUSE", fmt.Sprintf("暂停任务: %s", name), 0, "", 0, false)
	return nil
}

func (m *Manager) Resume(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[name]
	if !ok {
		return fmt.Errorf("任务 %s 不存在", name)
	}

	t.mu.Lock()
	if !t.Config.Paused {
		t.mu.Unlock()
		return fmt.Errorf("任务 %s 已经是运行状态", name)
	}
	t.Config.Paused = false
	t.Status = StatusIdle
	t.mu.Unlock()

	m.addDetail(name, "RESUME", fmt.Sprintf("恢复任务: %s", name), 0, "", 0, false)
	return nil
}

func (m *Manager) SetCronID(name string, id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tasks[name]; ok {
		t.CronID = id
	}
}

func (m *Manager) IsPaused(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[name]
	return ok && t.Config.Paused
}

func (m *Manager) IsRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[name]
	if !ok {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Running
}

func (m *Manager) Execute(ctx context.Context, name string, isManual bool) error {
	m.mu.RLock()
	t, ok := m.tasks[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("任务 %s 不存在", name)
	}

	t.mu.Lock()
	if t.Running {
		t.mu.Unlock()
		if isManual {
			return fmt.Errorf("任务 %s 正在执行中，无法手动触发", name)
		}
		return nil
	}
	t.Running = true
	t.Status = StatusRunning
	t.StartTime = time.Now()
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		t.Running = false
		t.mu.Unlock()
	}()

	maxRetries := t.Config.Retries
	retryDelay := time.Duration(t.Config.RetryDelay) * time.Second
	if retryDelay <= 0 {
		retryDelay = 1 * time.Second
	}

	var lastErr error
	var lastOutput string
	var lastExitCode int
	retryCount := 0

	for retryCount <= maxRetries {
		exitCode, output, err := m.runCommand(ctx, t.Config.Command)
		lastOutput = output
		lastExitCode = exitCode

		execLog := &logger.ExecutionLog{
			TaskName:   name,
			StartTime:  t.StartTime,
			EndTime:    time.Now(),
			ExitCode:   exitCode,
			Output:     output,
			IsManual:   isManual,
			RetryCount: retryCount,
		}

		if err != nil {
			execLog.Error = err.Error()
			lastErr = err
		}

		m.logger.Write(execLog)

		if exitCode == 0 {
			t.mu.Lock()
			t.Status = StatusSuccess
			t.LastEndTime = time.Now()
			t.LastExitCode = 0
			t.mu.Unlock()
			m.addDetail(name, "EXECUTE_SUCCESS", fmt.Sprintf("任务执行成功 (重试次数: %d)", retryCount), exitCode, output, retryCount, isManual)
			return nil
		}

		retryCount++
		if retryCount <= maxRetries {
			time.Sleep(retryDelay)
		}
	}

	t.mu.Lock()
	t.Status = StatusFailed
	t.LastEndTime = time.Now()
	t.LastExitCode = lastExitCode
	t.mu.Unlock()
	m.addDetail(name, "EXECUTE_FAILED", fmt.Sprintf("任务执行失败 (重试次数: %d)", maxRetries), lastExitCode, lastOutput, maxRetries, isManual)

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("任务 %s 执行失败，退出码: %d", name, lastExitCode)
}

func (m *Manager) runCommand(ctx context.Context, cmd string) (int, string, error) {
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf

	err := c.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return exitCode, buf.String(), err
}

func (m *Manager) addDetail(taskName, action, detail string, exitCode int, output string, retryNum int, isManual bool) {
	m.details = append(m.details, &DetailRecord{
		TaskName: taskName,
		Action:   action,
		Detail:   detail,
		Time:     time.Now(),
		ExitCode: exitCode,
		Output:   output,
		RetryNum: retryNum,
		IsManual: isManual,
	})
}

func (m *Manager) GetDetails() []*DetailRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	details := make([]*DetailRecord, len(m.details))
	copy(details, m.details)
	return details
}

func (m *Manager) SaveConfig(path string) error {
	cfg := &config.Config{
		Tasks: make([]*config.TaskConfig, 0),
	}

	m.mu.RLock()
	for _, t := range m.tasks {
		cfg.Tasks = append(cfg.Tasks, t.Config)
	}
	m.mu.RUnlock()

	return config.Save(path, cfg)
}
