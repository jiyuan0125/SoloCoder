package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"gocron/internal/task"
)

type Scheduler struct {
	cron      *cron.Cron
	manager   *task.Manager
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(mgr *task.Manager) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		manager: mgr,
		ctx:     ctx,
		cancel:  cancel,
		cron:    cron.New(cron.WithSeconds()),
	}
}

func (s *Scheduler) Start() {
	tasks := s.manager.List()
	for _, t := range tasks {
		if !t.Config.Paused {
			s.scheduleTask(t.Name, t.Config.Cron)
		}
	}
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cancel()
	<-s.cron.Stop().Done()
}

func (s *Scheduler) scheduleTask(name, cronExpr string) error {
	t, exists := s.manager.Get(name)
	if !exists {
		return nil
	}

	if t.CronID > 0 {
		s.cron.Remove(cron.EntryID(t.CronID))
		t.CronID = 0
	}

	id, err := s.cron.AddFunc(cronExpr, func() {
		if s.manager.IsPaused(name) {
			return
		}
		s.manager.Execute(s.ctx, name, false)
	})
	if err != nil {
		return err
	}

	s.manager.SetCronID(name, int(id))
	return nil
}

func (s *Scheduler) RefreshTask(name string) error {
	t, exists := s.manager.Get(name)
	if !exists {
		return nil
	}

	if t.Config.Paused {
		if t.CronID > 0 {
			s.cron.Remove(cron.EntryID(t.CronID))
			s.manager.SetCronID(name, 0)
		}
		return nil
	}

	return s.scheduleTask(name, t.Config.Cron)
}

func (s *Scheduler) AddTask(name string) error {
	return s.RefreshTask(name)
}

func (s *Scheduler) RemoveTask(name string) {
	t, exists := s.manager.Get(name)
	if !exists {
		return
	}
	if t.CronID > 0 {
		s.cron.Remove(cron.EntryID(t.CronID))
		s.manager.SetCronID(name, 0)
	}
}

func (s *Scheduler) TriggerManual(name string) error {
	return s.manager.Execute(s.ctx, name, true)
}

func (s *Scheduler) NextRun(name string) time.Time {
	t, exists := s.manager.Get(name)
	if !exists || t.CronID <= 0 {
		return time.Time{}
	}
	entry := s.cron.Entry(cron.EntryID(t.CronID))
	return entry.Next
}
