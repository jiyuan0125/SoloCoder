package schedule

import (
	"context"
	"errors"
	"fmt"

	"course-enroll/storage"
)

var ErrScheduleConflict = errors.New("schedule conflict")

type Conflict struct {
	CourseID   int64  `json:"course_id"`
	CourseName string `json:"course_name"`
}

type Checker struct {
	store storage.Store
}

func NewChecker(store storage.Store) *Checker {
	return &Checker{store: store}
}

func (c *Checker) CheckConflict(ctx context.Context, studentID int64, course *storage.Course) ([]Conflict, error) {
	enrollments, err := c.store.ListStudentEnrollments(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("list student enrollments: %w", err)
	}

	var conflicts []Conflict

	for _, enroll := range enrollments {
		if enroll.CourseID == course.ID {
			continue
		}

		enrolledCourse, err := c.store.GetCourse(ctx, enroll.CourseID)
		if err != nil {
			return nil, fmt.Errorf("get enrolled course: %w", err)
		}
		if enrolledCourse == nil {
			continue
		}

		if hasTimeOverlap(
			course.StartTime, course.EndTime,
			enrolledCourse.StartTime, enrolledCourse.EndTime,
		) {
			conflicts = append(conflicts, Conflict{
				CourseID:   enrolledCourse.ID,
				CourseName: enrolledCourse.Name,
			})
		}
	}

	return conflicts, nil
}

func hasTimeOverlap(start1, end1, start2, end2 interface{}) bool {
	s1 := toComparable(start1)
	e1 := toComparable(end1)
	s2 := toComparable(start2)
	e2 := toComparable(end2)

	return s1 < e2 && s2 < e1
}

func toComparable(v interface{}) int64 {
	switch t := v.(type) {
	case interface{ Unix() int64 }:
		return t.Unix()
	case int64:
		return t
	default:
		return 0
	}
}
