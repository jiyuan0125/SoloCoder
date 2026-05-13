package services

import (
	"course-schedule/database"
	"course-schedule/models"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func EnrollCourse(input *models.EnrollInput) error {
	course, err := GetCourseByID(input.CourseID)
	if err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRow(`SELECT id FROM enrollments WHERE course_id = ? AND student_name = ?`,
		input.CourseID, input.StudentName).Scan(&existingID)
	if err == nil {
		return ErrAlreadyEnrolled
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM enrollments WHERE course_id = ?`, input.CourseID).Scan(&count)
	if err != nil {
		return err
	}

	if count >= course.Capacity {
		return ErrCourseFull
	}

	_, err = tx.Exec(
		`INSERT INTO enrollments (course_id, student_name) VALUES (?, ?)`,
		input.CourseID, input.StudentName,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func DropCourse(input *models.DropInput) error {
	_, err := GetCourseByID(input.CourseID)
	if err != nil {
		return err
	}

	var enrollmentID int64
	err = database.DB.QueryRow(
		`SELECT id FROM enrollments WHERE course_id = ? AND student_name = ?`,
		input.CourseID, input.StudentName,
	).Scan(&enrollmentID)
	if err == sql.ErrNoRows {
		return ErrEnrollmentNotFound
	}
	if err != nil {
		return err
	}

	earliestStartTime, err := getCourseEarliestSchedule(input.CourseID)
	if err != nil {
		return err
	}

	if earliestStartTime != nil {
		now := database.Now()
		if now.After(earliestStartTime.Add(-24 * time.Hour)) {
			return fmt.Errorf("%w: 开课时间为 %s",
				ErrDropTooLate,
				earliestStartTime.Format("2006-01-02 15:04"))
		}
	}

	_, err = database.DB.Exec(
		`DELETE FROM enrollments WHERE id = ?`,
		enrollmentID,
	)
	return err
}

func getCourseEarliestSchedule(courseID int64) (*time.Time, error) {
	var startTimeStr sql.NullString
	err := database.DB.QueryRow(
		`SELECT start_time FROM schedules WHERE course_id = ? ORDER BY start_time ASC LIMIT 1`,
		courseID,
	).Scan(&startTimeStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !startTimeStr.Valid {
		return nil, nil
	}
	t, err := database.ParseTime(startTimeStr.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetEnrollmentsByCourse(courseID int64) ([]models.Enrollment, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, student_name, enrolled_at FROM enrollments WHERE course_id = ?`,
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		var enrolledAtStr sql.NullString
		if err := rows.Scan(&e.ID, &e.CourseID, &e.StudentName, &enrolledAtStr); err != nil {
			return nil, err
		}
		if enrolledAtStr.Valid {
			if t, err := database.ParseTime(enrolledAtStr.String); err == nil {
				e.EnrolledAt = &t
			}
		}
		enrollments = append(enrollments, e)
	}
	if enrollments == nil {
		enrollments = []models.Enrollment{}
	}
	return enrollments, nil
}

func GetEnrollmentsByStudent(studentName string) ([]models.Enrollment, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, student_name, enrolled_at FROM enrollments WHERE student_name = ?`,
		studentName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		var enrolledAtStr sql.NullString
		if err := rows.Scan(&e.ID, &e.CourseID, &e.StudentName, &enrolledAtStr); err != nil {
			return nil, err
		}
		if enrolledAtStr.Valid {
			if t, err := database.ParseTime(enrolledAtStr.String); err == nil {
				e.EnrolledAt = &t
			}
		}
		enrollments = append(enrollments, e)
	}
	if enrollments == nil {
		enrollments = []models.Enrollment{}
	}
	return enrollments, nil
}

func GetAllEnrollments() ([]models.Enrollment, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, student_name, enrolled_at FROM enrollments`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		var enrolledAtStr sql.NullString
		if err := rows.Scan(&e.ID, &e.CourseID, &e.StudentName, &enrolledAtStr); err != nil {
			return nil, err
		}
		if enrolledAtStr.Valid {
			if t, err := database.ParseTime(enrolledAtStr.String); err == nil {
				e.EnrolledAt = &t
			}
		}
		enrollments = append(enrollments, e)
	}
	if enrollments == nil {
		enrollments = []models.Enrollment{}
	}
	return enrollments, nil
}

func GetCourseEarliestStartTime(courseID int64) (*time.Time, error) {
	return getCourseEarliestSchedule(courseID)
}
