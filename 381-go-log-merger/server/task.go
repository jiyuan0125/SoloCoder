package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go-log-merger/protocol"
)

type Task struct {
	ID             string
	Status         protocol.TaskStatus
	InputFiles     []string
	OutputFile     string
	TimeFormat     string
	Resume         bool
	TotalFiles     int
	ProcessedLines int64
	Progress       float64
	Error          error
	CreatedAt      time.Time
	StartedAt      time.Time
	CompletedAt    time.Time
	stopCh         chan struct{}
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

func (tm *TaskManager) SubmitTask(req *protocol.SubmitTaskRequest) (*Task, error) {
	for _, f := range req.InputFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", f)
		}
	}

	task := &Task{
		ID:         protocol.GenerateTaskID(),
		Status:     protocol.TaskStatusPending,
		InputFiles: req.InputFiles,
		OutputFile: req.OutputFile,
		TimeFormat: req.TimeFormat,
		Resume:     req.Resume,
		TotalFiles: len(req.InputFiles),
		CreatedAt:  time.Now(),
		stopCh:     make(chan struct{}),
	}

	tm.mu.Lock()
	tm.tasks[task.ID] = task
	tm.mu.Unlock()

	go tm.executeTask(task)

	return task, nil
}

func (tm *TaskManager) GetTask(taskID string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, exists := tm.tasks[taskID]
	return task, exists
}

func (tm *TaskManager) StopAll() {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	for _, task := range tm.tasks {
		if task.stopCh != nil {
			close(task.stopCh)
		}
	}
}

func (tm *TaskManager) updateTaskStatus(taskID string, status protocol.TaskStatus, progress float64, lines int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if task, exists := tm.tasks[taskID]; exists {
		task.Status = status
		task.Progress = progress
		task.ProcessedLines = lines
		if status == protocol.TaskStatusRunning {
			task.StartedAt = time.Now()
		} else if status == protocol.TaskStatusCompleted || status == protocol.TaskStatusFailed {
			task.CompletedAt = time.Now()
		}
	}
}

func (tm *TaskManager) setTaskError(taskID string, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if task, exists := tm.tasks[taskID]; exists {
		task.Status = protocol.TaskStatusFailed
		task.Error = err
		task.CompletedAt = time.Now()
	}
}

func (tm *TaskManager) executeTask(task *Task) {
	tm.updateTaskStatus(task.ID, protocol.TaskStatusRunning, 0.0, 0)

	checkpointDir := filepath.Join(".", "checkpoints")
	os.MkdirAll(checkpointDir, 0755)

	checkpointFile := filepath.Join(checkpointDir, task.ID+".ckpt")

	var checkpoint *Checkpoint
	if task.Resume {
		var err error
		checkpoint, err = LoadCheckpoint(checkpointFile)
		if err != nil && !os.IsNotExist(err) {
			tm.setTaskError(task.ID, err)
			return
		}
	}

	merger, err := NewLogMerger(task.InputFiles, task.OutputFile, task.TimeFormat)
	if err != nil {
		tm.setTaskError(task.ID, err)
		return
	}
	defer merger.Close()

	if checkpoint != nil {
		merger.RestoreCheckpoint(checkpoint)
	}

	progressCh := make(chan ProgressUpdate, 100)

	go func() {
		for update := range progressCh {
			tm.updateTaskStatus(task.ID, protocol.TaskStatusRunning, update.Progress, update.Lines)

			ckpt := merger.CreateCheckpoint()
			if err := SaveCheckpoint(checkpointFile, ckpt); err != nil {
				fmt.Printf("保存检查点失败: %v\n", err)
			}
		}
	}()

	select {
	case <-task.stopCh:
		merger.Close()
		close(progressCh)
		tm.updateTaskStatus(task.ID, protocol.TaskStatusPending, task.Progress, task.ProcessedLines)
		return
	default:
	}

	err = merger.Merge(progressCh, task.stopCh)
	close(progressCh)

	if err != nil {
		tm.setTaskError(task.ID, err)
		return
	}

	tm.updateTaskStatus(task.ID, protocol.TaskStatusCompleted, 100.0, merger.TotalLines())

	if err := os.Remove(checkpointFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("清理检查点失败: %v\n", err)
	}
}
