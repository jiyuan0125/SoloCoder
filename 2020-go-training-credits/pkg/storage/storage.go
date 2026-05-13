package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"credits/pkg/models"
)

var (
	ErrCourseNotFound     = errors.New("课程不存在")
	ErrEmployeeNotFound   = errors.New("员工不存在")
	ErrCourseNameEmpty    = errors.New("课程名称不能为空")
	ErrInvalidCredits     = errors.New("学分必须大于 0")
	ErrInvalidCourseType  = errors.New("课程类型必须是 required 或 elective")
	ErrCourseAlreadyExists = errors.New("课程 ID 已存在")
	ErrEmployeeAlreadyExists = errors.New("员工 ID 已存在")
	ErrCourseExpired      = errors.New("课程已过期")
)

type Storage struct {
	mu       sync.RWMutex
	filePath string
	data     *DataStore
}

type DataStore struct {
	Courses    map[string]*models.Course    `json:"courses"`
	Employees  map[string]*models.Employee  `json:"employees"`
	CreditRecords []*models.CreditRecord     `json:"credit_records"`
}

func New(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	filePath := filepath.Join(dataDir, "data.json")

	s := &Storage{
		filePath: filePath,
		data: &DataStore{
			Courses:       make(map[string]*models.Course),
			Employees:     make(map[string]*models.Employee),
			CreditRecords: []*models.CreditRecord{},
		},
	}

	if err := s.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := s.save(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Storage) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, s.data); err != nil {
		return fmt.Errorf("解析数据文件失败: %w", err)
	}

	if s.data.Courses == nil {
		s.data.Courses = make(map[string]*models.Course)
	}
	if s.data.Employees == nil {
		s.data.Employees = make(map[string]*models.Employee)
	}
	if s.data.CreditRecords == nil {
		s.data.CreditRecords = []*models.CreditRecord{}
	}

	return nil
}

func (s *Storage) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入数据文件失败: %w", err)
	}

	return nil
}

func (s *Storage) AddCourse(course *models.Course) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if course.Name == "" {
		return ErrCourseNameEmpty
	}
	if course.Credits <= 0 {
		return ErrInvalidCredits
	}
	if course.Type != "required" && course.Type != "elective" {
		return ErrInvalidCourseType
	}

	if _, exists := s.data.Courses[course.ID]; exists {
		return ErrCourseAlreadyExists
	}

	s.data.Courses[course.ID] = course
	return s.save()
}

func (s *Storage) GetCourse(id string) (*models.Course, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	course, exists := s.data.Courses[id]
	if !exists {
		return nil, ErrCourseNotFound
	}
	return course, nil
}

func (s *Storage) ListCourses() []*models.Course {
	s.mu.RLock()
	defer s.mu.RUnlock()

	courses := make([]*models.Course, 0, len(s.data.Courses))
	for _, c := range s.data.Courses {
		courses = append(courses, c)
	}
	return courses
}

func (s *Storage) UpdateCourse(course *models.Course) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data.Courses[course.ID]; !exists {
		return ErrCourseNotFound
	}

	if course.Name == "" {
		return ErrCourseNameEmpty
	}
	if course.Credits <= 0 {
		return ErrInvalidCredits
	}
	if course.Type != "required" && course.Type != "elective" {
		return ErrInvalidCourseType
	}

	s.data.Courses[course.ID] = course
	return s.save()
}

func (s *Storage) DeleteCourse(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data.Courses[id]; !exists {
		return ErrCourseNotFound
	}

	delete(s.data.Courses, id)
	return s.save()
}

func (s *Storage) AddEmployee(employee *models.Employee) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data.Employees[employee.ID]; exists {
		return ErrEmployeeAlreadyExists
	}

	s.data.Employees[employee.ID] = employee
	return s.save()
}

func (s *Storage) GetEmployee(id string) (*models.Employee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	employee, exists := s.data.Employees[id]
	if !exists {
		return nil, ErrEmployeeNotFound
	}
	return employee, nil
}

func (s *Storage) ListEmployees() []*models.Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	employees := make([]*models.Employee, 0, len(s.data.Employees))
	for _, e := range s.data.Employees {
		employees = append(employees, e)
	}
	return employees
}

func (s *Storage) UpdateEmployee(employee *models.Employee) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data.Employees[employee.ID]; !exists {
		return ErrEmployeeNotFound
	}

	s.data.Employees[employee.ID] = employee
	return s.save()
}

func (s *Storage) DeleteEmployee(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data.Employees[id]; !exists {
		return ErrEmployeeNotFound
	}

	delete(s.data.Employees, id)
	return s.save()
}

func (s *Storage) GetEmployeesByDepartment(department string) []*models.Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	employees := []*models.Employee{}
	for _, e := range s.data.Employees {
		if e.Department == department {
			employees = append(employees, e)
		}
	}
	return employees
}

func (s *Storage) AddCreditRecord(record *models.CreditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	course, exists := s.data.Courses[record.CourseID]
	if !exists {
		return ErrCourseNotFound
	}

	if _, exists := s.data.Employees[record.EmployeeID]; !exists {
		return ErrEmployeeNotFound
	}

	s.data.CreditRecords = append(s.data.CreditRecords, record)

	employee := s.data.Employees[record.EmployeeID]
	if course.Type == "required" {
		if employee.CompletedRequiredCourses == nil {
			employee.CompletedRequiredCourses = make(map[string]bool)
		}
		employee.CompletedRequiredCourses[record.CourseID] = true
	}

	return s.save()
}

func (s *Storage) GetCreditRecordsByEmployee(employeeID string) []*models.CreditRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := []*models.CreditRecord{}
	for _, r := range s.data.CreditRecords {
		if r.EmployeeID == employeeID {
			records = append(records, r)
		}
	}
	return records
}

func (s *Storage) GetCreditRecordsByEmployeeAndYear(employeeID string, year int) []*models.CreditRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := []*models.CreditRecord{}
	for _, r := range s.data.CreditRecords {
		if r.EmployeeID == employeeID && r.CompletedAt.Year() == year {
			records = append(records, r)
		}
	}
	return records
}

func (s *Storage) GetAllCreditRecords() []*models.CreditRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]*models.CreditRecord, len(s.data.CreditRecords))
	copy(records, s.data.CreditRecords)
	return records
}
