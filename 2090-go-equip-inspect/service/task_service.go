package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"equip-inspect/database"
	"equip-inspect/models"
)

func GetTaskByID(taskID int64) (*models.InspectionTask, error) {
	var task models.InspectionTask
	var inspectorID sql.NullInt64
	var assignedAt sql.NullTime
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var reviewerID sql.NullInt64
	var reviewedAt sql.NullTime
	var totalScore sql.NullInt64

	err := database.DB.QueryRow(`
		SELECT id, plan_id, code, status, inspector_id, assigned_at, started_at, 
		       completed_at, reviewer_id, reviewed_at, COALESCE(review_comment, ''), due_date, 
		       total_score, created_at, updated_at
		FROM inspection_tasks WHERE id = ?
	`, taskID).Scan(
		&task.ID, &task.PlanID, &task.Code, &task.Status,
		&inspectorID, &assignedAt, &startedAt, &completedAt,
		&reviewerID, &reviewedAt, &task.ReviewComment, &task.DueDate,
		&totalScore, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	if inspectorID.Valid {
		val := inspectorID.Int64
		task.InspectorID = &val
	}
	if assignedAt.Valid {
		val := assignedAt.Time
		task.AssignedAt = &val
	}
	if startedAt.Valid {
		val := startedAt.Time
		task.StartedAt = &val
	}
	if completedAt.Valid {
		val := completedAt.Time
		task.CompletedAt = &val
	}
	if reviewerID.Valid {
		val := reviewerID.Int64
		task.ReviewerID = &val
	}
	if reviewedAt.Valid {
		val := reviewedAt.Time
		task.ReviewedAt = &val
	}
	if totalScore.Valid {
		val := int(totalScore.Int64)
		task.TotalScore = &val
	}

	return &task, nil
}

func AssignTask(taskID, inspectorID int64) (*models.InspectionTask, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentInspector sql.NullInt64
	var status string
	err = tx.QueryRow(`
		SELECT inspector_id, status FROM inspection_tasks WHERE id = ? LIMIT 1
	`, taskID).Scan(&currentInspector, &status)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	if currentInspector.Valid {
		return nil, fmt.Errorf("task already assigned to inspector %d", currentInspector.Int64)
	}

	if models.TaskStatus(status) != models.TaskStatusPending {
		return nil, fmt.Errorf("task status is %s, cannot assign", status)
	}

	now := time.Now()
	result, err := tx.Exec(`
		UPDATE inspection_tasks 
		SET inspector_id = ?, assigned_at = ?, status = ?, updated_at = ?
		WHERE id = ? AND inspector_id IS NULL AND status = ?
	`, inspectorID, now, models.TaskStatusInspecting, now, taskID, models.TaskStatusPending)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, errors.New("concurrent update: task already taken")
	}

	if err = UpdateStatistics(tx); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetTaskByID(taskID)
}

func ListTasks(status string, inspectorID int64, limit, offset int) ([]*models.InspectionTask, error) {
	query := `
		SELECT id, plan_id, code, status, inspector_id, assigned_at, started_at, 
		       completed_at, reviewer_id, reviewed_at, COALESCE(review_comment, ''), due_date, 
		       total_score, created_at, updated_at
		FROM inspection_tasks WHERE 1=1
	`
	var args []interface{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if inspectorID > 0 {
		query += " AND inspector_id = ?"
		args = append(args, inspectorID)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.InspectionTask
	for rows.Next() {
		var task models.InspectionTask
		var inspID sql.NullInt64
		var assignedAt sql.NullTime
		var startedAt sql.NullTime
		var completedAt sql.NullTime
		var revID sql.NullInt64
		var reviewedAt sql.NullTime
		var totalScore sql.NullInt64

		err = rows.Scan(
			&task.ID, &task.PlanID, &task.Code, &task.Status,
			&inspID, &assignedAt, &startedAt, &completedAt,
			&revID, &reviewedAt, &task.ReviewComment, &task.DueDate,
			&totalScore, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if inspID.Valid {
			val := inspID.Int64
			task.InspectorID = &val
		}
		if assignedAt.Valid {
			val := assignedAt.Time
			task.AssignedAt = &val
		}
		if startedAt.Valid {
			val := startedAt.Time
			task.StartedAt = &val
		}
		if completedAt.Valid {
			val := completedAt.Time
			task.CompletedAt = &val
		}
		if revID.Valid {
			val := revID.Int64
			task.ReviewerID = &val
		}
		if reviewedAt.Valid {
			val := reviewedAt.Time
			task.ReviewedAt = &val
		}
		if totalScore.Valid {
			val := int(totalScore.Int64)
			task.TotalScore = &val
		}

		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func GetTaskPoints(taskID int64) ([]*models.TaskPoint, error) {
	rows, err := database.DB.Query(`
		SELECT id, task_id, point_id, order_index, checked, checked_at, checked_by
		FROM task_points WHERE task_id = ? ORDER BY order_index
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*models.TaskPoint
	for rows.Next() {
		var tp models.TaskPoint
		var checkedAt sql.NullTime
		var checkedBy sql.NullInt64

		err = rows.Scan(&tp.ID, &tp.TaskID, &tp.PointID, &tp.OrderIndex, &tp.Checked, &checkedAt, &checkedBy)
		if err != nil {
			return nil, err
		}

		if checkedAt.Valid {
			val := checkedAt.Time
			tp.CheckedAt = &val
		}
		if checkedBy.Valid {
			val := checkedBy.Int64
			tp.CheckedBy = &val
		}

		points = append(points, &tp)
	}
	return points, nil
}

func GetNextExpectedPoint(taskID int64) (*models.TaskPoint, error) {
	points, err := GetTaskPoints(taskID)
	if err != nil {
		return nil, err
	}

	for _, p := range points {
		if !p.Checked {
			return p, nil
		}
	}
	return nil, nil
}

func CheckInPoint(taskID, pointID, inspectorID int64) (*models.TaskPoint, error) {
	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	if task.Status != models.TaskStatusInspecting {
		return nil, fmt.Errorf("task status is %s, cannot check in", task.Status)
	}

	if task.InspectorID == nil || *task.InspectorID != inspectorID {
		return nil, errors.New("task not assigned to this inspector")
	}

	nextPoint, err := GetNextExpectedPoint(taskID)
	if err != nil {
		return nil, err
	}

	if nextPoint == nil {
		return nil, errors.New("all points already checked")
	}

	if nextPoint.PointID != pointID {
		return nil, fmt.Errorf("must check point %d first (order %d), cannot skip to point %d",
			nextPoint.PointID, nextPoint.OrderIndex, pointID)
	}

	now := time.Now()
	result, err := database.DB.Exec(`
		UPDATE task_points SET checked = 1, checked_at = ?, checked_by = ?
		WHERE id = ? AND checked = 0
	`, now, inspectorID, nextPoint.ID)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, errors.New("point already checked")
	}

	return GetTaskPointByID(nextPoint.ID)
}

func GetTaskPointByID(tpID int64) (*models.TaskPoint, error) {
	var tp models.TaskPoint
	var checkedAt sql.NullTime
	var checkedBy sql.NullInt64

	err := database.DB.QueryRow(`
		SELECT id, task_id, point_id, order_index, checked, checked_at, checked_by
		FROM task_points WHERE id = ?
	`, tpID).Scan(&tp.ID, &tp.TaskID, &tp.PointID, &tp.OrderIndex, &tp.Checked, &checkedAt, &checkedBy)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	if checkedAt.Valid {
		val := checkedAt.Time
		tp.CheckedAt = &val
	}
	if checkedBy.Valid {
		val := checkedBy.Int64
		tp.CheckedBy = &val
	}

	return &tp, nil
}

func SubmitInspectionRecords(taskID, inspectorID int64, records []models.InspectionRecord) (int, error) {
	task, err := GetTaskByID(taskID)
	if err != nil {
		return 0, err
	}
	if task == nil {
		return 0, errors.New("task not found")
	}

	if task.Status != models.TaskStatusInspecting {
		return 0, fmt.Errorf("task status is %s, cannot submit records", task.Status)
	}

	if task.InspectorID == nil || *task.InspectorID != inspectorID {
		return 0, errors.New("task not assigned to this inspector")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	totalScore := 0
	count := 0

	for _, rec := range records {
		_, err = tx.Exec(`
			INSERT INTO inspection_records (task_point_id, check_item_id, score, description)
			VALUES (?, ?, ?, ?)
		`, rec.TaskPointID, rec.CheckItemID, rec.Score, rec.Description)
		if err != nil {
			return 0, err
		}
		totalScore += rec.Score
		count++
	}

	var finalScore int
	if count > 0 {
		finalScore = totalScore / count
	} else {
		finalScore = 100
	}

	_, err = tx.Exec(`UPDATE inspection_tasks SET total_score = ?, updated_at = ? WHERE id = ?`,
		finalScore, time.Now(), taskID)
	if err != nil {
		return 0, err
	}

	var newStatus models.TaskStatus
	isAbnormal := false
	if finalScore < 80 {
		newStatus = models.TaskStatusAbnormal
		isAbnormal = true
	} else {
		newStatus = models.TaskStatusPendingReview
	}

	now := time.Now()
	_, err = tx.Exec(`
		UPDATE inspection_tasks 
		SET status = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`, newStatus, now, now, taskID)
	if err != nil {
		return 0, err
	}

	if isAbnormal {
		if err = createRepairOrderForTask(tx, task); err != nil {
			return 0, err
		}
	}

	if err = UpdateStatistics(tx); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return finalScore, nil
}
