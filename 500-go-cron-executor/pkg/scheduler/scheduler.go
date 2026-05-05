package scheduler

import (
	"sync"
	"time"
	"cron-executor/pkg/model"
)

type Scheduler struct {
	tasks      map[string]*model.Task
	tasksMutex sync.RWMutex
	parser     *CronParser
	ticker     *time.Ticker
	stopChan   chan struct{}
	onTrigger  func(taskID string, isManual bool)
}

func NewScheduler(onTrigger func(taskID string, isManual bool)) *Scheduler {
	return &Scheduler{
		tasks:     make(map[string]*model.Task),
		parser:    NewCronParser(),
		stopChan:  make(chan struct{}),
		onTrigger: onTrigger,
	}
}

func (s *Scheduler) AddTask(task *model.Task) error {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	nextRun, err := s.parser.NextTime(task.CronExpr, time.Now())
	if err != nil {
		return err
	}

	task.NextRun = nextRun
	s.tasks[task.ID] = task
	return nil
}

func (s *Scheduler) RemoveTask(taskID string) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()
	delete(s.tasks, taskID)
}

func (s *Scheduler) GetTask(taskID string) *model.Task {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()
	return s.tasks[taskID]
}

func (s *Scheduler) GetTaskByName(name string) *model.Task {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()
	for _, task := range s.tasks {
		if task.Name == name {
			return task
		}
	}
	return nil
}

func (s *Scheduler) ListTasks() []*model.Task {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()
	tasks := make([]*model.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *Scheduler) UpdateNextRun(taskID string) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		nextRun, err := s.parser.NextTime(task.CronExpr, time.Now())
		if err == nil {
			task.NextRun = nextRun
		}
	}
}

func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(time.Second)
	go s.run()
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			s.checkAndTrigger()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) checkAndTrigger() {
	now := time.Now()

	s.tasksMutex.RLock()
	tasksToTrigger := make([]string, 0)

	for taskID, task := range s.tasks {
		if !task.NextRun.IsZero() && task.NextRun.Before(now) || task.NextRun.Equal(now) {
			tasksToTrigger = append(tasksToTrigger, taskID)
		}
	}
	s.tasksMutex.RUnlock()

	for _, taskID := range tasksToTrigger {
		s.UpdateNextRun(taskID)
		if s.onTrigger != nil {
			s.onTrigger(taskID, false)
		}
	}
}

func (s *Scheduler) TriggerManually(taskID string) {
	if s.onTrigger != nil {
		s.onTrigger(taskID, true)
	}
}
