package access

import (
	"sync"
	"time"

	"smart-park/common"
)

type Store struct {
	mu            sync.RWMutex
	employees     map[string]*common.Employee
	accessPoints  map[string]*common.AccessPoint
	areas         map[string]*common.Area
	rules         map[string]*common.AccessRule
	records       map[string]*common.AccessRecord
	lastAccess    map[string]time.Time
}

func NewStore() *Store {
	return &Store{
		employees:    make(map[string]*common.Employee),
		accessPoints: make(map[string]*common.AccessPoint),
		areas:        make(map[string]*common.Area),
		rules:        make(map[string]*common.AccessRule),
		records:      make(map[string]*common.AccessRecord),
		lastAccess:   make(map[string]time.Time),
	}
}

func (s *Store) AddEmployee(emp *common.Employee) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.employees[emp.ID] = emp
}

func (s *Store) GetEmployee(id string) *common.Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.employees[id]
}

func (s *Store) AddAccessPoint(ap *common.AccessPoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessPoints[ap.ID] = ap
}

func (s *Store) GetAccessPoint(id string) *common.AccessPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessPoints[id]
}

func (s *Store) GetAccessPointsByArea(areaID string) []*common.AccessPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.AccessPoint
	for _, ap := range s.accessPoints {
		if ap.AreaID == areaID {
			result = append(result, ap)
		}
	}
	return result
}

func (s *Store) UpdateAccessPointFault(id string, isFault bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ap, exists := s.accessPoints[id]; exists {
		ap.IsFault = isFault
		ap.IsOpenMode = isFault
	}
}

func (s *Store) AddArea(area *common.Area) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.areas[area.ID] = area
}

func (s *Store) GetArea(id string) *common.Area {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.areas[id]
}

func (s *Store) AddRule(rule *common.AccessRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.ID] = rule
}

func (s *Store) GetRulesByAccessPoint(apID string) []*common.AccessRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.AccessRule
	for _, rule := range s.rules {
		if rule.AccessPointID == apID && rule.IsActive {
			result = append(result, rule)
		}
	}
	return result
}

func (s *Store) AddRecord(record *common.AccessRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.ID] = record
	key := record.EmployeeID + "_" + record.AccessPointID
	s.lastAccess[key] = record.AccessTime
}

func (s *Store) GetLastAccessTime(employeeID, apID string) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := employeeID + "_" + apID
	t, exists := s.lastAccess[key]
	return t, exists
}

func (s *Store) GetAllRecords() []*common.AccessRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.AccessRecord
	for _, r := range s.records {
		result = append(result, r)
	}
	return result
}

func (s *Store) GetAllAccessPoints() []*common.AccessPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.AccessPoint
	for _, ap := range s.accessPoints {
		result = append(result, ap)
	}
	return result
}
