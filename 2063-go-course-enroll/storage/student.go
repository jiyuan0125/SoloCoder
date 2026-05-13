package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *SQLiteStore) CreateStudent(ctx context.Context, tx Tx, student *Student) (int64, error) {
	stx := tx.(*sqliteTx).tx

	result, err := stx.ExecContext(ctx, `
		INSERT INTO students (name) VALUES (?)
	`, student.Name)
	if err != nil {
		return 0, fmt.Errorf("create student: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) GetStudent(ctx context.Context, id int64) (*Student, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name FROM students WHERE id = ?`, id)

	student := &Student{}
	err := row.Scan(&student.ID, &student.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}
	return student, nil
}

func (s *SQLiteStore) CreateCompletedCourse(ctx context.Context, tx Tx, studentID, courseID int64) error {
	stx := tx.(*sqliteTx).tx

	_, err := stx.ExecContext(ctx, `
		INSERT OR REPLACE INTO completed_courses (student_id, course_id, completed)
		VALUES (?, ?, ?)
	`, studentID, courseID, timeToStr(time.Now()))
	if err != nil {
		return fmt.Errorf("create completed course: %w", err)
	}
	return nil
}

func (s *SQLiteStore) HasCompletedCourse(ctx context.Context, studentID, courseID int64) (bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT 1 FROM completed_courses WHERE student_id = ? AND course_id = ?
	`, studentID, courseID)

	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check completed course: %w", err)
	}
	return true, nil
}

func (s *SQLiteStore) GetCompletedCourses(ctx context.Context, studentID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT course_id FROM completed_courses WHERE student_id = ?
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("get completed courses: %w", err)
	}
	defer rows.Close()

	var courseIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan completed course: %w", err)
		}
		courseIDs = append(courseIDs, id)
	}

	return courseIDs, nil
}
