package core

import (
	"cleaning-service/internal/common"
	"sync"
)

type Store struct {
	mu sync.RWMutex

	clients      map[string]*common.Client
	contracts    map[string]*common.Contract
	serviceAreas map[string]*common.ServiceArea
	zones        map[string]*common.Zone
	teams        map[string]*common.Team
	cleaners     map[string]*common.Cleaner
	schedules    map[string]*common.WeeklySchedule
	leaves       map[string]*common.LeaveRequest
	tasks        map[string]*common.Task
	qualityChecks map[string]*common.QualityCheck
	todos        map[string]*common.TodoItem

	idCounter int64
}

func NewStore() *Store {
	return &Store{
		clients:      make(map[string]*common.Client),
		contracts:    make(map[string]*common.Contract),
		serviceAreas: make(map[string]*common.ServiceArea),
		zones:        make(map[string]*common.Zone),
		teams:        make(map[string]*common.Team),
		cleaners:     make(map[string]*common.Cleaner),
		schedules:    make(map[string]*common.WeeklySchedule),
		leaves:       make(map[string]*common.LeaveRequest),
		tasks:        make(map[string]*common.Task),
		qualityChecks: make(map[string]*common.QualityCheck),
		todos:        make(map[string]*common.TodoItem),
	}
}

func (s *Store) nextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idCounter++
	return prefix + fmtInt(s.idCounter)
}

func fmtInt(n int64) string {
	var buf [20]byte
	i := len(buf)
	for n >= 10 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	i--
	buf[i] = byte('0' + n)
	return string(buf[i:])
}

func (s *Store) SaveClient(c *common.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.ID] = c
}

func (s *Store) GetClient(id string) *common.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[id]
}

func (s *Store) ListClients() []*common.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clients := make([]*common.Client, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	return clients
}

func (s *Store) SaveContract(c *common.Contract) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contracts[c.ID] = c
}

func (s *Store) GetContract(id string) *common.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.contracts[id]
}

func (s *Store) ListContracts() []*common.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()
	contracts := make([]*common.Contract, 0, len(s.contracts))
	for _, c := range s.contracts {
		contracts = append(contracts, c)
	}
	return contracts
}

func (s *Store) SaveServiceArea(a *common.ServiceArea) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceAreas[a.ID] = a
}

func (s *Store) GetServiceArea(id string) *common.ServiceArea {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serviceAreas[id]
}

func (s *Store) ListServiceAreas() []*common.ServiceArea {
	s.mu.RLock()
	defer s.mu.RUnlock()
	areas := make([]*common.ServiceArea, 0, len(s.serviceAreas))
	for _, a := range s.serviceAreas {
		areas = append(areas, a)
	}
	return areas
}

func (s *Store) SaveZone(z *common.Zone) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zones[z.ID] = z
}

func (s *Store) GetZone(id string) *common.Zone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.zones[id]
}

func (s *Store) ListZones() []*common.Zone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	zones := make([]*common.Zone, 0, len(s.zones))
	for _, z := range s.zones {
		zones = append(zones, z)
	}
	return zones
}

func (s *Store) SaveTeam(t *common.Team) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams[t.ID] = t
}

func (s *Store) GetTeam(id string) *common.Team {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.teams[id]
}

func (s *Store) ListTeams() []*common.Team {
	s.mu.RLock()
	defer s.mu.RUnlock()
	teams := make([]*common.Team, 0, len(s.teams))
	for _, t := range s.teams {
		teams = append(teams, t)
	}
	return teams
}

func (s *Store) SaveCleaner(c *common.Cleaner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleaners[c.ID] = c
}

func (s *Store) GetCleaner(id string) *common.Cleaner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cleaners[id]
}

func (s *Store) ListCleaners() []*common.Cleaner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cleaners := make([]*common.Cleaner, 0, len(s.cleaners))
	for _, c := range s.cleaners {
		cleaners = append(cleaners, c)
	}
	return cleaners
}

func (s *Store) ListCleanersByTeam(teamID string) []*common.Cleaner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cleaners := make([]*common.Cleaner, 0)
	for _, c := range s.cleaners {
		if c.TeamID == teamID {
			cleaners = append(cleaners, c)
		}
	}
	return cleaners
}

func (s *Store) SaveSchedule(sc *common.WeeklySchedule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[sc.ID] = sc
}

func (s *Store) GetSchedule(id string) *common.WeeklySchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.schedules[id]
}

func (s *Store) GetScheduleByTeamAndWeek(teamID string, weekStart string) *common.WeeklySchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sc := range s.schedules {
		if sc.TeamID == teamID && sc.WeekStart.Format("2006-01-02") == weekStart {
			return sc
		}
	}
	return nil
}

func (s *Store) SaveLeave(l *common.LeaveRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leaves[l.ID] = l
}

func (s *Store) GetLeave(id string) *common.LeaveRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.leaves[id]
}

func (s *Store) ListLeaves() []*common.LeaveRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	leaves := make([]*common.LeaveRequest, 0, len(s.leaves))
	for _, l := range s.leaves {
		leaves = append(leaves, l)
	}
	return leaves
}

func (s *Store) SaveTask(t *common.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
}

func (s *Store) GetTask(id string) *common.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

func (s *Store) ListTasks() []*common.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*common.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (s *Store) LockTasks() {
	s.mu.Lock()
}

func (s *Store) UnlockTasks() {
	s.mu.Unlock()
}

func (s *Store) SaveQualityCheck(qc *common.QualityCheck) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.qualityChecks[qc.ID] = qc
}

func (s *Store) GetQualityCheck(id string) *common.QualityCheck {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.qualityChecks[id]
}

func (s *Store) ListQualityChecks() []*common.QualityCheck {
	s.mu.RLock()
	defer s.mu.RUnlock()
	qcs := make([]*common.QualityCheck, 0, len(s.qualityChecks))
	for _, qc := range s.qualityChecks {
		qcs = append(qcs, qc)
	}
	return qcs
}

func (s *Store) SaveTodo(t *common.TodoItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.todos[t.ID] = t
}

func (s *Store) GetTodo(id string) *common.TodoItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.todos[id]
}

func (s *Store) ListTodos() []*common.TodoItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	todos := make([]*common.TodoItem, 0, len(s.todos))
	for _, t := range s.todos {
		todos = append(todos, t)
	}
	return todos
}
