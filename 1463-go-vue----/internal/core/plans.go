package core

import (
	"errors"
	"time"

	"safetymanager/internal/api"
)

func (s *Store) CreateInspectionPlan(name, zoneID string, inspectorIDs, items []string, frequency api.Frequency) (*InspectionPlan, error) {
	if name == "" {
		return nil, errors.New("plan name cannot be empty")
	}
	if zoneID == "" {
		return nil, errors.New("zone ID cannot be empty")
	}
	if len(inspectorIDs) == 0 {
		return nil, errors.New("at least one inspector required")
	}
	if len(items) == 0 {
		return nil, errors.New("at least one inspection item required")
	}
	if frequency != api.FrequencyDaily && frequency != api.FrequencyWeekly && frequency != api.FrequencyMonthly {
		return nil, errors.New("invalid frequency")
	}

	s.zoneMu.RLock()
	_, zoneExists := s.zones[zoneID]
	s.zoneMu.RUnlock()
	if !zoneExists {
		return nil, errors.New("zone not found")
	}

	plan := &InspectionPlan{
		ID:           generateID(),
		Name:         name,
		ZoneID:       zoneID,
		InspectorIDs: append([]string{}, inspectorIDs...),
		Items:        append([]string{}, items...),
		Frequency:    frequency,
		CreatedAt:    time.Now(),
	}

	s.planMu.Lock()
	s.plans[plan.ID] = plan
	s.planMu.Unlock()

	s.appendAuditLog("plan_created", "plan", plan.ID, "name="+name)

	return plan, nil
}

func (s *Store) GetInspectionPlan(id string) (*InspectionPlan, error) {
	s.planMu.RLock()
	defer s.planMu.RUnlock()

	plan, exists := s.plans[id]
	if !exists {
		return nil, errors.New("plan not found")
	}
	return plan, nil
}

func (s *Store) GetAllInspectionPlans() []*InspectionPlan {
	s.planMu.RLock()
	defer s.planMu.RUnlock()

	plans := make([]*InspectionPlan, 0, len(s.plans))
	for _, plan := range s.plans {
		plans = append(plans, plan)
	}
	return plans
}

func (s *Store) GenerateTasks() error {
	now := time.Now()

	s.planMu.RLock()
	plans := make([]*InspectionPlan, 0, len(s.plans))
	for _, plan := range s.plans {
		plans = append(plans, plan)
	}
	s.planMu.RUnlock()

	for _, plan := range plans {
		if err := s.generateTasksForPlan(plan, now); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) generateTasksForPlan(plan *InspectionPlan, now time.Time) error {
	var interval time.Duration
	switch plan.Frequency {
	case api.FrequencyDaily:
		interval = 24 * time.Hour
	case api.FrequencyWeekly:
		interval = 7 * 24 * time.Hour
	case api.FrequencyMonthly:
		interval = 30 * 24 * time.Hour
	default:
		return errors.New("invalid frequency")
	}

	zoneIDs, err := s.GetZoneAndChildren(plan.ZoneID)
	if err != nil {
		return err
	}

	for _, zoneID := range zoneIDs {
		for _, inspectorID := range plan.InspectorIDs {
			scheduledFor := now
			if !plan.CreatedAt.IsZero() {
				scheduledFor = plan.CreatedAt
				for scheduledFor.Before(now) {
					scheduledFor = scheduledFor.Add(interval)
				}
			}

			if !s.taskExistsForSchedule(plan.ID, zoneID, inspectorID, scheduledFor) {
				items := make([]InspectionTaskItem, 0, len(plan.Items))
				for i, item := range plan.Items {
					items = append(items, InspectionTaskItem{
						Index: i,
						Name:  item,
					})
				}

				task := &InspectionTask{
					ID:           generateID(),
					PlanID:       plan.ID,
					ZoneID:       zoneID,
					InspectorID:  inspectorID,
					Items:        items,
					Status:       api.TaskStatusPending,
					ScheduledFor: scheduledFor,
				}

				s.taskMu.Lock()
				s.tasks[task.ID] = task
				s.taskMu.Unlock()

				s.appendAuditLog("task_created", "task", task.ID, "plan="+plan.Name+",inspector="+inspectorID)
			}
		}
	}

	return nil
}

func (s *Store) taskExistsForSchedule(planID, zoneID, inspectorID string, scheduledFor time.Time) bool {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	for _, task := range s.tasks {
		if task.PlanID == planID &&
			task.ZoneID == zoneID &&
			task.InspectorID == inspectorID &&
			task.ScheduledFor.Equal(scheduledFor) {
			return true
		}
	}
	return false
}

func buildPlanResponse(plan *InspectionPlan, zoneName string) api.InspectionPlanResponse {
	return api.InspectionPlanResponse{
		ID:           plan.ID,
		Name:         plan.Name,
		ZoneID:       plan.ZoneID,
		ZoneName:     zoneName,
		InspectorIDs: append([]string{}, plan.InspectorIDs...),
		Items:        append([]string{}, plan.Items...),
		Frequency:    plan.Frequency,
		CreatedAt:    plan.CreatedAt.Format(time.RFC3339),
	}
}
