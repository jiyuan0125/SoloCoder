package service

import (
	"database/sql"
	"time"

	"equip-inspect/database"
)

func UpdateStatistics(tx *sql.Tx) error {
	now := time.Now()
	today := now.Format("2006-01-02")

	var total, pending, completed, abnormal, normal, reviewPass, reviewFail, closed, overdue int

	err := tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks").Scan(&total)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status = 'pending'").Scan(&pending)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status IN ('normal', 'abnormal')").Scan(&completed)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status = 'abnormal'").Scan(&abnormal)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status = 'normal'").Scan(&normal)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status = 'review_pass'").Scan(&reviewPass)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status = 'review_fail'").Scan(&reviewFail)
	if err != nil {
		return err
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM inspection_tasks WHERE status = 'closed'").Scan(&closed)
	if err != nil {
		return err
	}

	err = tx.QueryRow(`
		SELECT COUNT(*) FROM inspection_tasks 
		WHERE status NOT IN ('closed') AND due_date < ?
	`, now).Scan(&overdue)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO statistics (date, total_tasks, pending_tasks, completed_tasks, abnormal_tasks, 
		                        normal_tasks, review_pass_tasks, review_fail_tasks, closed_tasks, overdue_tasks, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			total_tasks = excluded.total_tasks,
			pending_tasks = excluded.pending_tasks,
			completed_tasks = excluded.completed_tasks,
			abnormal_tasks = excluded.abnormal_tasks,
			normal_tasks = excluded.normal_tasks,
			review_pass_tasks = excluded.review_pass_tasks,
			review_fail_tasks = excluded.review_fail_tasks,
			closed_tasks = excluded.closed_tasks,
			overdue_tasks = excluded.overdue_tasks,
			updated_at = excluded.updated_at
	`, today, total, pending, completed, abnormal, normal, reviewPass, reviewFail, closed, overdue, now)

	return err
}

func GetStatistics(date string) (map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var total, pending, completed, abnormal, normal, reviewPass, reviewFail, closed, overdue int
	var updatedAt time.Time

	err := database.DB.QueryRow(`
		SELECT total_tasks, pending_tasks, completed_tasks, abnormal_tasks, 
		       normal_tasks, review_pass_tasks, review_fail_tasks, closed_tasks, overdue_tasks, updated_at
		FROM statistics WHERE date = ?
	`, date).Scan(&total, &pending, &completed, &abnormal, &normal, &reviewPass, &reviewFail, &closed, &overdue, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]interface{}{
				"date":              date,
				"total_tasks":       0,
				"pending_tasks":     0,
				"completed_tasks":   0,
				"abnormal_tasks":    0,
				"normal_tasks":      0,
				"review_pass_tasks": 0,
				"review_fail_tasks": 0,
				"closed_tasks":      0,
				"overdue_tasks":     0,
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"date":              date,
		"total_tasks":       total,
		"pending_tasks":     pending,
		"completed_tasks":   completed,
		"abnormal_tasks":    abnormal,
		"normal_tasks":      normal,
		"review_pass_tasks": reviewPass,
		"review_fail_tasks": reviewFail,
		"closed_tasks":      closed,
		"overdue_tasks":     overdue,
		"updated_at":        updatedAt,
	}, nil
}
