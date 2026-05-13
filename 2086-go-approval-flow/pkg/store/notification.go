package store

import (
	"approval-flow/pkg/db"
	"approval-flow/pkg/model"
	"approval-flow/pkg/util"
)

func CreateNotification(notif *model.Notification) error {
	if notif.ID == "" {
		notif.ID = util.NewUUID()
	}

	query := `INSERT INTO notifications (id, user_id, application_id, type, message) VALUES (?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, notif.ID, notif.UserID, notif.ApplicationID, notif.Type, notif.Message)
	return err
}

func ListNotificationsByUser(userID string) ([]*model.Notification, error) {
	query := `SELECT id, user_id, application_id, type, message, read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifs := []*model.Notification{}
	for rows.Next() {
		notif := &model.Notification{}
		var read int
		err := rows.Scan(&notif.ID, &notif.UserID, &notif.ApplicationID, &notif.Type, &notif.Message, &read, &notif.CreatedAt)
		if err != nil {
			return nil, err
		}
		notif.Read = util.IntToBool(read)
		notifs = append(notifs, notif)
	}
	return notifs, nil
}

func MarkNotificationAsRead(notifID string) error {
	query := `UPDATE notifications SET read = 1 WHERE id = ?`
	_, err := db.DB.Exec(query, notifID)
	return err
}
