package patrol

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"smart-park/common"
)

type AccessPointService interface {
	MarkAccessPointFault(apID string) error
	FixAccessPoint(apID string) error
}

type Service struct {
	store              *Store
	accessPointService AccessPointService
}

func NewService(store *Store, aps AccessPointService) *Service {
	return &Service{
		store:              store,
		accessPointService: aps,
	}
}

func (s *Service) CreatePatrolPoint(id, name string) (*common.PatrolPoint, error) {
	if id == "" || name == "" {
		return nil, errors.New("id and name are required")
	}
	p := &common.PatrolPoint{ID: id, Name: name}
	s.store.AddPoint(p)
	return p, nil
}

func (s *Service) CreatePatrolRoute(name string, pointIDs []string) (*common.PatrolRoute, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(pointIDs) == 0 {
		return nil, errors.New("at least one point required")
	}
	for _, pid := range pointIDs {
		if s.store.GetPoint(pid) == nil {
			return nil, errors.New("invalid point ID: " + pid)
		}
	}
	route := &common.PatrolRoute{
		ID:       uuid.New().String(),
		Name:     name,
		PointIDs: pointIDs,
	}
	s.store.AddRoute(route)
	return route, nil
}

func (s *Service) CreatePatrolTask(routeID, assigneeID, frequency string, startTime time.Time) (*common.PatrolTask, error) {
	if route := s.store.GetRoute(routeID); route == nil {
		return nil, errors.New("route not found")
	}
	if assigneeID == "" {
		return nil, errors.New("assignee is required")
	}
	task := &common.PatrolTask{
		ID:         uuid.New().String(),
		RouteID:    routeID,
		AssigneeID: assigneeID,
		Frequency:  frequency,
		StartTime:  startTime,
	}
	s.store.AddTask(task)
	return task, nil
}

func (s *Service) StartPatrolExecution(taskID string) (*common.PatrolExecution, error) {
	task := s.store.GetTask(taskID)
	if task == nil {
		return nil, errors.New("task not found")
	}
	execution := &common.PatrolExecution{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		StartTime: time.Now(),
		Status:    ExecutionStatusInProgress,
	}
	s.store.AddExecution(execution)
	return execution, nil
}

func (s *Service) CheckIn(
	executionID, pointID string,
	hasAnomaly bool,
	anomalyDesc, relatedAccessPointID string,
) (*common.PatrolCheckIn, error) {
	mu := s.store.GetExecutionLock(executionID)
	mu.Lock()
	defer mu.Unlock()

	execution := s.store.GetExecution(executionID)
	if execution == nil {
		return nil, errors.New("execution not found")
	}
	if execution.Status != ExecutionStatusInProgress {
		return nil, errors.New("execution already completed")
	}

	task := s.store.GetTask(execution.TaskID)
	if task == nil {
		return nil, errors.New("task not found")
	}

	route := s.store.GetRoute(task.RouteID)
	if route == nil {
		return nil, errors.New("route not found")
	}

	currentIdx := -1
	for i, pid := range route.PointIDs {
		if pid == pointID {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 {
		return nil, errors.New("point not in route")
	}

	existingCheckIns := s.store.GetCheckInsByExecution(executionID)
	checkedPointCount := len(existingCheckIns)

	if currentIdx != checkedPointCount {
		return nil, errors.New("must check in points in order")
	}

	checkIn := &common.PatrolCheckIn{
		ID:                   uuid.New().String(),
		ExecutionID:          executionID,
		PointID:              pointID,
		CheckInTime:          time.Now(),
		HasAnomaly:           hasAnomaly,
		AnomalyDesc:          anomalyDesc,
		RelatedAccessPointID: relatedAccessPointID,
	}
	s.store.AddCheckIn(checkIn)

	if hasAnomaly && relatedAccessPointID != "" {
		if err := s.accessPointService.MarkAccessPointFault(relatedAccessPointID); err != nil {
			return checkIn, err
		}
	}

	if checkedPointCount+1 == len(route.PointIDs) {
		execution.Status = ExecutionStatusCompleted
		s.store.UpdateExecution(execution)
	}

	return checkIn, nil
}

func (s *Service) GetAllRoutes() []*common.PatrolRoute {
	return s.store.GetAllRoutes()
}

func (s *Service) GetAllTasks() []*common.PatrolTask {
	return s.store.GetAllTasks()
}

func (s *Service) GetCheckIns(executionID string) []*common.PatrolCheckIn {
	checkIns := s.store.GetCheckInsByExecution(executionID)
	sort.Slice(checkIns, func(i, j int) bool {
		return checkIns[i].CheckInTime.Before(checkIns[j].CheckInTime)
	})
	return checkIns
}
