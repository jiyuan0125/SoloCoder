package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *SQLiteStore) CreateCourse(ctx context.Context, tx Tx, course *Course) (int64, error) {
	stx := tx.(*sqliteTx).tx

	now := time.Now()
	course.CreatedAt = now
	course.UpdatedAt = now

	result, err := stx.ExecContext(ctx, `
		INSERT INTO courses (name, capacity, start_time, end_time, enroll_start, enroll_end, prerequisites, resource_type, resource_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		course.Name,
		course.Capacity,
		timeToStr(course.StartTime),
		timeToStr(course.EndTime),
		timeToStr(course.EnrollStart),
		timeToStr(course.EnrollEnd),
		encodeInt64Slice(course.Prerequisites),
		course.ResourceType,
		course.ResourceID,
		timeToStr(course.CreatedAt),
		timeToStr(course.UpdatedAt),
	)
	if err != nil {
		return 0, fmt.Errorf("create course: %w", err)
	}

	return result.LastInsertId()
}

func (s *SQLiteStore) GetCourse(ctx context.Context, id int64) (*Course, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, capacity, start_time, end_time, enroll_start, enroll_end, prerequisites, resource_type, resource_id, created_at, updated_at
		FROM courses WHERE id = ?
	`, id)

	course := &Course{}
	var prereqsStr string
	var startTime, endTime, enrollStart, enrollEnd, createdAt, updatedAt string

	err := row.Scan(
		&course.ID,
		&course.Name,
		&course.Capacity,
		&startTime,
		&endTime,
		&enrollStart,
		&enrollEnd,
		&prereqsStr,
		&course.ResourceType,
		&course.ResourceID,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get course: %w", err)
	}

	course.StartTime = strToTime(startTime)
	course.EndTime = strToTime(endTime)
	course.EnrollStart = strToTime(enrollStart)
	course.EnrollEnd = strToTime(enrollEnd)
	course.Prerequisites = decodeInt64Slice(prereqsStr)
	course.CreatedAt = strToTime(createdAt)
	course.UpdatedAt = strToTime(updatedAt)

	return course, nil
}

func (s *SQLiteStore) ListCourses(ctx context.Context) ([]*Course, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, capacity, start_time, end_time, enroll_start, enroll_end, prerequisites, resource_type, resource_id, created_at, updated_at
		FROM courses ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	var courses []*Course
	for rows.Next() {
		course := &Course{}
		var prereqsStr string
		var startTime, endTime, enrollStart, enrollEnd, createdAt, updatedAt string

		err := rows.Scan(
			&course.ID,
			&course.Name,
			&course.Capacity,
			&startTime,
			&endTime,
			&enrollStart,
			&enrollEnd,
			&prereqsStr,
			&course.ResourceType,
			&course.ResourceID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}

		course.StartTime = strToTime(startTime)
		course.EndTime = strToTime(endTime)
		course.EnrollStart = strToTime(enrollStart)
		course.EnrollEnd = strToTime(enrollEnd)
		course.Prerequisites = decodeInt64Slice(prereqsStr)
		course.CreatedAt = strToTime(createdAt)
		course.UpdatedAt = strToTime(updatedAt)

		courses = append(courses, course)
	}

	return courses, nil
}

func (s *SQLiteStore) UpdateCourse(ctx context.Context, tx Tx, course *Course) error {
	stx := tx.(*sqliteTx).tx
	course.UpdatedAt = time.Now()

	_, err := stx.ExecContext(ctx, `
		UPDATE courses SET name=?, capacity=?, start_time=?, end_time=?, enroll_start=?, enroll_end=?, prerequisites=?, resource_type=?, resource_id=?, updated_at=?
		WHERE id=?
	`,
		course.Name,
		course.Capacity,
		timeToStr(course.StartTime),
		timeToStr(course.EndTime),
		timeToStr(course.EnrollStart),
		timeToStr(course.EnrollEnd),
		encodeInt64Slice(course.Prerequisites),
		course.ResourceType,
		course.ResourceID,
		timeToStr(course.UpdatedAt),
		course.ID,
	)
	if err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	return nil
}
