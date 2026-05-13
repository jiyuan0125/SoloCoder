package services

import (
	"course-schedule/database"
	"course-schedule/models"
	"database/sql"
)

func CreateNotification(courseID int64, studentName, message string) error {
	_, err := database.DB.Exec(
		`INSERT INTO notifications (course_id, student_name, message) VALUES (?, ?, ?)`,
		courseID, studentName, message,
	)
	return err
}

func GetNotificationsByStudent(studentName string) ([]models.Notification, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, student_name, message, created_at 
		 FROM notifications WHERE student_name = ? ORDER BY created_at DESC`,
		studentName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var createdAtStr sql.NullString
		if err := rows.Scan(&n.ID, &n.CourseID, &n.StudentName, &n.Message, &createdAtStr); err != nil {
			return nil, err
		}
		if createdAtStr.Valid {
			if t, err := database.ParseTime(createdAtStr.String); err == nil {
				n.CreatedAt = &t
			}
		}
		notifications = append(notifications, n)
	}
	if notifications == nil {
		notifications = []models.Notification{}
	}
	return notifications, nil
}

func GetNotificationsByCourse(courseID int64) ([]models.Notification, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, student_name, message, created_at 
		 FROM notifications WHERE course_id = ? ORDER BY created_at DESC`,
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var createdAtStr sql.NullString
		if err := rows.Scan(&n.ID, &n.CourseID, &n.StudentName, &n.Message, &createdAtStr); err != nil {
			return nil, err
		}
		if createdAtStr.Valid {
			if t, err := database.ParseTime(createdAtStr.String); err == nil {
				n.CreatedAt = &t
			}
		}
		notifications = append(notifications, n)
	}
	if notifications == nil {
		notifications = []models.Notification{}
	}
	return notifications, nil
}

func GetAllNotifications() ([]models.Notification, error) {
	rows, err := database.DB.Query(
		`SELECT id, course_id, student_name, message, created_at 
		 FROM notifications ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var createdAtStr sql.NullString
		if err := rows.Scan(&n.ID, &n.CourseID, &n.StudentName, &n.Message, &createdAtStr); err != nil {
			return nil, err
		}
		if createdAtStr.Valid {
			if t, err := database.ParseTime(createdAtStr.String); err == nil {
				n.CreatedAt = &t
			}
		}
		notifications = append(notifications, n)
	}
	if notifications == nil {
		notifications = []models.Notification{}
	}
	return notifications, nil
}
