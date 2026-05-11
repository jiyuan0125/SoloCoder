package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"green-care-management/pkg/api"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrAlreadyExists   = errors.New("already exists")
	ErrCapacityExceeded = errors.New("worker capacity exceeded")
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getSkillLevelRequired(maintenanceType api.MaintenanceType, plantType api.PlantType) api.SkillLevel {
	if maintenanceType == api.MaintenancePrune && plantType == api.PlantTypeTree {
		return api.SkillSenior
	}
	if maintenanceType == api.MaintenancePrune && plantType == api.PlantTypeShrub {
		return api.SkillIntermediate
	}
	return api.SkillJunior
}

func getHourlyRate(level api.SkillLevel) float64 {
	switch level {
	case api.SkillJunior:
		return 30.0
	case api.SkillIntermediate:
		return 45.0
	case api.SkillSenior:
		return 60.0
	default:
		return 30.0
	}
}

func skillLevelMeets(workerLevel, requiredLevel api.SkillLevel) bool {
	levelOrder := map[api.SkillLevel]int{
		api.SkillJunior:       1,
		api.SkillIntermediate: 2,
		api.SkillSenior:       3,
	}
	return levelOrder[workerLevel] >= levelOrder[requiredLevel]
}

func (s *Service) CreateZone(name, description string) (*Zone, error) {
	if name == "" {
		return nil, ErrInvalidRequest
	}

	now := time.Now()
	zone := &Zone{
		ID:          generateID(),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.store.SaveZone(zone)
	return zone, nil
}

func (s *Service) GetZones() []*Zone {
	return s.store.GetZones()
}

func (s *Service) GetZone(id string) (*Zone, error) {
	zone, ok := s.store.GetZone(id)
	if !ok {
		return nil, ErrNotFound
	}
	return zone, nil
}

func (s *Service) CreatePlant(req *api.CreatePlantRequest) (*Plant, error) {
	if req.ZoneID == "" || req.Name == "" {
		return nil, ErrInvalidRequest
	}

	if _, ok := s.store.GetZone(req.ZoneID); !ok {
		return nil, ErrNotFound
	}

	now := time.Now()
	plant := &Plant{
		ID:           generateID(),
		ZoneID:       req.ZoneID,
		Name:         req.Name,
		Variety:      req.Variety,
		PlantType:    req.PlantType,
		PlantingDate: req.PlantingDate,
		Location:     req.Location,
		HealthStatus: req.HealthStatus,
		Description:  req.Description,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.store.SavePlant(plant)
	return plant, nil
}

func (s *Service) GetPlants() []*Plant {
	return s.store.GetPlants()
}

func (s *Service) GetPlantsByZone(zoneID string) []*Plant {
	return s.store.GetPlantsByZone(zoneID)
}

func (s *Service) GetPlant(id string) (*Plant, error) {
	plant, ok := s.store.GetPlant(id)
	if !ok {
		return nil, ErrNotFound
	}
	return plant, nil
}

func (s *Service) UpdatePlant(id string, req *api.UpdatePlantRequest) (*Plant, error) {
	plant, ok := s.store.GetPlant(id)
	if !ok {
		return nil, ErrNotFound
	}

	plant.Name = req.Name
	plant.Variety = req.Variety
	plant.PlantingDate = req.PlantingDate
	plant.Location = req.Location
	plant.HealthStatus = req.HealthStatus
	plant.Description = req.Description
	plant.UpdatedAt = time.Now()

	s.store.SavePlant(plant)
	return plant, nil
}

func (s *Service) DeletePlant(id string) error {
	if _, ok := s.store.GetPlant(id); !ok {
		return ErrNotFound
	}
	s.store.DeletePlant(id)
	return nil
}

func (s *Service) CreateWorker(req *api.CreateWorkerRequest) (*Worker, error) {
	if req.EmployeeID == "" || req.Name == "" {
		return nil, ErrInvalidRequest
	}

	if _, ok := s.store.GetZone(req.ZoneID); !ok {
		return nil, ErrNotFound
	}

	for _, w := range s.store.GetWorkers() {
		if w.EmployeeID == req.EmployeeID {
			return nil, ErrAlreadyExists
		}
	}

	now := time.Now()
	worker := &Worker{
		ID:         generateID(),
		EmployeeID: req.EmployeeID,
		Name:       req.Name,
		SkillLevel: req.SkillLevel,
		ZoneID:     req.ZoneID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.store.SaveWorker(worker)
	return worker, nil
}

func (s *Service) GetWorkers() []*Worker {
	return s.store.GetWorkers()
}

func (s *Service) GetWorker(id string) (*Worker, error) {
	worker, ok := s.store.GetWorker(id)
	if !ok {
		return nil, ErrNotFound
	}
	return worker, nil
}

func (s *Service) UpdateWorker(id string, req *api.UpdateWorkerRequest) (*Worker, error) {
	worker, ok := s.store.GetWorker(id)
	if !ok {
		return nil, ErrNotFound
	}

	if req.ZoneID != "" {
		if _, ok := s.store.GetZone(req.ZoneID); !ok {
			return nil, ErrNotFound
		}
		worker.ZoneID = req.ZoneID
	}

	worker.Name = req.Name
	worker.SkillLevel = req.SkillLevel
	worker.UpdatedAt = time.Now()

	s.store.SaveWorker(worker)
	return worker, nil
}

func (s *Service) CreateMaintenancePlan(req *api.CreateMaintenancePlanRequest) (*MaintenancePlan, error) {
	if req.ZoneID == "" {
		return nil, ErrInvalidRequest
	}

	if _, ok := s.store.GetZone(req.ZoneID); !ok {
		return nil, ErrNotFound
	}

	now := time.Now()
	plan := &MaintenancePlan{
		ID:        generateID(),
		ZoneID:    req.ZoneID,
		PlantType: req.PlantType,
		Items:     req.Items,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.store.SaveMaintenancePlan(plan)
	return plan, nil
}

func (s *Service) GetMaintenancePlans() []*MaintenancePlan {
	return s.store.GetMaintenancePlans()
}

func (s *Service) GetMaintenancePlan(id string) (*MaintenancePlan, error) {
	plan, ok := s.store.GetMaintenancePlan(id)
	if !ok {
		return nil, ErrNotFound
	}
	return plan, nil
}

func (s *Service) GetTasks() []*Task {
	return s.store.GetTasks()
}

func (s *Service) GetTask(id string) (*Task, error) {
	task, ok := s.store.GetTask(id)
	if !ok {
		return nil, ErrNotFound
	}
	return task, nil
}

func (s *Service) GetTasksByZone(zoneID string) []*Task {
	return s.store.GetTasksByZone(zoneID)
}

func (s *Service) GetTasksByWorker(workerID string) []*Task {
	return s.store.GetTasksByWorker(workerID)
}

func (s *Service) CreateAdhocTask(req *api.CreateAdhocTaskRequest) (*Task, error) {
	if req.ZoneID == "" || req.Title == "" {
		return nil, ErrInvalidRequest
	}

	if _, ok := s.store.GetZone(req.ZoneID); !ok {
		return nil, ErrNotFound
	}

	now := time.Now()
	task := &Task{
		ID:              generateID(),
		TaskType:        api.TaskAdhoc,
		ZoneID:          req.ZoneID,
		PlantIDs:        req.PlantIDs,
		MaintenanceType: req.MaintenanceType,
		Title:           req.Title,
		Description:     req.Description,
		Status:          api.TaskPending,
		EstimatedHours:  req.EstimatedHours,
		RejectCount:     0,
		RejectReasons:   []string{},
		Priority:        req.Priority,
		DueDate:         now.AddDate(0, 0, 3),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	workerID, err := s.assignWorker(task)
	if err != nil && err != ErrCapacityExceeded {
		return nil, err
	}
	if err == nil {
		task.AssignedWorkerID = workerID
		task.Status = api.TaskInProgress
	}

	s.store.SaveTask(task)
	return task, nil
}

func (s *Service) GenerateDailyTasks(date time.Time) ([]*Task, error) {
	var generatedTasks []*Task
	dateStr := date.Format("2006-01-02")

	for _, zone := range s.store.GetZones() {
		for _, plantType := range []api.PlantType{api.PlantTypeTree, api.PlantTypeShrub, api.PlantTypeLawn, api.PlantTypeFlower} {
			plants := s.getPlantsByTypeInZone(zone.ID, plantType)
			if len(plants) == 0 {
				continue
			}

			plans := s.store.GetPlansByZoneAndType(zone.ID, string(plantType))
			for _, plan := range plans {
				for _, item := range plan.Items {
					taskKey := s.buildTaskKey(zone.ID, plantType, item.MaintenanceType, dateStr)
					if s.taskExistsForDay(taskKey) {
						continue
					}

					interval := item.IntervalDays
					if interval <= 0 {
						interval = 7
					}

					dayOfWeek := int(date.Weekday())
					if interval == 7 && dayOfWeek != 1 {
						continue
					}
					if interval == 14 && dayOfWeek != 1 {
						continue
					}
					if interval == 30 && date.Day() != 1 {
						continue
					}
					if interval == 90 && date.Day() != 1 && (date.Month()%3 == 0) {
						continue
					}

					var plantIDs []string
					for _, p := range plants {
						plantIDs = append(plantIDs, p.ID)
					}

					now := time.Now()
					task := &Task{
						ID:              generateID(),
						TaskType:        api.TaskRegular,
						ZoneID:          zone.ID,
						PlantIDs:        plantIDs,
						MaintenanceType: item.MaintenanceType,
						Title:           s.generateTaskTitle(plantType, item.MaintenanceType),
						Description:     "Regular maintenance task",
						Status:          api.TaskPending,
						EstimatedHours:  item.EstimatedHours,
						RejectCount:     0,
						RejectReasons:   []string{},
						Priority:        1,
						DueDate:         date.AddDate(0, 0, 1),
						CreatedAt:       now,
						UpdatedAt:       now,
					}

					workerID, err := s.assignWorker(task)
					if err == nil {
						task.AssignedWorkerID = workerID
						task.Status = api.TaskInProgress
					}

					s.store.SaveTask(task)
					generatedTasks = append(generatedTasks, task)
				}
			}
		}
	}

	return generatedTasks, nil
}

func (s *Service) getPlantsByTypeInZone(zoneID string, plantType api.PlantType) []*Plant {
	var plants []*Plant
	for _, p := range s.store.GetPlantsByZone(zoneID) {
		if p.PlantType == plantType {
			plants = append(plants, p)
		}
	}
	return plants
}

func (s *Service) buildTaskKey(zoneID string, plantType api.PlantType, mType api.MaintenanceType, dateStr string) string {
	return zoneID + "-" + string(plantType) + "-" + string(mType) + "-" + dateStr
}

func (s *Service) taskExistsForDay(key string) bool {
	return false
}

func (s *Service) generateTaskTitle(plantType api.PlantType, mType api.MaintenanceType) string {
	pt := map[api.PlantType]string{
		api.PlantTypeTree:  "乔木",
		api.PlantTypeShrub: "灌木",
		api.PlantTypeLawn:  "草坪",
		api.PlantTypeFlower: "花卉",
	}
	mt := map[api.MaintenanceType]string{
		api.MaintenanceWater:   "浇水",
		api.MaintenancePrune:   "修剪",
		api.MaintenanceFertilize: "施肥",
		api.MaintenanceSpray:   "喷药",
		api.MaintenanceOther:   "其他",
	}
	return pt[plantType] + "-" + mt[mType]
}

func (s *Service) assignWorker(task *Task) (string, error) {
	var requiredSkill api.SkillLevel
	if len(task.PlantIDs) > 0 {
		if plant, ok := s.store.GetPlant(task.PlantIDs[0]); ok {
			requiredSkill = getSkillLevelRequired(task.MaintenanceType, plant.PlantType)
		}
	} else {
		requiredSkill = api.SkillJunior
	}

	zoneWorkers := s.store.GetWorkersByZone(task.ZoneID)
	for _, w := range zoneWorkers {
		if skillLevelMeets(w.SkillLevel, requiredSkill) {
			if s.checkWorkerCapacity(w.ID, task.EstimatedHours) {
				return w.ID, nil
			}
		}
	}

	for _, w := range s.store.GetWorkers() {
		if w.ZoneID == task.ZoneID {
			continue
		}
		if skillLevelMeets(w.SkillLevel, requiredSkill) {
			if s.checkWorkerCapacity(w.ID, task.EstimatedHours) {
				return w.ID, nil
			}
		}
	}

	return "", ErrCapacityExceeded
}

func (s *Service) checkWorkerCapacity(workerID string, estimatedHours float64) bool {
	const maxDailyHours = 8.0

	var assignedHours float64
	today := time.Now().Format("2006-01-02")

	for _, t := range s.store.GetTasksByWorker(workerID) {
		taskDate := t.CreatedAt.Format("2006-01-02")
		if taskDate == today {
			if t.Status == api.TaskPending || t.Status == api.TaskInProgress || t.Status == api.TaskReviewing {
				assignedHours += t.EstimatedHours
			}
		}
	}

	return (assignedHours + estimatedHours) <= maxDailyHours
}

func (s *Service) ExecuteTask(req *api.ExecuteTaskRequest) (*Task, error) {
	task, ok := s.store.GetTask(req.TaskID)
	if !ok {
		return nil, ErrNotFound
	}

	if task.Status != api.TaskInProgress {
		return nil, ErrInvalidRequest
	}

	task.ActualHours = req.ActualHours
	task.MaterialCost = req.MaterialCost
	task.WorkerNotes = req.Notes
	task.Status = api.TaskReviewing
	task.UpdatedAt = time.Now()

	s.store.SaveTask(task)
	return task, nil
}

func (s *Service) ReviewTask(req *api.ReviewTaskRequest) (*Task, error) {
	task, ok := s.store.GetTask(req.TaskID)
	if !ok {
		return nil, ErrNotFound
	}

	if task.Status != api.TaskReviewing {
		return nil, ErrInvalidRequest
	}

	if req.Approved {
		task.Status = api.TaskApproved
		task.UpdatedAt = time.Now()
		s.store.SaveTask(task)
		return task, nil
	}

	task.RejectCount++
	task.RejectReasons = append(task.RejectReasons, req.Reason)

	if task.RejectCount >= 3 {
		task.TaskType = api.TaskEmergency
		task.Status = api.TaskPending
		task.Priority = 10
		task.UpdatedAt = time.Now()
		s.store.SaveTask(task)
		return task, nil
	}

	task.Status = api.TaskRejected
	task.UpdatedAt = time.Now()

	workerID, err := s.assignWorker(task)
	if err == nil {
		task.AssignedWorkerID = workerID
		task.Status = api.TaskInProgress
	}

	s.store.SaveTask(task)
	return task, nil
}

func (s *Service) GenerateCostReport(year, month int) (*api.CostReportResponse, error) {
	report := &api.CostReportResponse{
		Year:     year,
		Month:    month,
		TotalCost: 0.0,
	}

	zoneCosts := make(map[string]*api.ZoneCostReport)

	for _, task := range s.store.GetTasks() {
		if task.Status != api.TaskApproved {
			continue
		}
		if task.CreatedAt.Year() != year || int(task.CreatedAt.Month()) != month {
			continue
		}

		zone, ok := s.store.GetZone(task.ZoneID)
		if !ok {
			continue
		}

		zoneReport, exists := zoneCosts[zone.ID]
		if !exists {
			zoneReport = &api.ZoneCostReport{
				ZoneID:       zone.ID,
				ZoneName:     zone.Name,
				MaterialCost: 0.0,
				LaborCost:    0.0,
				TaskCount:    0,
			}
			zoneCosts[zone.ID] = zoneReport
		}

		zoneReport.MaterialCost += task.MaterialCost
		zoneReport.TaskCount++

		if worker, ok := s.store.GetWorker(task.AssignedWorkerID); ok {
			rate := getHourlyRate(worker.SkillLevel)
			laborCost := rate * task.ActualHours
			zoneReport.LaborCost += laborCost
		}
	}

	var totalCost float64
	for _, zr := range zoneCosts {
		zr.TotalCost = zr.MaterialCost + zr.LaborCost
		report.ZoneReports = append(report.ZoneReports, *zr)
		totalCost += zr.TotalCost
	}
	report.TotalCost = totalCost

	return report, nil
}
