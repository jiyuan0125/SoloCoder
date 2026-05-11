package core

import (
	"errors"
	"time"

	"safetymanager/internal/api"
)

func (s *Store) GetInspectionTask(id string) (*InspectionTask, error) {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (s *Store) GetAllInspectionTasks() []*InspectionTask {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	tasks := make([]*InspectionTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *Store) SubmitInspection(taskID string, submittedItems []api.SubmitInspectionItem) ([]*Hazard, error) {
	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}
	if len(submittedItems) == 0 {
		return nil, errors.New("no items submitted")
	}

	s.taskMu.Lock()
	task, exists := s.tasks[taskID]
	if !exists {
		s.taskMu.Unlock()
		return nil, errors.New("task not found")
	}

	if task.Status == api.TaskStatusComplete {
		s.taskMu.Unlock()
		return nil, errors.New("task already completed")
	}

	itemMap := make(map[int]InspectionTaskItem)
	for _, item := range task.Items {
		itemMap[item.Index] = item
	}

	for _, submitted := range submittedItems {
		item, found := itemMap[submitted.Index]
		if !found {
			s.taskMu.Unlock()
			return nil, errors.New("item index out of range")
		}
		if submitted.Result != api.ItemResultPass && submitted.Result != api.ItemResultFail {
			s.taskMu.Unlock()
			return nil, errors.New("invalid item result")
		}
		if submitted.Result == api.ItemResultFail && submitted.Description == "" {
			s.taskMu.Unlock()
			return nil, errors.New("failed item requires description")
		}
		item.Result = &submitted.Result
		if submitted.Description != "" {
			item.Description = &submitted.Description
		}
		itemMap[submitted.Index] = item
	}

	if len(submittedItems) != len(task.Items) {
		s.taskMu.Unlock()
		return nil, errors.New("not all items submitted")
	}

	task.Items = make([]InspectionTaskItem, 0, len(itemMap))
	for i := 0; i < len(itemMap); i++ {
		task.Items = append(task.Items, itemMap[i])
	}

	now := time.Now()
	task.Status = api.TaskStatusComplete
	task.CompletedAt = &now
	s.taskMu.Unlock()

	s.appendAuditLog("task_completed", "task", taskID, "status=completed")

	hazards, err := s.createHazardsFromTask(task)
	if err != nil {
		return nil, err
	}

	return hazards, nil
}

func (s *Store) createHazardsFromTask(task *InspectionTask) ([]*Hazard, error) {
	hazards := make([]*Hazard, 0)

	for _, item := range task.Items {
		if item.Result == nil || *item.Result != api.ItemResultFail {
			continue
		}

		level := determineHazardLevel(item.Description)
		dueAt := calculateDueDate(level)

		respDept := determineResponsibleDept(task.ZoneID)

		hazard := &Hazard{
			ID:              generateID(),
			TaskID:          task.ID,
			ZoneID:          task.ZoneID,
			ItemName:        item.Name,
			Description:     *item.Description,
			Level:           level,
			Status:          api.HazardStatusInProgress,
			ResponsibleDept: respDept,
			DueAt:           dueAt,
			CreatedAt:       time.Now(),
		}

		s.hazardMu.Lock()
		s.hazards[hazard.ID] = hazard
		s.hazardMu.Unlock()

		hazards = append(hazards, hazard)

		s.appendAuditLog("hazard_created", "hazard", hazard.ID,
			"item="+item.Name+",level="+string(level)+",dept="+respDept)
	}

	return hazards, nil
}

func determineHazardLevel(description *string) api.HazardLevel {
	if description == nil {
		return api.HazardLevelGeneral
	}
	return api.HazardLevelGeneral
}

func calculateDueDate(level api.HazardLevel) time.Time {
	now := time.Now()
	switch level {
	case api.HazardLevelGeneral:
		return now.Add(7 * 24 * time.Hour)
	case api.HazardLevelMajor:
		return now.Add(3 * 24 * time.Hour)
	case api.HazardLevelCritical:
		return now.Add(24 * time.Hour)
	default:
		return now.Add(7 * 24 * time.Hour)
	}
}

func determineResponsibleDept(zoneID string) string {
	return "生产部门"
}

func buildTaskResponse(task *InspectionTask, planName, zoneName string) api.InspectionTaskResponse {
	items := make([]api.TaskItemResponse, 0, len(task.Items))
	for _, item := range task.Items {
		items = append(items, api.TaskItemResponse{
			Index:       item.Index,
			Name:        item.Name,
			Result:      toItemResultPtr(item.Result),
			Description: toStringPtrPtr(item.Description),
		})
	}

	return api.InspectionTaskResponse{
		ID:           task.ID,
		PlanID:       task.PlanID,
		PlanName:     planName,
		ZoneID:       task.ZoneID,
		ZoneName:     zoneName,
		InspectorID:  task.InspectorID,
		Items:        items,
		Status:       task.Status,
		ScheduledFor: task.ScheduledFor.Format(time.RFC3339),
		CompletedAt:  toTimePtrPtr(task.CompletedAt),
	}
}
