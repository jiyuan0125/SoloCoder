package server

import (
	"time"

	"go-mailing-list/pkg/protocol"
)

type Scheduler struct {
	store     *Store
	queue     *SendQueue
	alertMgr  *AlertManager
	service   *Service

	stopChan chan struct{}
	running  bool
}

func NewScheduler(store *Store, queue *SendQueue, alertMgr *AlertManager) *Scheduler {
	return &Scheduler{
		store:    store,
		queue:    queue,
		alertMgr: alertMgr,
		stopChan: make(chan struct{}),
	}
}

func (s *Scheduler) SetService(service *Service) {
	s.service = service
}

func (s *Scheduler) Start() {
	s.running = true
	go s.run()
}

func (s *Scheduler) Stop() {
	s.running = false
	close(s.stopChan)
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ticker.C:
			s.checkScheduledTasks()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) checkScheduledTasks() {
	pendingTasks := s.store.GetPendingScheduledTasks()

	for _, task := range pendingTasks {
		if task.Status != protocol.TaskStatusScheduled {
			continue
		}

		list, exists := s.store.GetMailingList(task.ListID)
		if !exists || list.IsPaused {
			continue
		}

		subscribers, err := s.store.GetActiveSubscribers(task.ListID)
		if err != nil || len(subscribers) == 0 {
			continue
		}

		if task.IsABTest {
			abTest, exists := s.store.GetABTest(task.ABTestID)
			if exists {
				go s.service.runABTest(task, abTest, subscribers)
			}
		} else {
			s.queue.AddTask(task, subscribers)
		}
	}
}
