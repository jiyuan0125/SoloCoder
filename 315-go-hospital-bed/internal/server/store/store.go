package store

import (
	"encoding/json"
	"fmt"
	"hospital-bed/internal/server/model"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	state model.SystemState
	mu    sync.RWMutex
	path  string
}

func NewStore(dataPath string) *Store {
	s := &Store{
		state: model.SystemState{
			Departments: make(map[string]*model.Department),
			Patients:    make(map[string]*model.PatientRecord),
		},
		path: dataPath,
	}
	s.load()
	return s
}

func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		return
	}

	var state model.SystemState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return
	}

	if state.Departments == nil {
		state.Departments = make(map[string]*model.Department)
	}
	if state.Patients == nil {
		state.Patients = make(map[string]*model.PatientRecord)
	}
	s.state = state
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	err = os.WriteFile(tempPath, data, 0644)
	if err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}

func (s *Store) GetDepartment(name string) *model.Department {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Departments[name]
}

func (s *Store) GetOrCreateDepartment(name string) *model.Department {
	s.mu.Lock()
	defer s.mu.Unlock()
	if dept, exists := s.state.Departments[name]; exists {
		return dept
	}
	dept := &model.Department{
		Name:  name,
		Wards: make(map[string]*model.Ward),
		Queue: []*model.QueueEntry{},
	}
	s.state.Departments[name] = dept
	return dept
}

func (s *Store) UpdateDepartment(dept *model.Department) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Departments[dept.Name] = dept
	return s.save()
}

func (s *Store) GetPatient(patientID string) *model.PatientRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Patients[patientID]
}

func (s *Store) UpdatePatient(patient *model.PatientRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Patients[patient.PatientID] = patient
	return s.save()
}

func (s *Store) DeletePatient(patientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.state.Patients, patientID)
	return s.save()
}

func (s *Store) SaveLocked() error {
	return s.save()
}

func (s *Store) SaveAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.save()
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) RLock() {
	s.mu.RLock()
}

func (s *Store) RUnlock() {
	s.mu.RUnlock()
}

func (s *Store) State() *model.SystemState {
	return &s.state
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("名称不能为空")
	}
	if len([]rune(name)) > 20 {
		return fmt.Errorf("名称不能超过20个字符")
	}
	return nil
}

func ValidateDepartmentName(name string) error {
	return validateName(name)
}

func ValidateWardName(name string) error {
	return validateName(name)
}
