package core

import (
	"firemanagement/pkg/api"
	"time"
)

type DrillService struct {
	store *Store
}

func NewDrillService(store *Store) *DrillService {
	return &DrillService{store: store}
}

func (s *DrillService) CreateDrillPlan(req api.CreateDrillPlanRequest) (string, error) {
	if !req.Type.Valid() {
		return "", ErrInvalidDrillType
	}

	plan := api.DrillPlan{
		Type:             req.Type,
		Name:             req.Name,
		ScheduledTime:    req.ScheduledTime,
		PlannedAttendees: req.PlannedAttendees,
		CreatedAt:        time.Now(),
	}
	return s.store.AddDrillPlan(plan), nil
}

func (s *DrillService) GetDrillPlan(id string) (*api.DrillPlan, error) {
	plan, exists := s.store.GetDrillPlan(id)
	if !exists {
		return nil, ErrNotFound
	}
	return plan, nil
}

func (s *DrillService) ListDrillPlans() []api.DrillPlan {
	return s.store.GetAllDrillPlans()
}

func (s *DrillService) CompleteDrill(planID string, req api.CompleteDrillRequest) (string, error) {
	plan, exists := s.store.GetDrillPlan(planID)
	if !exists {
		return "", ErrNotFound
	}

	existing := s.store.ListDrillRecords(func(r api.DrillRecord) bool {
		return r.PlanID == planID
	})
	if len(existing) > 0 {
		return "", ErrPlanAlreadyCompleted
	}

	record := api.DrillRecord{
		PlanID:          planID,
		Plan:            plan,
		ActualAttendees: req.ActualAttendees,
		DurationMinutes: req.DurationMinutes,
		Improvements:    req.Improvements,
		CompletedAt:     time.Now(),
	}
	return s.store.AddDrillRecord(record), nil
}

func (s *DrillService) GetDrillRecord(id string) (*api.DrillRecord, error) {
	record, exists := s.store.GetDrillRecord(id)
	if !exists {
		return nil, ErrNotFound
	}
	if plan, exists := s.store.GetDrillPlan(record.PlanID); exists {
		record.Plan = plan
	}
	return record, nil
}

func (s *DrillService) ListDrillRecords() []api.DrillRecord {
	records := s.store.ListDrillRecords(nil)
	for i := range records {
		if plan, exists := s.store.GetDrillPlan(records[i].PlanID); exists {
			records[i].Plan = plan
		}
	}
	return records
}

func (s *DrillService) CheckQuarterlyDrillRequirement() (bool, error) {
	now := time.Now()
	year, quarter := getCurrentQuarter(now)
	start, end := getQuarterRange(year, quarter)

	records := s.store.ListDrillRecords(func(r api.DrillRecord) bool {
		if r.Plan == nil {
			return false
		}
		return r.Plan.Type == api.DrillTypeComprehensive &&
			!r.CompletedAt.Before(start) &&
			r.CompletedAt.Before(end)
	})

	return len(records) > 0, nil
}

func (s *DrillService) CreateQuarterlyReminder() (string, error) {
	now := time.Now()
	refID := "quarterly_" + now.Format("2006-Q1")

	if s.store.ReminderExists(refID, api.ReminderTypeQuarterlyDrill) {
		return "", nil
	}

	reminder := api.Reminder{
		Type:        api.ReminderTypeQuarterlyDrill,
		Title:       "季度综合演练提醒",
		Content:     "本季度尚未安排或完成综合演练，请尽快安排以满足季度演练要求",
		ReferenceID: refID,
		CreatedAt:   time.Now(),
		Read:        false,
	}

	return s.store.AddReminder(reminder), nil
}

func (s *DrillService) ListReminders(req api.ListRemindersRequest) []api.Reminder {
	return s.store.ListReminders(func(r api.Reminder) bool {
		if req.Type != nil && r.Type != *req.Type {
			return false
		}
		if req.Read != nil && r.Read != *req.Read {
			return false
		}
		return true
	})
}

func (s *DrillService) MarkReminderRead(id string, read bool) error {
	return s.store.UpdateReminder(id, func(r *api.Reminder) {
		r.Read = read
	})
}

func (s *DrillService) GetReminder(id string) (*api.Reminder, error) {
	reminder, exists := s.store.GetReminder(id)
	if !exists {
		return nil, ErrNotFound
	}
	return reminder, nil
}

func getCurrentQuarter(t time.Time) (int, int) {
	year := t.Year()
	quarter := (int(t.Month())-1)/3 + 1
	return year, quarter
}

func getQuarterRange(year, quarter int) (time.Time, time.Time) {
	startMonth := time.Month((quarter-1)*3 + 1)
	start := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.Local)
	endMonth := startMonth + 3
	if endMonth > 12 {
		endMonth = 1
		year++
	}
	end := time.Date(year, endMonth, 1, 0, 0, 0, 0, time.Local)
	return start, end
}
