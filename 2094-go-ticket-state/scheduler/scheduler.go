package scheduler

import (
	"context"
	"log"
	"time"

	"ticket-system/database"
	"ticket-system/models"
)

type Scheduler struct {
	db *database.DB
}

func NewScheduler(db *database.DB) *Scheduler {
	return &Scheduler{db: db}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("[scheduler] started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[scheduler] stopped")
			return
		case <-ticker.C:
			s.ProcessAutoResolve(ctx)
			s.ProcessAutoClose(ctx)
		}
	}
}

func (s *Scheduler) ProcessAutoResolve(ctx context.Context) {
	cutoff := time.Now().Add(-1 * time.Duration(models.CustomerConfirmationDays) * 24 * time.Hour)

	ids, err := s.db.GetPendingAutoResolve(ctx, cutoff)
	if err != nil {
		log.Printf("[scheduler] error getting pending auto-resolve: %v", err)
		return
	}

	for _, id := range ids {
		if err := s.db.AutoResolveTicket(ctx, id); err != nil {
			log.Printf("[scheduler] error auto-resolving ticket %d: %v", id, err)
		} else {
			log.Printf("[scheduler] auto-resolved ticket %d (customer confirmation timeout)", id)
		}
	}
}

func (s *Scheduler) ProcessAutoClose(ctx context.Context) {
	cutoff := time.Now().Add(-1 * time.Duration(models.AutoCloseDays) * 24 * time.Hour)

	ids, err := s.db.GetPendingAutoClose(ctx, cutoff)
	if err != nil {
		log.Printf("[scheduler] error getting pending auto-close: %v", err)
		return
	}

	for _, id := range ids {
		if err := s.db.AutoCloseTicket(ctx, id); err != nil {
			log.Printf("[scheduler] error auto-closing ticket %d: %v", id, err)
		} else {
			log.Printf("[scheduler] auto-closed ticket %d (resolved timeout)", id)
		}
	}
}
