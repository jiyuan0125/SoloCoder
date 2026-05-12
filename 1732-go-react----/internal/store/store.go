package store

import (
	"credit-system/internal/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	mu              sync.RWMutex
	courses         map[string]*model.Course
	teachers        map[string]*model.Teacher
	collegeHistories map[string][]*model.CollegeHistory
	enrollments     map[string]*model.Enrollment
	todos           map[string]*model.Todo
}

func New() *Store {
	return &Store{
		courses:         make(map[string]*model.Course),
		teachers:        make(map[string]*model.Teacher),
		collegeHistories: make(map[string][]*model.CollegeHistory),
		enrollments:     make(map[string]*model.Enrollment),
		todos:           make(map[string]*model.Todo),
	}
}

func (s *Store) CreateCourse(c *model.Course) *model.Course {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	c.EnrolledCount = 0
	s.courses[c.ID] = c
	return c
}

func (s *Store) GetCourse(id string) *model.Course {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.courses[id]
}

func (s *Store) GetAllCourses() []*model.Course {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Course, 0, len(s.courses))
	for _, c := range s.courses {
		result = append(result, c)
	}
	return result
}

func (s *Store) UpdateCourse(c *model.Course) *model.Course {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.courses[c.ID] = c
	return c
}

func (s *Store) DeleteCourse(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.courses, id)
}

func (s *Store) CreateTeacher(t *model.Teacher) *model.Teacher {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	s.teachers[t.ID] = t
	history := &model.CollegeHistory{
		ID:        uuid.New().String(),
		TeacherID: t.ID,
		College:   t.College,
		StartDate: time.Now(),
		EndDate:   nil,
	}
	s.collegeHistories[t.ID] = []*model.CollegeHistory{history}
	return t
}

func (s *Store) GetTeacher(id string) *model.Teacher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.teachers[id]
}

func (s *Store) GetAllTeachers() []*model.Teacher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Teacher, 0, len(s.teachers))
	for _, t := range s.teachers {
		result = append(result, t)
	}
	return result
}

func (s *Store) UpdateTeacher(t *model.Teacher) *model.Teacher {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teachers[t.ID] = t
	return t
}

func (s *Store) TransferTeacher(teacherID, newCollege string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	teacher := s.teachers[teacherID]
	if teacher == nil {
		return
	}
	now := time.Now()
	histories := s.collegeHistories[teacherID]
	if len(histories) > 0 {
		histories[len(histories)-1].EndDate = &now
	}
	newHistory := &model.CollegeHistory{
		ID:        uuid.New().String(),
		TeacherID: teacherID,
		College:   newCollege,
		StartDate: now,
		EndDate:   nil,
	}
	s.collegeHistories[teacherID] = append(histories, newHistory)
	teacher.College = newCollege
}

func (s *Store) GetCollegeHistories(teacherID string) []*model.CollegeHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.collegeHistories[teacherID]
}

func (s *Store) CreateEnrollment(e *model.Enrollment) *model.Enrollment {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = uuid.New().String()
	e.EnrolledAt = time.Now()
	e.UpdatedAt = e.EnrolledAt
	s.enrollments[e.ID] = e
	if course := s.courses[e.CourseID]; course != nil {
		course.EnrolledCount++
	}
	return e
}

func (s *Store) GetEnrollment(id string) *model.Enrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enrollments[id]
}

func (s *Store) GetEnrollmentsByTeacher(teacherID string) []*model.Enrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*model.Enrollment{}
	for _, e := range s.enrollments {
		if e.TeacherID == teacherID {
			result = append(result, e)
		}
	}
	return result
}

func (s *Store) GetEnrollmentsByCourse(courseID string) []*model.Enrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*model.Enrollment{}
	for _, e := range s.enrollments {
		if e.CourseID == courseID {
			result = append(result, e)
		}
	}
	return result
}

func (s *Store) FindEnrollment(courseID, teacherID string) *model.Enrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.enrollments {
		if e.CourseID == courseID && e.TeacherID == teacherID {
			return e
		}
	}
	return nil
}

func (s *Store) UpdateEnrollment(e *model.Enrollment) *model.Enrollment {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.UpdatedAt = time.Now()
	s.enrollments[e.ID] = e
	return e
}

func (s *Store) HasEnrollmentsForCourse(courseID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.enrollments {
		if e.CourseID == courseID {
			return true
		}
	}
	return false
}

func (s *Store) CreateTodo(t *model.Todo) *model.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = uuid.New().String()
	t.IsRead = false
	t.CreatedAt = time.Now()
	s.todos[t.ID] = t
	return t
}

func (s *Store) GetTodosByTeacher(teacherID string) []*model.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*model.Todo{}
	for _, t := range s.todos {
		if t.TeacherID == teacherID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) MarkTodoRead(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if todo := s.todos[id]; todo != nil {
		todo.IsRead = true
	}
}
