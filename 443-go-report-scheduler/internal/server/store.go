package server

import (
	"errors"
	"go-report-scheduler/pkg/common"
	"sync"
	"time"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskAlreadyExists  = errors.New("task already exists")
	ErrExecutionNotFound  = errors.New("execution not found")
	ErrInvalidCronExpr    = errors.New("invalid cron expression")
	ErrDependencyNotFound = errors.New("dependency task not found")
)

type Store struct {
	tasks       map[string]*common.Task
	executions  map[string]*common.Execution
	taskExecIDs map[string][]string 
	rwMutex     sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		tasks:       make(map[string]*common.Task),
		executions:  make(map[string]*common.Execution),
		taskExecIDs: make(map[string][]string),
	}
}

func (s *Store) CreateTask(req *common.CreateTaskRequest) (*common.Task, error) {
	if err := common.ValidateCronExpression(req.CronExpression); err != nil {
		return nil, ErrInvalidCronExpr
	}

	for _, depID := range req.DependsOn {
		if _, exists := s.tasks[depID]; !exists {
			return nil, ErrDependencyNotFound
		}
	}

	taskID := common.GenerateTaskID()
	now := time.Now()

	nextRunTime, err := common.GetNextRunTime(req.CronExpression, now)
	if err != nil {
		return nil, err
	}

	task := &common.Task{
		ID:                 taskID,
		Name:               req.Name,
		CronExpression:     req.CronExpression,
		NextRunTime:        nextRunTime,
		Parameters:         req.Parameters,
		Status:             common.TaskStatusActive,
		RetryCount:         0,
		ConsecutiveFailures: 0,
		CreatedAt:          now,
		UpdatedAt:          now,
		DependsOn:          req.DependsOn,
	}

	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	s.tasks[taskID] = task
	s.taskExecIDs[taskID] = make([]string, 0)

	return task, nil
}

func (s *Store) GetTask(id string) (*common.Task, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *Store) ListTasks(status *common.TaskStatus) []*common.Task {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	result := make([]*common.Task, 0)
	for _, task := range s.tasks {
		if status == nil || task.Status == *status {
			result = append(result, task)
		}
	}
	return result
}

func (s *Store) UpdateTask(req *common.UpdateTaskRequest) (*common.Task, error) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[req.ID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	if req.CronExpression != nil {
		if err := common.ValidateCronExpression(*req.CronExpression); err != nil {
			return nil, ErrInvalidCronExpr
		}
		task.CronExpression = *req.CronExpression
		
		nextRunTime, err := common.GetNextRunTime(*req.CronExpression, time.Now())
		if err == nil {
			task.NextRunTime = nextRunTime
		}
	}

	if req.Name != nil {
		task.Name = *req.Name
	}

	if req.Parameters != nil {
		task.Parameters = req.Parameters
	}

	task.UpdatedAt = time.Now()

	return task, nil
}

func (s *Store) DeleteTask(id string) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = common.TaskStatusDeleted
	task.UpdatedAt = time.Now()

	return nil
}

func (s *Store) PauseTask(id string) (*common.Task, error) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	if task.Status == common.TaskStatusActive {
		now := time.Now()
		task.Status = common.TaskStatusPaused
		task.PausedAt = common.TimePtr(now)
		task.UpdatedAt = now
	}

	return task, nil
}

func (s *Store) ResumeTask(id string) (*common.Task, error) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	if task.Status == common.TaskStatusPaused || task.Status == common.TaskStatusError {
		now := time.Now()
		task.Status = common.TaskStatusActive
		task.PausedAt = nil
		task.ConsecutiveFailures = 0
		task.RetryCount = 0
		
		nextRunTime, err := common.GetNextRunTime(task.CronExpression, now)
		if err == nil {
			task.NextRunTime = nextRunTime
		}
		task.UpdatedAt = now
	}

	return task, nil
}

func (s *Store) UpdateTaskNextRunTime(id string, nextTime time.Time) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if exists {
		task.NextRunTime = nextTime
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) UpdateTaskLastRunTime(id string, lastTime time.Time) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if exists {
		task.LastRunTime = common.TimePtr(lastTime)
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) IncrementRetryCount(id string) int {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if exists {
		task.RetryCount++
		task.UpdatedAt = time.Now()
		return task.RetryCount
	}
	return 0
}

func (s *Store) IncrementConsecutiveFailures(id string) int {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if exists {
		task.ConsecutiveFailures++
		task.UpdatedAt = time.Now()
		return task.ConsecutiveFailures
	}
	return 0
}

func (s *Store) ResetConsecutiveFailures(id string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if exists {
		task.ConsecutiveFailures = 0
		task.RetryCount = 0
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) SetTaskError(id string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	task, exists := s.tasks[id]
	if exists {
		task.Status = common.TaskStatusError
		task.UpdatedAt = time.Now()
	}
}

func (s *Store) CreateExecution(taskID string, triggerType common.TriggerType) *common.Execution {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	execID := common.GenerateExecutionID()
	now := time.Now()

	execution := &common.Execution{
		ID:          execID,
		TaskID:      taskID,
		TriggerType: triggerType,
		Status:      common.ExecutionStatusPending,
		CreatedAt:   now,
	}

	s.executions[execID] = execution
	s.taskExecIDs[taskID] = append(s.taskExecIDs[taskID], execID)

	return execution
}

func (s *Store) GetExecution(id string) (*common.Execution, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	exec, exists := s.executions[id]
	if !exists {
		return nil, ErrExecutionNotFound
	}
	return exec, nil
}

func (s *Store) ListExecutions(req *common.ListExecutionsRequest) []*common.Execution {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	var execIDs []string
	if req.TaskID != nil {
		execIDs = s.taskExecIDs[*req.TaskID]
	} else {
		execIDs = make([]string, 0, len(s.executions))
		for id := range s.executions {
			execIDs = append(execIDs, id)
		}
	}

	result := make([]*common.Execution, 0)
	for _, id := range execIDs {
		exec := s.executions[id]
		if req.Status != nil && exec.Status != *req.Status {
			continue
		}
		if req.TriggerType != nil && exec.TriggerType != *req.TriggerType {
			continue
		}
		result = append(result, exec)
	}

	return result
}

func (s *Store) UpdateExecutionStart(execID string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	exec, exists := s.executions[execID]
	if exists {
		now := time.Now()
		exec.Status = common.ExecutionStatusRunning
		exec.StartTime = common.TimePtr(now)
	}
}

func (s *Store) UpdateExecutionSuccess(execID string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	exec, exists := s.executions[execID]
	if exists {
		now := time.Now()
		exec.Status = common.ExecutionStatusSuccess
		exec.EndTime = common.TimePtr(now)
		if exec.StartTime != nil {
			exec.DurationMs = common.DurationMs(*exec.StartTime, now)
		}
	}
}

func (s *Store) UpdateExecutionFailed(execID string, errorMsg string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	exec, exists := s.executions[execID]
	if exists {
		now := time.Now()
		exec.Status = common.ExecutionStatusFailed
		exec.ErrorMsg = errorMsg
		exec.EndTime = common.TimePtr(now)
		if exec.StartTime != nil {
			exec.DurationMs = common.DurationMs(*exec.StartTime, now)
		}
	}
}

func (s *Store) UpdateExecutionSkipped(execID string, reason string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	exec, exists := s.executions[execID]
	if exists {
		exec.Status = common.ExecutionStatusSkipped
		exec.ErrorMsg = reason
	}
}

func (s *Store) IncrementExecutionRetry(execID string) int {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	exec, exists := s.executions[execID]
	if exists {
		exec.RetryNumber++
		return exec.RetryNumber
	}
	return 0
}

func (s *Store) CleanupOldExecutions(months int) int {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	cutoff := time.Now().AddDate(0, -months, 0)
	count := 0

	toDelete := make([]string, 0)
	for id, exec := range s.executions {
		if exec.CreatedAt.Before(cutoff) {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		exec := s.executions[id]
		delete(s.executions, id)
		
		if execIDs, exists := s.taskExecIDs[exec.TaskID]; exists {
			newIDs := make([]string, 0)
			for _, eid := range execIDs {
				if eid != id {
					newIDs = append(newIDs, eid)
				}
			}
			s.taskExecIDs[exec.TaskID] = newIDs
		}
		count++
	}

	return count
}

func (s *Store) GetTasksToExecute(now time.Time) []*common.Task {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	result := make([]*common.Task, 0)
	for _, task := range s.tasks {
		if task.Status != common.TaskStatusActive {
			continue
		}
		if !task.NextRunTime.After(now) && !task.NextRunTime.Equal(now) {
			continue
		}
		if now.After(task.NextRunTime) || now.Equal(task.NextRunTime) {
			result = append(result, task)
		}
	}

	return result
}

func (s *Store) GetDependentTasks(taskID string) []*common.Task {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	result := make([]*common.Task, 0)
	for _, task := range s.tasks {
		if task.Status != common.TaskStatusActive {
			continue
		}
		for _, dep := range task.DependsOn {
			if dep == taskID {
				result = append(result, task)
				break
			}
		}
	}

	return result
}
