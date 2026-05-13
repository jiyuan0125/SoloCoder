package course

import (
	"context"
	"errors"
	"fmt"
	"time"

	"course-enroll/storage"
)

var (
	ErrInvalidCapacity = errors.New("capacity must be greater than 0")
	ErrCourseNotFound  = errors.New("course not found")
	ErrInvalidTime     = errors.New("invalid time range")
)

type Manager struct {
	store storage.Store
}

func NewManager(store storage.Store) *Manager {
	return &Manager{store: store}
}

type CreateCourseRequest struct {
	Name          string    `json:"name"`
	Capacity      int       `json:"capacity"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	EnrollStart   time.Time `json:"enroll_start"`
	EnrollEnd     time.Time `json:"enroll_end"`
	Prerequisites []int64   `json:"prerequisites"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
}

type CourseInfo struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Capacity      int       `json:"capacity"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	EnrollStart   time.Time `json:"enroll_start"`
	EnrollEnd     time.Time `json:"enroll_end"`
	Prerequisites []int64   `json:"prerequisites"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (m *Manager) CreateCourse(ctx context.Context, req *CreateCourseRequest) (*CourseInfo, error) {
	if req.Capacity <= 0 {
		return nil, ErrInvalidCapacity
	}
	if !req.StartTime.Before(req.EndTime) {
		return nil, ErrInvalidTime
	}
	if !req.EnrollStart.Before(req.EnrollEnd) {
		return nil, ErrInvalidTime
	}
	if req.Prerequisites == nil {
		req.Prerequisites = []int64{}
	}

	tx, err := m.store.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	course := &storage.Course{
		Name:          req.Name,
		Capacity:      req.Capacity,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		EnrollStart:   req.EnrollStart,
		EnrollEnd:     req.EnrollEnd,
		Prerequisites: req.Prerequisites,
		ResourceType:  req.ResourceType,
		ResourceID:    req.ResourceID,
	}

	id, err := m.store.CreateCourse(ctx, tx, course)
	if err != nil {
		return nil, err
	}

	if req.ResourceType != "" && req.ResourceID != "" {
		err = m.store.LogResourceAssociation(ctx, tx, &storage.ResourceAssociation{
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
			Operation:    "create_course",
			EntityType:   "course",
			EntityID:     fmt.Sprintf("%d", id),
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return m.GetCourse(ctx, id)
}

func (m *Manager) GetCourse(ctx context.Context, id int64) (*CourseInfo, error) {
	course, err := m.store.GetCourse(ctx, id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}

	return &CourseInfo{
		ID:            course.ID,
		Name:          course.Name,
		Capacity:      course.Capacity,
		StartTime:     course.StartTime,
		EndTime:       course.EndTime,
		EnrollStart:   course.EnrollStart,
		EnrollEnd:     course.EnrollEnd,
		Prerequisites: course.Prerequisites,
		ResourceType:  course.ResourceType,
		ResourceID:    course.ResourceID,
		CreatedAt:     course.CreatedAt,
		UpdatedAt:     course.UpdatedAt,
	}, nil
}

func (m *Manager) ListCourses(ctx context.Context) ([]*CourseInfo, error) {
	courses, err := m.store.ListCourses(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*CourseInfo, len(courses))
	for i, c := range courses {
		result[i] = &CourseInfo{
			ID:            c.ID,
			Name:          c.Name,
			Capacity:      c.Capacity,
			StartTime:     c.StartTime,
			EndTime:       c.EndTime,
			EnrollStart:   c.EnrollStart,
			EnrollEnd:     c.EnrollEnd,
			Prerequisites: c.Prerequisites,
			ResourceType:  c.ResourceType,
			ResourceID:    c.ResourceID,
			CreatedAt:     c.CreatedAt,
			UpdatedAt:     c.UpdatedAt,
		}
	}
	return result, nil
}

func (m *Manager) GetEnrolledCount(ctx context.Context, courseID int64) (int, error) {
	enrollments, err := m.store.ListCourseEnrollments(ctx, courseID)
	if err != nil {
		return 0, err
	}
	return len(enrollments), nil
}
