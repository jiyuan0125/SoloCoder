package models

import "time"

type Course struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name" binding:"required"`
	Instructor   string     `json:"instructor" binding:"required"`
	Classroom    string     `json:"classroom" binding:"required"`
	Capacity     int        `json:"capacity" binding:"required,min=1"`
	Description  string     `json:"description"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

type Schedule struct {
	ID           int64      `json:"id"`
	CourseID     int64      `json:"course_id" binding:"required"`
	SeriesID     *string    `json:"series_id"`
	StartTime    time.Time  `json:"start_time" binding:"required"`
	DurationHours int       `json:"duration_hours" binding:"required,min=1"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

type Enrollment struct {
	ID           int64      `json:"id"`
	CourseID     int64      `json:"course_id"`
	StudentName  string     `json:"student_name"`
	EnrolledAt   *time.Time `json:"enrolled_at,omitempty"`
}

type Notification struct {
	ID          int64      `json:"id"`
	CourseID    int64      `json:"course_id"`
	StudentName string     `json:"student_name"`
	Message     string     `json:"message"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type ConflictInfo struct {
	ConflictingCourse string `json:"conflicting_course"`
	ConflictPeriod    string `json:"conflict_period"`
}

type ScheduleInput struct {
	CourseID      int64      `json:"course_id" binding:"required"`
	SeriesID      *string    `json:"series_id"`
	StartTime     time.Time  `json:"start_time" binding:"required"`
	DurationHours int        `json:"duration_hours" binding:"required,min=1"`
}

type EnrollInput struct {
	CourseID    int64  `json:"course_id" binding:"required"`
	StudentName string `json:"student_name" binding:"required"`
}

type DropInput struct {
	CourseID    int64  `json:"course_id" binding:"required"`
	StudentName string `json:"student_name" binding:"required"`
}

func (s *Schedule) EndTime() time.Time {
	return s.StartTime.Add(time.Duration(s.DurationHours) * time.Hour)
}

func (s *Schedule) BufferEndTime() time.Time {
	return s.EndTime().Add(15 * time.Minute)
}

func (s *Schedule) BufferStartTime() time.Time {
	return s.StartTime.Add(-15 * time.Minute)
}
