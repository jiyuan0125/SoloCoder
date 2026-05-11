package core

import (
	"cleaning-service/internal/common"
	"fmt"
	"sync"
	"time"
)

type dailyWorkerHours struct {
	mu    sync.Mutex
	hours map[string]float64
}

func newDailyWorkerHours() *dailyWorkerHours {
	return &dailyWorkerHours{hours: make(map[string]float64)}
}

func (d *dailyWorkerHours) get(cleanerID string) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.hours[cleanerID]
}

func (d *dailyWorkerHours) tryAdd(cleanerID string, hours float64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	current := d.hours[cleanerID]
	if current+hours > 8 {
		return false
	}
	d.hours[cleanerID] = current + hours
	return true
}

func (s *Service) GenerateWeeklyTasks(weekStart, weekEnd time.Time) (*common.GenerateTasksResponse, error) {
	contracts := s.store.ListContracts()
	areas := s.store.ListServiceAreas()

	areaByID := make(map[string]*common.ServiceArea)
	for _, a := range areas {
		areaByID[a.ID] = a
	}

	areaClientNames := make(map[string]string)
	areaAddresses := make(map[string]string)
	clients := s.store.ListClients()
	for _, c := range clients {
		if c.Contract != nil {
			for _, saID := range c.Contract.ServiceAreas {
				areaClientNames[saID] = c.Name
				areaAddresses[saID] = c.Contract.ServiceAddress
			}
		}
	}

	var allTasks []*common.Task

	dayCount := 0
	for d := weekStart; !d.After(weekEnd); d = d.AddDate(0, 0, 1) {
		dayCount++
	}

	dayDates := make([]time.Time, 0, dayCount)
	for d := weekStart; !d.After(weekEnd); d = d.AddDate(0, 0, 1) {
		dayDates = append(dayDates, d)
	}

	for _, contract := range contracts {
		weekdayFn := getServiceWeekdays(contract.Frequency)

		for _, areaID := range contract.ServiceAreas {
			area := areaByID[areaID]
			if area == nil {
				continue
			}

			for _, d := range dayDates {
				if !weekdayFn(d.Weekday()) {
					continue
				}

				if d.Before(contract.StartDate) || d.After(contract.EndDate) {
					continue
				}

				cleanType := common.CleanTypeBasic
				if isDeepCleanDay(d, contract) {
					cleanType = common.CleanTypeDeep
				}

				estimatedHours := calculateEstimatedHours(area.Area, cleanType)

				task := &common.Task{
					ID:              s.store.nextID("TK"),
					TaskNo:          generateTaskNo(d),
					ClientName:      areaClientNames[areaID],
					ServiceAddress:  areaAddresses[areaID],
					ServiceAreaID:   area.ID,
					ServiceAreaName: area.Name,
					CleanType:       cleanType,
					EstimatedHours:  estimatedHours,
					Date:            d,
					Status:          common.TaskStatusPending,
					CreatedAt:       time.Now(),
				}

				allTasks = append(allTasks, task)
			}
		}
	}

	tasksByDay := make(map[string][]*common.Task)
	for _, task := range allTasks {
		dayKey := task.Date.Format("2006-01-02")
		tasksByDay[dayKey] = append(tasksByDay[dayKey], task)
	}

	totalAssigned := 0
	totalPending := 0

	for _, dayTasks := range tasksByDay {
		assigned, pending := s.assignDailyTasks(dayTasks)
		totalAssigned += assigned
		totalPending += pending
	}

	for _, task := range allTasks {
		s.store.SaveTask(task)
	}

	return &common.GenerateTasksResponse{
		TotalTasks: len(allTasks),
		Assigned:   totalAssigned,
		Pending:    totalPending,
	}, nil
}

func (s *Service) assignDailyTasks(tasks []*common.Task) (int, int) {
	workerHours := newDailyWorkerHours()
	assigned := 0
	pending := 0

	cleaners := s.store.ListCleaners()
	teams := s.store.ListTeams()
	leaves := s.store.ListLeaves()

	teamByID := make(map[string]*common.Team)
	for _, t := range teams {
		teamByID[t.ID] = t
	}

	cleanerOnLeave := make(map[string]map[string]bool)
	for _, lv := range leaves {
		if lv.Status == common.LeaveStatusApproved {
			dayKey := lv.Date.Format("2006-01-02")
			if cleanerOnLeave[lv.CleanerID] == nil {
				cleanerOnLeave[lv.CleanerID] = make(map[string]bool)
			}
			cleanerOnLeave[lv.CleanerID][dayKey] = true
		}
	}

	areaByID := make(map[string]*common.ServiceArea)
	for _, a := range s.store.ListServiceAreas() {
		areaByID[a.ID] = a
	}

	for _, task := range tasks {
		dayKey := task.Date.Format("2006-01-02")
		area := areaByID[task.ServiceAreaID]
		var zoneID string
		if area != nil {
			zoneID = area.ZoneID
		}

		var selectedCleaner *common.Cleaner

		for _, cleaner := range cleaners {
			if cleanerOnLeave[cleaner.ID] != nil && cleanerOnLeave[cleaner.ID][dayKey] {
				continue
			}

			if task.CleanType == common.CleanTypeDeep && cleaner.Skill != common.SkillDeep {
				continue
			}

			team := teamByID[cleaner.TeamID]
			if team != nil && team.ZoneID == zoneID {
				if workerHours.tryAdd(cleaner.ID, task.EstimatedHours) {
					selectedCleaner = cleaner
					break
				}
			}
		}

		if selectedCleaner == nil {
			for _, cleaner := range cleaners {
				if cleanerOnLeave[cleaner.ID] != nil && cleanerOnLeave[cleaner.ID][dayKey] {
					continue
				}

				if task.CleanType == common.CleanTypeDeep && cleaner.Skill != common.SkillDeep {
					continue
				}

				if workerHours.tryAdd(cleaner.ID, task.EstimatedHours) {
					selectedCleaner = cleaner
					break
				}
			}
		}

		if selectedCleaner != nil {
			task.CleanerID = selectedCleaner.ID
			task.CleanerName = selectedCleaner.Name
			task.Status = common.TaskStatusAssigned
			assigned++
		} else {
			task.Status = common.TaskStatusPending
			pending++
		}
	}

	return assigned, pending
}

func (s *Service) CompleteTask(taskID string) (*common.Task, error) {
	task := s.store.GetTask(taskID)
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	if task.Status != common.TaskStatusAssigned && task.Status != common.TaskStatusRecheck {
		return nil, fmt.Errorf("task is not in assigned status")
	}

	task.Status = common.TaskStatusCompleted
	s.store.SaveTask(task)

	return task, nil
}

func (s *Service) ListTasks() []*common.Task {
	return s.store.ListTasks()
}

func getServiceWeekdays(frequency common.Frequency) func(time.Weekday) bool {
	switch frequency {
	case common.FrequencyDaily:
		return func(d time.Weekday) bool {
			return d != time.Saturday && d != time.Sunday
		}
	case common.FrequencyTriWeek:
		return func(d time.Weekday) bool {
			return d == time.Monday || d == time.Wednesday || d == time.Friday
		}
	case common.FrequencyBiWeek:
		return func(d time.Weekday) bool {
			return d == time.Monday || d == time.Thursday
		}
	case common.FrequencyWeekly:
		return func(d time.Weekday) bool {
			return d == time.Monday
		}
	default:
		return func(d time.Weekday) bool { return false }
	}
}

func isDeepCleanDay(d time.Time, contract *common.Contract) bool {
	weekNum := int(d.Sub(contract.StartDate).Hours()/24/7) + 1
	return weekNum%4 == 0
}

func calculateEstimatedHours(area float64, cleanType common.CleanType) float64 {
	baseRate := 100.0
	hours := area / baseRate
	if cleanType == common.CleanTypeDeep {
		hours *= 1.5
	}
	if hours < 0.5 {
		hours = 0.5
	}
	return hours
}

func generateTaskNo(date time.Time) string {
	return fmt.Sprintf("TK%s%s", date.Format("20060102"), fmtInt(time.Now().UnixNano()%100000))
}
