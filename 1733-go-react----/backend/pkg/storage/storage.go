package storage

import (
	"tdm-system/pkg/models"
	"sync"
)

type Store struct {
	mu             sync.RWMutex
	Trainings      map[string]models.Training
	Registrations  map[string]models.Registration
	Teachers       map[string]models.Teacher
	Evaluations    map[string]models.TeachingEvaluation
	Applications   map[string]models.TitleApplication
	ReviewOpenYear int
	nextID         int
}

var store = &Store{
	Trainings:      make(map[string]models.Training),
	Registrations:  make(map[string]models.Registration),
	Teachers:       make(map[string]models.Teacher),
	Evaluations:    make(map[string]models.TeachingEvaluation),
	Applications:   make(map[string]models.TitleApplication),
	ReviewOpenYear: 2026,
	nextID:         1,
}

func Get() *Store {
	return store
}

func (s *Store) NextID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	return string(rune('A' + id%26)) + string(rune('0' + id/26%10)) + string(rune('0' + id/260%10))
}

func (s *Store) SaveTraining(t models.Training) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Trainings[t.ID] = t
}

func (s *Store) GetTraining(id string) (models.Training, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.Trainings[id]
	return t, ok
}

func (s *Store) AllTrainings() []models.Training {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Training, 0, len(s.Trainings))
	for _, t := range s.Trainings {
		result = append(result, t)
	}
	return result
}

func (s *Store) SaveRegistration(r models.Registration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Registrations[r.ID] = r
}

func (s *Store) RegistrationsByTraining(tid string) []models.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Registration, 0)
	for _, r := range s.Registrations {
		if r.TrainingID == tid {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) RegistrationsByTeacher(tid string) []models.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Registration, 0)
	for _, r := range s.Registrations {
		if r.TeacherID == tid {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) GetRegistration(id string) (models.Registration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.Registrations[id]
	return r, ok
}

func (s *Store) SaveTeacher(t models.Teacher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Teachers[t.ID] = t
}

func (s *Store) GetTeacher(id string) (models.Teacher, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.Teachers[id]
	return t, ok
}

func (s *Store) AllTeachers() []models.Teacher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Teacher, 0, len(s.Teachers))
	for _, t := range s.Teachers {
		result = append(result, t)
	}
	return result
}

func (s *Store) SaveEvaluation(e models.TeachingEvaluation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Evaluations[e.ID] = e
}

func (s *Store) GetEvaluation(id string) (models.TeachingEvaluation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.Evaluations[id]
	return e, ok
}

func (s *Store) EvaluationsByTeacher(tid string) []models.TeachingEvaluation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.TeachingEvaluation, 0)
	for _, e := range s.Evaluations {
		if e.TeacherID == tid {
			result = append(result, e)
		}
	}
	return result
}

func (s *Store) AllEvaluations() []models.TeachingEvaluation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.TeachingEvaluation, 0, len(s.Evaluations))
	for _, e := range s.Evaluations {
		result = append(result, e)
	}
	return result
}

func (s *Store) SaveApplication(a models.TitleApplication) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Applications[a.ID] = a
}

func (s *Store) GetApplication(id string) (models.TitleApplication, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.Applications[id]
	return a, ok
}

func (s *Store) ApplicationsByTeacher(tid string) []models.TitleApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.TitleApplication, 0)
	for _, a := range s.Applications {
		if a.TeacherID == tid {
			result = append(result, a)
		}
	}
	return result
}

func (s *Store) AllApplications() []models.TitleApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.TitleApplication, 0, len(s.Applications))
	for _, a := range s.Applications {
		result = append(result, a)
	}
	return result
}

func (s *Store) GetReviewOpenYear() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ReviewOpenYear
}
