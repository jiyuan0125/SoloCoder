package server

import (
	"fmt"
	"go-report-scheduler/pkg/common"
	"sync"
	"time"
)

const (
	MaxRetryCount       = 3
	RetryIntervalMinutes = 5
	DataRetentionMonths = 12
	PollIntervalSeconds = 10
)

type Scheduler struct {
	store          *Store
	running        bool
	stopChan       chan struct{}
	workerWG       sync.WaitGroup
	executingTasks map[string]bool
	execMutex      sync.Mutex
}

func NewScheduler(store *Store) *Scheduler {
	return &Scheduler{
		store:          store,
		stopChan:       make(chan struct{}),
		executingTasks: make(map[string]bool),
	}
}

func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true
	
	go s.pollLoop()
	go s.cleanupLoop()
	
	fmt.Println("Scheduler started")
}

func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false
	
	close(s.stopChan)
	s.workerWG.Wait()
	
	fmt.Println("Scheduler stopped")
}

func (s *Scheduler) pollLoop() {
	ticker := time.NewTicker(PollIntervalSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndExecuteTasks()
		}
	}
}

func (s *Scheduler) cleanupLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.cleanupOldData()
		}
	}
}

func (s *Scheduler) cleanupOldData() {
	count := s.store.CleanupOldExecutions(DataRetentionMonths)
	if count > 0 {
		fmt.Printf("Cleaned up %d old executions\n", count)
	}
}

func (s *Scheduler) checkAndExecuteTasks() {
	now := time.Now()
	tasks := s.store.GetTasksToExecute(now)

	for _, task := range tasks {
		s.execMutex.Lock()
		if s.executingTasks[task.ID] {
			s.execMutex.Unlock()
			continue
		}
		s.executingTasks[task.ID] = true
		s.execMutex.Unlock()

		nextRunTime, err := common.GetNextRunTime(task.CronExpression, now)
		if err == nil {
			s.store.UpdateTaskNextRunTime(task.ID, nextRunTime)
		}

		s.workerWG.Add(1)
		go func(t *common.Task) {
			defer s.workerWG.Done()
			defer func() {
				s.execMutex.Lock()
				delete(s.executingTasks, t.ID)
				s.execMutex.Unlock()
			}()

			s.executeTaskWithRetry(t, common.TriggerTypeScheduled)
		}(task)
	}
}

func (s *Scheduler) executeTaskWithRetry(task *common.Task, triggerType common.TriggerType) string {
	execution := s.store.CreateExecution(task.ID, triggerType)
	execID := execution.ID

	success := false
	var lastError error

	for retry := 0; retry <= MaxRetryCount; retry++ {
		if retry > 0 {
			fmt.Printf("Retrying task %s (attempt %d/%d) after %d minutes...\n", 
				task.ID, retry, MaxRetryCount, RetryIntervalMinutes)
			
			select {
			case <-s.stopChan:
				s.store.UpdateExecutionSkipped(execID, "Scheduler stopped during retry wait")
				return execID
			case <-time.After(time.Duration(RetryIntervalMinutes) * time.Minute):
			}

			s.store.IncrementExecutionRetry(execID)
		}

		err := s.executeSingleTask(task, execID)
		if err == nil {
			success = true
			lastError = nil
			break
		}
		lastError = err
	}

	if success {
		s.store.ResetConsecutiveFailures(task.ID)
		s.triggerDependentTasks(task.ID)
	} else {
		errorMsg := ""
		if lastError != nil {
			errorMsg = lastError.Error()
		}
		s.store.UpdateExecutionFailed(execID, errorMsg)
		
		consecutiveFailures := s.store.IncrementConsecutiveFailures(task.ID)
		if consecutiveFailures >= MaxRetryCount {
			s.store.SetTaskError(task.ID)
			fmt.Printf("Task %s marked as error after %d consecutive failures\n", 
				task.ID, MaxRetryCount)
		}

		s.skipDependentTasks(task.ID, fmt.Sprintf("Dependency task %s failed", task.ID))
	}

	return execID
}

func (s *Scheduler) executeSingleTask(task *common.Task, execID string) error {
	s.store.UpdateExecutionStart(execID)
	s.store.UpdateTaskLastRunTime(task.ID, time.Now())

	fmt.Printf("Executing task: %s (ID: %s)\n", task.Name, task.ID)
	
	time.Sleep(1 * time.Second)
	
	fmt.Printf("Task %s execution completed\n", task.ID)

	s.store.UpdateExecutionSuccess(execID)
	return nil
}

func (s *Scheduler) triggerDependentTasks(parentTaskID string) {
	dependentTasks := s.store.GetDependentTasks(parentTaskID)
	
	for _, depTask := range dependentTasks {
		if !s.checkAllDependenciesSucceeded(depTask) {
			continue
		}

		s.execMutex.Lock()
		if s.executingTasks[depTask.ID] {
			s.execMutex.Unlock()
			continue
		}
		s.executingTasks[depTask.ID] = true
		s.execMutex.Unlock()

		s.workerWG.Add(1)
		go func(t *common.Task) {
			defer s.workerWG.Done()
			defer func() {
				s.execMutex.Lock()
				delete(s.executingTasks, t.ID)
				s.execMutex.Unlock()
			}()

			s.executeTaskWithRetry(t, common.TriggerTypeScheduled)
		}(depTask)
	}
}

func (s *Scheduler) skipDependentTasks(parentTaskID string, reason string) {
	dependentTasks := s.store.GetDependentTasks(parentTaskID)
	
	for _, depTask := range dependentTasks {
		execution := s.store.CreateExecution(depTask.ID, common.TriggerTypeScheduled)
		s.store.UpdateExecutionSkipped(execution.ID, reason)
		
		s.skipDependentTasks(depTask.ID, fmt.Sprintf("Dependency task %s was skipped", depTask.ID))
	}
}

func (s *Scheduler) checkAllDependenciesSucceeded(task *common.Task) bool {
	if len(task.DependsOn) == 0 {
		return true
	}

	for _, depID := range task.DependsOn {
		executions := s.store.ListExecutions(&common.ListExecutionsRequest{
			TaskID: &depID,
		})

		latestSuccess := false
		for _, exec := range executions {
			if exec.Status == common.ExecutionStatusSuccess {
				latestSuccess = true
				break
			}
		}

		if !latestSuccess {
			return false
		}
	}

	return true
}

func (s *Scheduler) ManualTrigger(taskID string) (string, error) {
	task, err := s.store.GetTask(taskID)
	if err != nil {
		return "", err
	}

	if task.Status == common.TaskStatusDeleted {
		return "", fmt.Errorf("task is deleted")
	}

	execID := s.executeTaskWithRetry(task, common.TriggerTypeManual)
	return execID, nil
}
