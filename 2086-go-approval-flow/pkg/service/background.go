package service

import (
	"log"
	"time"

	"approval-flow/pkg/store"
)

type BackgroundService struct {
	approvalService *ApprovalService
	interval        time.Duration
	stopCh          chan struct{}
}

func NewBackgroundService() *BackgroundService {
	return &BackgroundService{
		approvalService: NewApprovalService(),
		interval:        10 * time.Minute,
		stopCh:          make(chan struct{}),
	}
}

func (b *BackgroundService) Start() {
	ticker := time.NewTicker(b.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				b.processTimeouts()
			case <-b.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (b *BackgroundService) Stop() {
	close(b.stopCh)
}

func (b *BackgroundService) processTimeouts() {
	apps, err := store.ListPendingApplications()
	if err != nil {
		log.Printf("Failed to list pending applications: %v", err)
		return
	}

	for _, app := range apps {
		if err := b.approvalService.EscalateTimeoutApplication(app); err != nil {
			log.Printf("Failed to escalate application %s: %v", app.ID, err)
		}
	}
}

func (b *BackgroundService) RunNow() {
	b.processTimeouts()
}
