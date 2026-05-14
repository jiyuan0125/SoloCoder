package services

import (
	"course-schedule/database"
	"course-schedule/models"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCourseNotFound        = errors.New("course not found")
	ErrClassroomConflict     = errors.New("classroom time conflict")
	ErrInstructorConflict    = errors.New("instructor time conflict")
	ErrSeriesOrderInvalid    = errors.New("series time not in order")
	ErrCourseFull            = errors.New("course is full")
	ErrEnrollmentNotFound    = errors.New("enrollment not found")
	ErrDropTooLate           = errors.New("cannot drop course within 24 hours of start time")
	ErrAlreadyEnrolled       = errors.New("already enrolled")
)

type ConflictDetail struct {
	CourseName   string
	StartTime    time.Time
	EndTime      time.Time
}

func CreateCourse(course *models.Course) (int64, error) {
	result, err := database.DB.Exec(
		`INSERT INTO courses (name, instructor, classroom, capacity, description) VALUES (?, ?, ?, ?, ?)`,
		course.Name, course.Instructor, course.Classroom, course.Capacity, course.Description,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetCourseByID(id int64) (*models.Course, error) {
	var course models.Course
	var createdAtStr string
	err := database.DB.QueryRow(
		`SELECT id, name, instructor, classroom, capacity, description, created_at 
		 FROM courses WHERE id = ?`, id,
	).Scan(&course.ID, &course.Name, &course.Instructor, &course.Classroom, 
		&course.Capacity, &course.Description, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, ErrCourseNotFound
	}
	if err != nil {
		return nil, err
	}
	if t, err := database.ParseTime(createdAtStr); err == nil {
		course.CreatedAt = &t
	}
	return &course, nil
}

func GetAllCourses() ([]models.Course, error) {
	rows, err := database.DB.Query(
		`SELECT id, name, instructor, classroom, capacity, description, created_at FROM courses`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var c models.Course
		var createdAtStr string
		if err := rows.Scan(&c.ID, &c.Name, &c.Instructor, &c.Classroom,
			&c.Capacity, &c.Description, &createdAtStr); err != nil {
			return nil, err
		}
		if t, err := database.ParseTime(createdAtStr); err == nil {
			c.CreatedAt = &t
		}
		courses = append(courses, c)
	}
	if courses == nil {
		courses = []models.Course{}
	}
	return courses, nil
}

func UpdateCourse(id int64, course *models.Course) error {
	existing, err := GetCourseByID(id)
	if err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`UPDATE courses SET name=?, instructor=?, classroom=?, capacity=?, description=? WHERE id=?`,
		course.Name, course.Instructor, course.Classroom, course.Capacity, course.Description, id,
	)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrCourseNotFound
	}

	if existing.Instructor != course.Instructor || existing.Classroom != course.Classroom ||
		existing.Name != course.Name {
		students, err := getEnrolledStudentsTx(tx, id)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("课程 '%s' 信息已更新：", existing.Name)
		if existing.Name != course.Name {
			msg += fmt.Sprintf("名称变更为 '%s'；", course.Name)
		}
		if existing.Instructor != course.Instructor {
			msg += fmt.Sprintf("讲师变更为 '%s'；", course.Instructor)
		}
		if existing.Classroom != course.Classroom {
			msg += fmt.Sprintf("教室变更为 '%s'；", course.Classroom)
		}
		for _, student := range students {
			if err := createNotificationTx(tx, id, student, msg); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func CreateSchedule(input *models.ScheduleInput) (int64, error) {
	course, err := GetCourseByID(input.CourseID)
	if err != nil {
		return 0, err
	}

	input.StartTime = input.StartTime.UTC()

	if input.SeriesID != nil && *input.SeriesID != "" {
		lastSchedule, err := getLastScheduleInSeries(input.CourseID, *input.SeriesID)
		if err != nil {
			return 0, err
		}
		if lastSchedule != nil {
			if !input.StartTime.After(lastSchedule.EndTime()) {
				return 0, ErrSeriesOrderInvalid
			}
		}
	}

	classroomConflict, err := checkClassroomConflict(course.Classroom, input.StartTime, input.DurationHours, 0)
	if err != nil {
		return 0, err
	}
	if classroomConflict != nil {
		return 0, fmt.Errorf("%w: %s (%s - %s)",
			ErrClassroomConflict,
			classroomConflict.CourseName,
			classroomConflict.StartTime.Format("2006-01-02 15:04"),
			classroomConflict.EndTime.Format("2006-01-02 15:04"))
	}

	instructorConflict, err := checkInstructorConflict(course.Instructor, input.StartTime, input.DurationHours, 0)
	if err != nil {
		return 0, err
	}
	if instructorConflict != nil {
		return 0, fmt.Errorf("%w: %s (%s - %s)",
			ErrInstructorConflict,
			instructorConflict.CourseName,
			instructorConflict.StartTime.Format("2006-01-02 15:04"),
			instructorConflict.EndTime.Format("2006-01-02 15:04"))
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO schedules (course_id, series_id, start_time, duration_hours) VALUES (?, ?, ?, ?)`,
		input.CourseID, input.SeriesID, database.FormatTime(input.StartTime), input.DurationHours,
	)
	if err != nil {
		return 0, err
	}

	scheduleID, _ := result.LastInsertId()

	students, err := getEnrolledStudentsTx(tx, input.CourseID)
	if err != nil {
		return 0, err
	}
	endTime := input.StartTime.Add(time.Duration(input.DurationHours) * time.Hour)
	msg := fmt.Sprintf("课程 '%s' 新增排期：%s 至 %s",
		course.Name,
		input.StartTime.Format("2006-01-02 15:04"),
		endTime.Format("2006-01-02 15:04"))
	for _, student := range students {
		if err := createNotificationTx(tx, input.CourseID, student, msg); err != nil {
			return 0, err
		}
	}

	return scheduleID, tx.Commit()
}

func GetScheduleByID(id int64) (*models.Schedule, error) {
	var s models.Schedule
	var startTimeStr string
	var createdAtStr sql.NullString
	err := database.DB.QueryRow(
		`SELECT id, course_id, series_id, start_time, duration_hours, created_at 
		 FROM schedules WHERE id = ?`, id,
	).Scan(&s.ID, &s.CourseID, &s.SeriesID, &startTimeStr, &s.DurationHours, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, errors.New("schedule not found")
	}
	if err != nil {
		return nil, err
	}
	if t, err := database.ParseTime(startTimeStr); err == nil {
		s.StartTime = t
	}
	if createdAtStr.Valid {
		if t, err := database.ParseTime(createdAtStr.String); err == nil {
			s.CreatedAt = &t
		}
	}
	return &s, nil
}

func GetAllSchedules() ([]models.Schedule, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, series_id, start_time, duration_hours, created_at FROM schedules ORDER BY start_time ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		var startTimeStr string
		var createdAtStr sql.NullString
		if err := rows.Scan(&s.ID, &s.CourseID, &s.SeriesID, &startTimeStr, &s.DurationHours, &createdAtStr); err != nil {
			return nil, err
		}
		if t, err := database.ParseTime(startTimeStr); err == nil {
			s.StartTime = t
		}
		if createdAtStr.Valid {
			if t, err := database.ParseTime(createdAtStr.String); err == nil {
				s.CreatedAt = &t
			}
		}
		schedules = append(schedules, s)
	}
	if schedules == nil {
		schedules = []models.Schedule{}
	}
	return schedules, nil
}

func GetSchedulesByCourse(courseID int64) ([]models.Schedule, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, series_id, start_time, duration_hours, created_at 
		 FROM schedules WHERE course_id = ? ORDER BY start_time ASC`, courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		var startTimeStr string
		var createdAtStr sql.NullString
		if err := rows.Scan(&s.ID, &s.CourseID, &s.SeriesID, &startTimeStr, &s.DurationHours, &createdAtStr); err != nil {
			return nil, err
		}
		if t, err := database.ParseTime(startTimeStr); err == nil {
			s.StartTime = t
		}
		if createdAtStr.Valid {
			if t, err := database.ParseTime(createdAtStr.String); err == nil {
				s.CreatedAt = &t
			}
		}
		schedules = append(schedules, s)
	}
	if schedules == nil {
		schedules = []models.Schedule{}
	}
	return schedules, nil
}

func getLastScheduleInSeries(courseID int64, seriesID string) (*models.Schedule, error) {
	var s models.Schedule
	var startTimeStr string
	var createdAtStr sql.NullString
	err := database.DB.QueryRow(
		`SELECT id, course_id, series_id, start_time, duration_hours, created_at 
		 FROM schedules WHERE course_id = ? AND series_id = ? 
		 ORDER BY start_time DESC LIMIT 1`,
		courseID, seriesID,
	).Scan(&s.ID, &s.CourseID, &s.SeriesID, &startTimeStr, &s.DurationHours, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if t, err := database.ParseTime(startTimeStr); err == nil {
		s.StartTime = t
	}
	if createdAtStr.Valid {
		if t, err := database.ParseTime(createdAtStr.String); err == nil {
			s.CreatedAt = &t
		}
	}
	return &s, nil
}

func checkClassroomConflict(classroom string, startTime time.Time, durationHours int, excludeScheduleID int64) (*ConflictDetail, error) {
	endTime := startTime.Add(time.Duration(durationHours) * time.Hour)

	rows, err := database.DB.Query(`
		SELECT c.name, s.start_time, s.duration_hours
		FROM schedules s
		JOIN courses c ON s.course_id = c.id
		WHERE c.classroom = ? AND s.id != ?
	`, classroom, excludeScheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var startTimeStr string
		var durHours int
		if err := rows.Scan(&name, &startTimeStr, &durHours); err != nil {
			return nil, err
		}
		sStart, err := database.ParseTime(startTimeStr)
		if err != nil {
			return nil, err
		}
		sEnd := sStart.Add(time.Duration(durHours) * time.Hour)

		if startTime.Before(sEnd.Add(15*time.Minute)) && endTime.Add(15*time.Minute).After(sStart) {
			return &ConflictDetail{
				CourseName: name,
				StartTime:  sStart,
				EndTime:    sEnd,
			}, nil
		}
	}
	return nil, nil
}

func checkInstructorConflict(instructor string, startTime time.Time, durationHours int, excludeScheduleID int64) (*ConflictDetail, error) {
	endTime := startTime.Add(time.Duration(durationHours) * time.Hour)

	rows, err := database.DB.Query(`
		SELECT c.name, s.start_time, s.duration_hours
		FROM schedules s
		JOIN courses c ON s.course_id = c.id
		WHERE c.instructor = ? AND s.id != ?
	`, instructor, excludeScheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var startTimeStr string
		var durHours int
		if err := rows.Scan(&name, &startTimeStr, &durHours); err != nil {
			return nil, err
		}
		sStart, err := database.ParseTime(startTimeStr)
		if err != nil {
			return nil, err
		}
		sEnd := sStart.Add(time.Duration(durHours) * time.Hour)

		if startTime.Before(sEnd) && endTime.After(sStart) {
			return &ConflictDetail{
				CourseName: name,
				StartTime:  sStart,
				EndTime:    sEnd,
			}, nil
		}
	}
	return nil, nil
}

func getEnrolledStudentsTx(tx *sql.Tx, courseID int64) ([]string, error) {
	rows, err := tx.Query(`SELECT student_name FROM enrollments WHERE course_id = ?`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		students = append(students, name)
	}
	return students, nil
}

func createNotificationTx(tx *sql.Tx, courseID int64, studentName, message string) error {
	_, err := tx.Exec(
		`INSERT INTO notifications (course_id, student_name, message) VALUES (?, ?, ?)`,
		courseID, studentName, message,
	)
	return err
}
