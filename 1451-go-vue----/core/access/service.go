package access

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"smart-park/common"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) AddEmployee(id, name, department string) *common.Employee {
	emp := &common.Employee{
		ID:         id,
		Name:       name,
		Department: department,
	}
	s.store.AddEmployee(emp)
	return emp
}

func (s *Service) CreateAccessPoint(id, name, areaID, buildingID string) (*common.AccessPoint, error) {
	if id == "" || name == "" {
		return nil, errors.New("id and name are required")
	}
	ap := &common.AccessPoint{
		ID:         id,
		Name:       name,
		AreaID:     areaID,
		BuildingID: buildingID,
		IsFault:    false,
		IsOpenMode: false,
	}
	s.store.AddAccessPoint(ap)
	return ap, nil
}

func (s *Service) CreateArea(id, name string) (*common.Area, error) {
	if id == "" || name == "" {
		return nil, errors.New("id and name are required")
	}
	area := &common.Area{ID: id, Name: name}
	s.store.AddArea(area)
	return area, nil
}

func (s *Service) CreateAccessRule(apID string, allowedDepts []string, startTime, endTime string) (*common.AccessRule, error) {
	if ap := s.store.GetAccessPoint(apID); ap == nil {
		return nil, errors.New("access point not found")
	}
	rule := &common.AccessRule{
		ID:                 uuid.New().String(),
		AccessPointID:      apID,
		AllowedDepartments: allowedDepts,
		StartTime:          startTime,
		EndTime:            endTime,
		IsActive:           true,
	}
	s.store.AddRule(rule)
	return rule, nil
}

func (s *Service) BatchSetRulesByArea(areaID string, allowedDepts []string, startTime, endTime string) ([]*common.AccessRule, error) {
	aps := s.store.GetAccessPointsByArea(areaID)
	if len(aps) == 0 {
		return nil, errors.New("no access points found in area")
	}
	var rules []*common.AccessRule
	for _, ap := range aps {
		rule, err := s.CreateAccessRule(ap.ID, allowedDepts, startTime, endTime)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (s *Service) ProcessAccess(employeeID, apID, accessType string) (*common.AccessRecord, error) {
	employee := s.store.GetEmployee(employeeID)
	if employee == nil {
		return nil, errors.New("employee not found")
	}

	ap := s.store.GetAccessPoint(apID)
	if ap == nil {
		return nil, errors.New("access point not found")
	}

	if ap.IsFault || ap.IsOpenMode {
		return s.createRecord(employee, ap, accessType), nil
	}

	now := time.Now()
	if s.shouldDedup(employeeID, apID, now) {
		return nil, errors.New("duplicate access within 5 minutes")
	}

	if !s.checkRules(employee, ap, now) {
		return nil, errors.New("access denied")
	}

	record := s.createRecord(employee, ap, accessType)
	return record, nil
}

func (s *Service) shouldDedup(employeeID, apID string, now time.Time) bool {
	lastTime, exists := s.store.GetLastAccessTime(employeeID, apID)
	if !exists {
		return false
	}
	return now.Sub(lastTime) < 5*time.Minute
}

func (s *Service) checkRules(employee *common.Employee, ap *common.AccessPoint, now time.Time) bool {
	rules := s.store.GetRulesByAccessPoint(ap.ID)

	if len(rules) == 0 {
		hour := now.Hour()
		if hour >= 0 && hour < 6 {
			return employee.Department == "安保"
		}
		return true
	}

	for _, rule := range rules {
		if s.matchesRule(employee, rule, now) {
			return true
		}
	}
	return false
}

func (s *Service) matchesRule(employee *common.Employee, rule *common.AccessRule, now time.Time) bool {
	deptAllowed := false
	for _, dept := range rule.AllowedDepartments {
		if dept == employee.Department {
			deptAllowed = true
			break
		}
	}
	if !deptAllowed {
		return false
	}

	if rule.StartTime != "" && rule.EndTime != "" {
		nowStr := now.Format("15:04")
		if nowStr < rule.StartTime || nowStr > rule.EndTime {
			return false
		}
	}

	return true
}

func (s *Service) createRecord(employee *common.Employee, ap *common.AccessPoint, accessType string) *common.AccessRecord {
	record := &common.AccessRecord{
		ID:              uuid.New().String(),
		EmployeeID:      employee.ID,
		EmployeeName:    employee.Name,
		Department:      employee.Department,
		AccessTime:      time.Now(),
		AccessPointID:   ap.ID,
		AccessPointName: ap.Name,
		AccessType:      accessType,
	}
	s.store.AddRecord(record)
	return record
}

func (s *Service) MarkAccessPointFault(apID string) error {
	if ap := s.store.GetAccessPoint(apID); ap == nil {
		return errors.New("access point not found")
	}
	s.store.UpdateAccessPointFault(apID, true)
	return nil
}

func (s *Service) FixAccessPoint(apID string) error {
	if ap := s.store.GetAccessPoint(apID); ap == nil {
		return errors.New("access point not found")
	}
	s.store.UpdateAccessPointFault(apID, false)
	return nil
}

func (s *Service) GetAllRecords() []*common.AccessRecord {
	return s.store.GetAllRecords()
}

func (s *Service) GetAllAccessPoints() []*common.AccessPoint {
	return s.store.GetAllAccessPoints()
}

func (s *Service) GetEmployee(id string) *common.Employee {
	return s.store.GetEmployee(id)
}

func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	for strings.HasPrefix(phone, "0") {
		phone = phone[1:]
	}
	return phone
}
