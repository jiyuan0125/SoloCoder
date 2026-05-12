package services

import (
	"learning-platform/internal/models"
	"learning-platform/internal/storage"
	"learning-platform/internal/utils"
	"learning-platform/pkg/dag"
	"time"
)

type CourseService struct {
	storage *storage.Storage
}

func NewCourseService(s *storage.Storage) *CourseService {
	return &CourseService{storage: s}
}

func (s *CourseService) CreateCourse(course *models.Course) (*models.Course, error) {
	if course.ID == "" {
		course.ID = utils.GenerateUUID()
	}
	course.CreatedAt = time.Now()
	course.UpdatedAt = time.Now()

	if course.Units == nil {
		course.Units = []models.Unit{}
	}

	err := s.storage.CreateCourse(course)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (s *CourseService) GetCourse(id string) (*models.Course, bool) {
	return s.storage.GetCourse(id)
}

func (s *CourseService) GetAllCourses() []*models.Course {
	coursesMap := s.storage.GetAllCourses()
	courses := make([]*models.Course, 0, len(coursesMap))
	for _, c := range coursesMap {
		courses = append(courses, c)
	}
	return courses
}

func (s *CourseService) AddPrerequisite(courseID, prereqID string) error {
	if courseID == prereqID {
		return &storage.CycleError{Message: "course cannot be prerequisite of itself"}
	}
	return s.storage.AddPrerequisite(courseID, prereqID)
}

func (s *CourseService) GetCourseGraph() *dag.Graph {
	return s.storage.GetCourseGraph()
}

func (s *CourseService) GetTopologicalOrder() ([]string, error) {
	return s.storage.GetCourseGraph().TopologicalSort()
}

func (s *CourseService) GetCourseLevels() (map[string]int, error) {
	return s.storage.GetCourseGraph().GetLevels()
}
