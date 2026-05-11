package core

import (
	"firemanagement/pkg/api"
	"time"
)

type InspectionService struct {
	store *Store
}

func NewInspectionService(store *Store) *InspectionService {
	return &InspectionService{store: store}
}

func (s *InspectionService) CreateInspectionPoint(req api.CreateInspectionPointRequest) string {
	point := api.InspectionPoint{
		Name:     req.Name,
		Location: req.Location,
	}
	return s.store.AddInspectionPoint(point)
}

func (s *InspectionService) GetInspectionPoint(id string) (*api.InspectionPoint, error) {
	point, exists := s.store.GetInspectionPoint(id)
	if !exists {
		return nil, ErrNotFound
	}
	return point, nil
}

func (s *InspectionService) ListInspectionPoints() []api.InspectionPoint {
	return s.store.GetAllInspectionPoints()
}

func (s *InspectionService) CreateInspectionRoute(req api.CreateInspectionRouteRequest) (string, error) {
	points := make([]api.InspectionPoint, 0, len(req.PointIDs))
	for _, pointID := range req.PointIDs {
		point, exists := s.store.GetInspectionPoint(pointID)
		if !exists {
			return "", ErrNotFound
		}
		points = append(points, *point)
	}

	route := api.InspectionRoute{
		Name:   req.Name,
		Points: points,
	}
	return s.store.AddInspectionRoute(route), nil
}

func (s *InspectionService) GetInspectionRoute(id string) (*api.InspectionRoute, error) {
	route, exists := s.store.GetInspectionRoute(id)
	if !exists {
		return nil, ErrNotFound
	}
	return route, nil
}

func (s *InspectionService) ListInspectionRoutes() []api.InspectionRoute {
	return s.store.GetAllInspectionRoutes()
}

func (s *InspectionService) CreateInspectionPlan(req api.CreateInspectionPlanRequest) (string, error) {
	if !req.Frequency.Valid() {
		return "", ErrInvalidFrequency
	}

	_, exists := s.store.GetInspectionRoute(req.RouteID)
	if !exists {
		return "", ErrNotFound
	}

	plan := api.InspectionPlan{
		RouteID:       req.RouteID,
		InspectorID:   req.InspectorID,
		InspectorName: req.InspectorName,
		Frequency:     req.Frequency,
		StartTime:     req.StartTime,
	}
	return s.store.AddInspectionPlan(plan), nil
}

func (s *InspectionService) GetInspectionPlan(id string) (*api.InspectionPlan, error) {
	plan, exists := s.store.GetInspectionPlan(id)
	if !exists {
		return nil, ErrNotFound
	}
	if route, exists := s.store.GetInspectionRoute(plan.RouteID); exists {
		plan.Route = route
	}
	return plan, nil
}

func (s *InspectionService) ListInspectionPlans() []api.InspectionPlan {
	plans := s.store.GetAllInspectionPlans()
	for i := range plans {
		if route, exists := s.store.GetInspectionRoute(plans[i].RouteID); exists {
			plans[i].Route = route
		}
	}
	return plans
}

func (s *InspectionService) CreateInspectionTask(planID string) (string, error) {
	plan, exists := s.store.GetInspectionPlan(planID)
	if !exists {
		return "", ErrNotFound
	}

	route, exists := s.store.GetInspectionRoute(plan.RouteID)
	if !exists {
		return "", ErrNotFound
	}

	taskPoints := make([]api.TaskPoint, 0, len(route.Points))
	for i, point := range route.Points {
		taskPoints = append(taskPoints, api.TaskPoint{
			PointID:    point.ID,
			Point:      &route.Points[i],
			OrderIndex: i,
			Checked:    false,
			Normal:     nil,
			CheckedAt:  nil,
		})
	}

	task := api.InspectionTask{
		PlanID:        plan.ID,
		RouteID:       plan.RouteID,
		Route:         route,
		InspectorID:   plan.InspectorID,
		InspectorName: plan.InspectorName,
		Status:        api.TaskStatusPending,
		CreatedAt:     time.Now(),
		CompletedAt:   nil,
		Points:        taskPoints,
	}

	return s.store.AddInspectionTask(task), nil
}

func (s *InspectionService) GetInspectionTask(id string) (*api.InspectionTask, error) {
	task, exists := s.store.GetInspectionTask(id)
	if !exists {
		return nil, ErrNotFound
	}
	if route, exists := s.store.GetInspectionRoute(task.RouteID); exists {
		task.Route = route
	}
	if plan, exists := s.store.GetInspectionPlan(task.PlanID); exists {
		task.Plan = plan
	}
	return task, nil
}

func (s *InspectionService) ListInspectionTasks() []api.InspectionTask {
	tasks := s.store.ListInspectionTasks(nil)
	for i := range tasks {
		if route, exists := s.store.GetInspectionRoute(tasks[i].RouteID); exists {
			tasks[i].Route = route
		}
	}
	return tasks
}

func (s *InspectionService) CheckTaskPoint(taskID, pointID string, normal bool) error {
	task, exists := s.store.GetInspectionTask(taskID)
	if !exists {
		return ErrNotFound
	}

	if task.Status == api.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}

	if task.Status == api.TaskStatusPending {
		task.Status = api.TaskStatusInProgress
	}

	var targetIdx int
	var found bool
	for i, p := range task.Points {
		if p.PointID == pointID {
			targetIdx = i
			found = true
			break
		}
	}
	if !found {
		return ErrNotFound
	}

	if task.Points[targetIdx].Checked {
		return ErrPointAlreadyChecked
	}

	for i := 0; i < targetIdx; i++ {
		if !task.Points[i].Checked {
			return ErrInvalidPointOrder
		}
	}

	now := time.Now()
	task.Points[targetIdx].Checked = true
	task.Points[targetIdx].Normal = &normal
	task.Points[targetIdx].CheckedAt = &now

	allChecked := true
	for _, p := range task.Points {
		if !p.Checked {
			allChecked = false
			break
		}
	}
	if allChecked {
		task.Status = api.TaskStatusCompleted
		task.CompletedAt = &now
	}

	return s.store.UpdateInspectionTask(taskID, func(t *api.InspectionTask) {
		*t = *task
	})
}

func (s *InspectionService) GenerateDailyTasks() error {
	now := time.Now()
	today := now.Format("2006-01-02")

	for _, plan := range s.store.GetAllInspectionPlans() {
		if plan.StartTime.After(now) {
			continue
		}

		var shouldGenerate bool
		switch plan.Frequency {
		case api.InspectionFrequencyDaily:
			shouldGenerate = true
		case api.InspectionFrequencyWeekly:
			shouldGenerate = now.Weekday() == time.Monday
		case api.InspectionFrequencyMonthly:
			shouldGenerate = now.Day() == 1
		}

		if !shouldGenerate {
			continue
		}

		tasks := s.store.ListInspectionTasks(func(t api.InspectionTask) bool {
			if t.PlanID != plan.ID {
				return false
			}
			return t.CreatedAt.Format("2006-01-02") == today
		})

		if len(tasks) == 0 {
			_, err := s.CreateInspectionTask(plan.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *InspectionService) GetTaskSummary(taskID string) (*api.TaskSummary, error) {
	task, err := s.GetInspectionTask(taskID)
	if err != nil {
		return nil, err
	}

	summary := &api.TaskSummary{
		TaskID:      task.ID,
		TotalPoints: len(task.Points),
	}

	for _, p := range task.Points {
		if p.Checked {
			summary.CheckedPoints++
			if p.Normal != nil && *p.Normal {
				summary.NormalPoints++
			} else if p.Normal != nil && !*p.Normal {
				var pointName, location string
				if p.Point != nil {
					pointName = p.Point.Name
					location = p.Point.Location
				}
				summary.AbnormalItems = append(summary.AbnormalItems, api.AbnormalItem{
					PointID:   p.PointID,
					PointName: pointName,
					Location:  location,
				})
			}
		}
	}

	if summary.CheckedPoints > 0 {
		summary.PassRate = float64(summary.NormalPoints) / float64(summary.CheckedPoints) * 100
	}

	return summary, nil
}
