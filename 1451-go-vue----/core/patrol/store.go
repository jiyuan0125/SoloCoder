package patrol

import (
	"sync"

	"smart-park/common"
)

const (
	ExecutionStatusInProgress = "in_progress"
	ExecutionStatusCompleted  = "completed"
)

type Store struct {
	mu          sync.RWMutex
	points      map[string]*common.PatrolPoint
	routes      map[string]*common.PatrolRoute
	tasks       map[string]*common.PatrolTask
	executions  map[string]*common.PatrolExecution
	checkIns    map[string]*common.PatrolCheckIn
	execMutex   sync.Map
}

func NewStore() *Store {
	return &Store{
		points:     make(map[string]*common.PatrolPoint),
		routes:     make(map[string]*common.PatrolRoute),
		tasks:      make(map[string]*common.PatrolTask),
		executions: make(map[string]*common.PatrolExecution),
		checkIns:   make(map[string]*common.PatrolCheckIn),
	}
}

func (s *Store) AddPoint(p *common.PatrolPoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.points[p.ID] = p
}

func (s *Store) GetPoint(id string) *common.PatrolPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.points[id]
}

func (s *Store) AddRoute(r *common.PatrolRoute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[r.ID] = r
}

func (s *Store) GetRoute(id string) *common.PatrolRoute {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routes[id]
}

func (s *Store) AddTask(t *common.PatrolTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
}

func (s *Store) GetTask(id string) *common.PatrolTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

func (s *Store) AddExecution(e *common.PatrolExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[e.ID] = e
}

func (s *Store) GetExecution(id string) *common.PatrolExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.executions[id]
}

func (s *Store) UpdateExecution(e *common.PatrolExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[e.ID] = e
}

func (s *Store) AddCheckIn(c *common.PatrolCheckIn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkIns[c.ID] = c
}

func (s *Store) GetCheckInsByExecution(executionID string) []*common.PatrolCheckIn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.PatrolCheckIn
	for _, c := range s.checkIns {
		if c.ExecutionID == executionID {
			result = append(result, c)
		}
	}
	return result
}

func (s *Store) GetExecutionLock(executionID string) *sync.Mutex {
	actual, _ := s.execMutex.LoadOrStore(executionID, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (s *Store) GetAllRoutes() []*common.PatrolRoute {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.PatrolRoute
	for _, r := range s.routes {
		result = append(result, r)
	}
	return result
}

func (s *Store) GetAllTasks() []*common.PatrolTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.PatrolTask
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}
