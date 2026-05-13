package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *SQLiteStore) CreateEnrollment(ctx context.Context, tx Tx, enrollment *Enrollment) (int64, error) {
	stx := tx.(*sqliteTx).tx

	enrollment.EnrolledAt = time.Now()

	result, err := stx.ExecContext(ctx, `
		INSERT INTO enrollments (course_id, student_id, status, drop_count, enrolled_at)
		VALUES (?, ?, ?, ?, ?)
	`,
		enrollment.CourseID,
		enrollment.StudentID,
		enrollment.Status,
		enrollment.DropCount,
		timeToStr(enrollment.EnrolledAt),
	)
	if err != nil {
		return 0, fmt.Errorf("create enrollment: %w", err)
	}

	return result.LastInsertId()
}

func (s *SQLiteStore) GetEnrollment(ctx context.Context, courseID, studentID int64) (*Enrollment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, course_id, student_id, status, drop_count, enrolled_at, dropped_at
		FROM enrollments WHERE course_id = ? AND student_id = ?
	`, courseID, studentID)

	enrollment := &Enrollment{}
	var enrolledAt, droppedAt sql.NullString

	err := row.Scan(
		&enrollment.ID,
		&enrollment.CourseID,
		&enrollment.StudentID,
		&enrollment.Status,
		&enrollment.DropCount,
		&enrolledAt,
		&droppedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get enrollment: %w", err)
	}

	if enrolledAt.Valid {
		enrollment.EnrolledAt = strToTime(enrolledAt.String)
	}
	if droppedAt.Valid {
		enrollment.DroppedAt = strToTime(droppedAt.String)
	}

	return enrollment, nil
}

func (s *SQLiteStore) ListStudentEnrollments(ctx context.Context, studentID int64) ([]*Enrollment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id, student_id, status, drop_count, enrolled_at, dropped_at
		FROM enrollments WHERE student_id = ? AND status = ?
	`, studentID, "enrolled")
	if err != nil {
		return nil, fmt.Errorf("list student enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*Enrollment
	for rows.Next() {
		e := &Enrollment{}
		var enrolledAt, droppedAt sql.NullString

		err := rows.Scan(
			&e.ID,
			&e.CourseID,
			&e.StudentID,
			&e.Status,
			&e.DropCount,
			&enrolledAt,
			&droppedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}

		if enrolledAt.Valid {
			e.EnrolledAt = strToTime(enrolledAt.String)
		}
		if droppedAt.Valid {
			e.DroppedAt = strToTime(droppedAt.String)
		}

		enrollments = append(enrollments, e)
	}

	return enrollments, nil
}

func (s *SQLiteStore) ListCourseEnrollments(ctx context.Context, courseID int64) ([]*Enrollment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id, student_id, status, drop_count, enrolled_at, dropped_at
		FROM enrollments WHERE course_id = ? AND status = ?
	`, courseID, "enrolled")
	if err != nil {
		return nil, fmt.Errorf("list course enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*Enrollment
	for rows.Next() {
		e := &Enrollment{}
		var enrolledAt, droppedAt sql.NullString

		err := rows.Scan(
			&e.ID,
			&e.CourseID,
			&e.StudentID,
			&e.Status,
			&e.DropCount,
			&enrolledAt,
			&droppedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}

		if enrolledAt.Valid {
			e.EnrolledAt = strToTime(enrolledAt.String)
		}
		if droppedAt.Valid {
			e.DroppedAt = strToTime(droppedAt.String)
		}

		enrollments = append(enrollments, e)
	}

	return enrollments, nil
}

func (s *SQLiteStore) UpdateEnrollment(ctx context.Context, tx Tx, enrollment *Enrollment) error {
	stx := tx.(*sqliteTx).tx

	var droppedAtPtr *string
	if !enrollment.DroppedAt.IsZero() {
		str := timeToStr(enrollment.DroppedAt)
		droppedAtPtr = &str
	}

	_, err := stx.ExecContext(ctx, `
		UPDATE enrollments SET status=?, drop_count=?, dropped_at=? WHERE id=?
	`,
		enrollment.Status,
		enrollment.DropCount,
		droppedAtPtr,
		enrollment.ID,
	)
	if err != nil {
		return fmt.Errorf("update enrollment: %w", err)
	}
	return nil
}
