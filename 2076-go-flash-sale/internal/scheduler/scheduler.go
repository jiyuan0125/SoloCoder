package scheduler

import (
	"log"
	"sync"
	"time"

	"flashsale/internal/service"
)

type Scheduler struct {
	orderService  *service.OrderService
	reportService *service.ReportService
	stopCh      chan struct{}
	wg         sync.WaitGroup
	running     bool
}

func NewScheduler(
	orderService *service.OrderService,
	reportService *service.ReportService,
) *Scheduler {
	return &Scheduler{
		orderService:  orderService,
		reportService: reportService,
		stopCh:      make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true

	s.wg.Add(2)

	go s.cleanupExpiredOrders()
	go s.processEndedActivities()
}

func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false

	close(s.stopCh)
	s.wg.Wait()
}

func (s *Scheduler) cleanupExpiredOrders() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runCleanup()
		}
	}
}

func (s *Scheduler) runCleanup() {
	now := time.Now()
	cancelledOrders, err := s.orderService.CleanupExpiredOrders(now)
	if err != nil {
		log.Printf("Error cleaning up expired orders: %v", err)
		return
	}

	if len(cancelledOrders) > 0 {
		log.Printf("Cancelled %d expired orders: %v", len(cancelledOrders), cancelledOrders)
	}
}

func (s *Scheduler) processEndedActivities() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runProcessEndedActivities()
		}
	}
}

func (s *Scheduler) runProcessEndedActivities() {
	log.Println("Processing ended activities...")
}
