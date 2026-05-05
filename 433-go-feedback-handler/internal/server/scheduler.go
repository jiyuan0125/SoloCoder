package server

import (
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	service *Service
	tickers []*time.Ticker
	stop    chan struct{}
	wg      sync.WaitGroup
}

func NewScheduler(service *Service) *Scheduler {
	return &Scheduler{
		service: service,
		tickers: make([]*time.Ticker, 0),
		stop:    make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")

	s.wg.Add(3)

	complaintTicker := time.NewTicker(1 * time.Hour)
	s.tickers = append(s.tickers, complaintTicker)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-complaintTicker.C:
				escalated := s.service.CheckAndEscalateComplaints()
				if escalated > 0 {
					log.Printf("Scheduler: Escalated %d overdue complaints", escalated)
				}
			case <-s.stop:
				complaintTicker.Stop()
				return
			}
		}
	}()

	reminderTicker := time.NewTicker(1 * time.Hour)
	s.tickers = append(s.tickers, reminderTicker)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-reminderTicker.C:
				reminded := s.service.CheckAndSendReminders()
				if reminded > 0 {
					log.Printf("Scheduler: Sent %d reminders for stale feedbacks", reminded)
				}
			case <-s.stop:
				reminderTicker.Stop()
				return
			}
		}
	}()

	userLimitTicker := time.NewTicker(24 * time.Hour)
	s.tickers = append(s.tickers, userLimitTicker)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-userLimitTicker.C:
				restored := s.service.CheckAndRestoreUserLimits()
				if restored > 0 {
					log.Printf("Scheduler: Restored %d users from review", restored)
				}
			case <-s.stop:
				userLimitTicker.Stop()
				return
			}
		}
	}()

	log.Println("Scheduler started successfully")
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	close(s.stop)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) RunNow() {
	log.Println("Running all scheduled tasks immediately...")

	escalated := s.service.CheckAndEscalateComplaints()
	if escalated > 0 {
		log.Printf("Escalated %d overdue complaints", escalated)
	} else {
		log.Println("No overdue complaints found")
	}

	reminded := s.service.CheckAndSendReminders()
	if reminded > 0 {
		log.Printf("Sent %d reminders for stale feedbacks", reminded)
	} else {
		log.Println("No stale feedbacks found")
	}

	restored := s.service.CheckAndRestoreUserLimits()
	if restored > 0 {
		log.Printf("Restored %d users from review", restored)
	} else {
		log.Println("No users to restore from review")
	}
}
