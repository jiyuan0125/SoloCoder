package scheduler

import (
	"fmt"
	"log"
	"sync"

	"cron-scheduler/executor"
	"cron-scheduler/models"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron      *cron.Cron
	tasks     map[string]*models.Task
	entries   map[string]cron.EntryID
	executor  *executor.Executor
	mu        sync.RWMutex
	parser    cron.Parser
}

func New() *Scheduler {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	return &Scheduler{
		cron:     cron.New(cron.WithParser(parser)),
		tasks:    make(map[string]*models.Task),
		entries:  make(map[string]cron.EntryID),
		executor: executor.New(),
		parser:   parser,
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) ValidateCron(expr string) error {
	_, err := s.parser.Parse(expr)
	return err
}

func (s *Scheduler) AddTask(task *models.Task) error {
	if err := s.ValidateCron(task.CronExpr); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.ID] = task

	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.runTask(task.ID)
	})
	if err != nil {
		delete(s.tasks, task.ID)
		return fmt.Errorf("failed to schedule task: %w", err)
	}

	s.entries[task.ID] = entryID
	s.updateNextRun(task.ID)

	return nil
}

func (s *Scheduler) PauseTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	if task.Status == models.TaskStatusPaused {
		return nil
	}

	entryID, exists := s.entries[id]
	if exists {
		s.cron.Remove(entryID)
		delete(s.entries, id)
	}

	task.Status = models.TaskStatusPaused
	return nil
}

func (s *Scheduler) ResumeTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	if task.Status == models.TaskStatusRunning {
		return nil
	}

	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.runTask(task.ID)
	})
	if err != nil {
		return fmt.Errorf("failed to resume task: %w", err)
	}

	s.entries[id] = entryID
	task.Status = models.TaskStatusRunning
	s.updateNextRun(id)

	return nil
}

func (s *Scheduler) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	entryID, exists := s.entries[id]
	if exists {
		s.cron.Remove(entryID)
		delete(s.entries, id)
	}

	delete(s.tasks, id)
	return nil
}

func (s *Scheduler) GetTask(id string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	return task, exists
}

func (s *Scheduler) ListTasks() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *Scheduler) GetTaskLogs(id string) ([]*models.ExecutionLog, error) {
	task, exists := s.GetTask(id)
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task.GetExecutions(), nil
}

func (s *Scheduler) runTask(id string) {
	task, exists := s.GetTask(id)
	if !exists {
		return
	}

	go func() {
		log.Printf("executing task: %s (%s)", task.Name, task.ID)

		execLog := s.executor.Execute(task)
		task.AddExecution(execLog)

		log.Printf("task %s completed: %s, duration: %dms", task.Name, execLog.Status, execLog.DurationMS)

		s.mu.Lock()
		s.updateNextRun(id)
		s.mu.Unlock()
	}()
}

func (s *Scheduler) updateNextRun(id string) {
	task, exists := s.tasks[id]
	if !exists {
		return
	}

	entryID, exists := s.entries[id]
	if !exists {
		return
	}

	entry := s.cron.Entry(entryID)
	task.SetNextRun(entry.Next)
}
