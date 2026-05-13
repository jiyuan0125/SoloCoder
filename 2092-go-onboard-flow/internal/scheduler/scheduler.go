package scheduler

import (
	"log"
	"time"

	"onboard-flow/internal/db"
	"onboard-flow/internal/model"
)

type Scheduler struct {
	stop chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		stop: make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		s.checkProbationReminders()

		for {
			select {
			case <-ticker.C:
				s.checkProbationReminders()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stop)
	log.Println("Scheduler stopped")
}

func (s *Scheduler) checkProbationReminders() {
	employees, err := db.GetEmployeesForProbationReminder()
	if err != nil {
		log.Printf("Error checking probation reminders: %v", err)
		return
	}

	for _, emp := range employees {
		err = s.triggerProbationReview(emp)
		if err != nil {
			log.Printf("Error triggering probation review for employee %d: %v", emp.ID, err)
		}
	}
}

func (s *Scheduler) triggerProbationReview(emp *model.Employee) error {
	log.Printf("Triggering probation review for employee: %s (ID: %d)", emp.Name, emp.ID)

	_, err := db.GetOrCreateProbationReview(emp.ID)
	if err != nil {
		return err
	}

	err = db.UpdateEmployeeCurrentStep(emp.ID, model.StepProbationPass)
	if err != nil {
		return err
	}

	log.Printf("Probation review triggered for %s. Step changed to probation_pass", emp.Name)
	return nil
}

func (s *Scheduler) RunManualCheck() {
	log.Println("Running manual probation reminder check...")
	s.checkProbationReminders()
}
