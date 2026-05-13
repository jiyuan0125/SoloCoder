package main

import (
	"database/sql"
	"log"
	"time"
)

type EmailSender struct{}
type SMSSender struct{}
type SiteSender struct{}

func (s *EmailSender) Send(user *User, content string) error {
	log.Printf("[EMAIL] Sending to %s (%s): %s", user.Username, user.Email, content)
	return nil
}

func (s *SMSSender) Send(user *User, content string) error {
	log.Printf("[SMS] Sending to %s (%s): %s", user.Username, user.Phone, content)
	return nil
}

func (s *SiteSender) Send(user *User, content string) error {
	log.Printf("[SITE] Sending to %s: %s", user.Username, content)
	return nil
}

func getUserPreference(userID int64, messageType string) (string, error) {
	var channel string
	err := db.QueryRow(
		`SELECT channel FROM notification_preferences 
		 WHERE user_id = ? AND message_type = ?`,
		userID, messageType,
	).Scan(&channel)

	if err == sql.ErrNoRows {
		return ChannelSite.String(), nil
	}
	if err != nil {
		return "", err
	}
	return channel, nil
}

func getUserByID(userID int64) (*User, error) {
	var user User
	var deptID sql.NullInt64

	err := db.QueryRow(
		`SELECT id, username, email, phone, department_id, created_at 
		 FROM users WHERE id = ?`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &deptID, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	if deptID.Valid {
		id := deptID.Int64
		user.DepartmentID = &id
	}

	return &user, nil
}

func userExists(userID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func sendViaChannel(user *User, channel string, content string) error {
	switch channel {
	case ChannelEmail.String():
		sender := &EmailSender{}
		return sender.Send(user, content)
	case ChannelSMS.String():
		sender := &SMSSender{}
		return sender.Send(user, content)
	case ChannelSite.String():
		sender := &SiteSender{}
		return sender.Send(user, content)
	default:
		return nil
	}
}

func createNotificationRecord(userID int64, messageType, content, channel string, status NotificationStatus, degradedFrom *string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO notifications (user_id, message_type, content, channel, status, degraded_from)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, messageType, content, channel, string(status), degradedFrom,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func updateNotificationStatus(id int64, status NotificationStatus, channel string, retryCount int, nextRetryAt *time.Time) error {
	var nextRetryAtStr *string
	if nextRetryAt != nil {
		formatted := nextRetryAt.Format(time.RFC3339)
		nextRetryAtStr = &formatted
	}

	_, err := db.Exec(
		`UPDATE notifications 
		 SET status = ?, channel = ?, retry_count = ?, next_retry_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		string(status), channel, retryCount, nextRetryAtStr, id,
	)
	return err
}

func getUsersByDepartment(deptID int64) ([]*User, error) {
	rows, err := db.Query(
		`SELECT id, username, email, phone, department_id, created_at 
		 FROM users WHERE department_id = ?`,
		deptID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var deptID sql.NullInt64
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &deptID, &user.CreatedAt); err != nil {
			return nil, err
		}
		if deptID.Valid {
			id := deptID.Int64
			user.DepartmentID = &id
		}
		users = append(users, &user)
	}

	return users, nil
}

func departmentExists(deptID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM departments WHERE id = ?`, deptID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func SendNotification(userID int64, messageType, content string) (*NotificationRecord, error) {
	user, err := getUserByID(userID)
	if err != nil {
		return nil, err
	}

	preferredChannel, err := getUserPreference(userID, messageType)
	if err != nil {
		return nil, err
	}

	recordID, err := createNotificationRecord(userID, messageType, content, preferredChannel, StatusPending, nil)
	if err != nil {
		return nil, err
	}

	err = sendViaChannel(user, preferredChannel, content)
	if err != nil {
		log.Printf("Failed to send via %s, attempting downgrade to site message: %v", preferredChannel, err)

		if preferredChannel != ChannelSite.String() {
			err = sendViaChannel(user, ChannelSite.String(), content)
			if err != nil {
				log.Printf("Downgrade to site message also failed: %v", err)
				nextRetry := time.Now().Add(GetRetryInterval())
				updateNotificationStatus(recordID, StatusPending, preferredChannel, 1, &nextRetry)
			} else {
				updateNotificationStatus(recordID, StatusDowngraded, ChannelSite.String(), 0, nil)
			}
		} else {
			nextRetry := time.Now().Add(GetRetryInterval())
			updateNotificationStatus(recordID, StatusPending, preferredChannel, 1, &nextRetry)
		}
	} else {
		updateNotificationStatus(recordID, StatusSuccess, preferredChannel, 0, nil)
	}

	var record NotificationRecord
	var nextRetry sql.NullString
	var degradedFrom sql.NullString

	err = db.QueryRow(
		`SELECT id, user_id, message_type, content, channel, status, retry_count, 
		        next_retry_at, created_at, updated_at, degraded_from
		 FROM notifications WHERE id = ?`,
		recordID,
	).Scan(&record.ID, &record.UserID, &record.MessageType, &record.Content, &record.Channel,
		&record.Status, &record.RetryCount, &nextRetry, &record.CreatedAt, &record.UpdatedAt, &degradedFrom)

	if err != nil {
		return nil, err
	}

	if nextRetry.Valid {
		val := nextRetry.String
		record.NextRetryAt = &val
	}
	if degradedFrom.Valid {
		val := degradedFrom.String
		record.DegradedFrom = &val
	}

	return &record, nil
}

func RetryPendingNotifications() error {
	now := time.Now()
	rows, err := db.Query(
		`SELECT id, user_id, message_type, content, channel, retry_count
		 FROM notifications 
		 WHERE status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)`,
		string(StatusPending), now.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID int64
		var messageType, content, channel string
		var retryCount int

		if err := rows.Scan(&id, &userID, &messageType, &content, &channel, &retryCount); err != nil {
			log.Printf("Failed to scan notification: %v", err)
			continue
		}

		user, err := getUserByID(userID)
		if err != nil {
			log.Printf("Failed to get user %d: %v", userID, err)
			continue
		}

		err = sendViaChannel(user, channel, content)
		newRetryCount := retryCount + 1

		if err != nil {
			log.Printf("Retry %d failed for notification %d: %v", newRetryCount, id, err)
			if newRetryCount >= MaxRetryCount() {
				updateNotificationStatus(id, StatusFailed, channel, newRetryCount, nil)
			} else {
				nextRetry := time.Now().Add(GetRetryInterval())
				updateNotificationStatus(id, StatusPending, channel, newRetryCount, &nextRetry)
			}
		} else {
			log.Printf("Retry succeeded for notification %d", id)
			updateNotificationStatus(id, StatusSuccess, channel, newRetryCount, nil)
		}
	}

	return nil
}
