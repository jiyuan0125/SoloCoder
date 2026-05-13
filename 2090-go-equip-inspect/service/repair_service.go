package service

import (
	"database/sql"
	"time"

	"equip-inspect/database"
	"equip-inspect/models"
)

func createRepairOrderForTask(tx *sql.Tx, task *models.InspectionTask) error {
	deviceIDs, err := getTaskDeviceIDs(tx, task.ID)
	if err != nil {
		return err
	}

	for _, deviceID := range deviceIDs {
		_, err = tx.Exec(`
			INSERT INTO repair_orders (task_id, device_id, status, description, created_at, updated_at)
			VALUES (?, ?, 'pending', '设备巡检异常，需要维修', ?, ?)
		`, task.ID, deviceID, time.Now(), time.Now())
		if err != nil {
			return err
		}
	}
	return nil
}

func getTaskDeviceIDs(tx *sql.Tx, taskID int64) ([]int64, error) {
	rows, err := tx.Query(`
		SELECT DISTINCT ip.device_id
		FROM task_points tp
		JOIN inspection_points ip ON tp.point_id = ip.id
		WHERE tp.task_id = ?
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func AssignRepairOrder(orderID, repairerID int64) (*models.RepairOrder, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.Exec(`
		UPDATE repair_orders SET repairer_id = ?, status = 'assigned', assigned_at = ?, updated_at = ?
		WHERE id = ? AND status = 'pending'
	`, repairerID, now, now, orderID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetRepairOrderByID(orderID)
}

func CompleteRepairOrder(orderID int64) (*models.RepairOrder, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var taskID int64
	err = tx.QueryRow(`SELECT task_id FROM repair_orders WHERE id = ?`, orderID).Scan(&taskID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	_, err = tx.Exec(`
		UPDATE repair_orders SET status = 'completed', completed_at = ?, updated_at = ? WHERE id = ?
	`, now, now, orderID)
	if err != nil {
		return nil, err
	}

	var pendingCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM repair_orders WHERE task_id = ? AND status IN ('pending', 'assigned', 'in_progress')
	`, taskID).Scan(&pendingCount)
	if err != nil {
		return nil, err
	}

	if pendingCount == 0 {
		_, err = tx.Exec(`
			UPDATE inspection_tasks SET status = 'pending_review', updated_at = ? WHERE id = ?
		`, now, taskID)
		if err != nil {
			return nil, err
		}

		if err = UpdateStatistics(tx); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetRepairOrderByID(orderID)
}

func GetRepairOrderByID(orderID int64) (*models.RepairOrder, error) {
	var ro models.RepairOrder
	var repairerID sql.NullInt64
	var assignedAt sql.NullTime
	var completedAt sql.NullTime

	err := database.DB.QueryRow(`
		SELECT id, task_id, device_id, repairer_id, status, description, assigned_at, completed_at, created_at, updated_at
		FROM repair_orders WHERE id = ?
	`, orderID).Scan(&ro.ID, &ro.TaskID, &ro.DeviceID, &repairerID, &ro.Status, &ro.Description,
		&assignedAt, &completedAt, &ro.CreatedAt, &ro.UpdatedAt)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	if repairerID.Valid {
		val := repairerID.Int64
		ro.RepairerID = &val
	}
	if assignedAt.Valid {
		val := assignedAt.Time
		ro.AssignedAt = &val
	}
	if completedAt.Valid {
		val := completedAt.Time
		ro.CompletedAt = &val
	}

	return &ro, nil
}

func ListRepairOrders(status string, deviceID int64, limit, offset int) ([]*models.RepairOrder, error) {
	query := `
		SELECT id, task_id, device_id, repairer_id, status, description, assigned_at, completed_at, created_at, updated_at
		FROM repair_orders WHERE 1=1
	`
	var args []interface{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if deviceID > 0 {
		query += " AND device_id = ?"
		args = append(args, deviceID)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.RepairOrder
	for rows.Next() {
		var ro models.RepairOrder
		var repairerID sql.NullInt64
		var assignedAt sql.NullTime
		var completedAt sql.NullTime

		err = rows.Scan(&ro.ID, &ro.TaskID, &ro.DeviceID, &repairerID, &ro.Status, &ro.Description,
			&assignedAt, &completedAt, &ro.CreatedAt, &ro.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if repairerID.Valid {
			val := repairerID.Int64
			ro.RepairerID = &val
		}
		if assignedAt.Valid {
			val := assignedAt.Time
			ro.AssignedAt = &val
		}
		if completedAt.Valid {
			val := completedAt.Time
			ro.CompletedAt = &val
		}

		orders = append(orders, &ro)
	}

	return orders, nil
}
