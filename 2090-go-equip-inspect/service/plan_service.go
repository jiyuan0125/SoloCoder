package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"equip-inspect/database"
	"equip-inspect/models"
)

func CreatePlan(name, frequency string, pointIDs []int64) (*models.InspectionPlan, error) {
	if frequency != "daily" && frequency != "weekly" && frequency != "monthly" {
		return nil, errors.New("invalid frequency: must be daily, weekly, or monthly")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.Exec(`
		INSERT INTO inspection_plans (name, frequency, start_date, active, created_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)
	`, name, frequency, now, now, now)
	if err != nil {
		return nil, err
	}

	planID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for i, pointID := range pointIDs {
		_, err = tx.Exec(`
			INSERT INTO plan_points (plan_id, point_id, order_index)
			VALUES (?, ?, ?)
		`, planID, pointID, i+1)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetPlanByID(planID)
}

func GetPlanByID(planID int64) (*models.InspectionPlan, error) {
	var plan models.InspectionPlan
	var lastGen sql.NullTime

	err := database.DB.QueryRow(`
		SELECT id, name, frequency, start_date, last_generate_at, active, created_at, updated_at
		FROM inspection_plans WHERE id = ?
	`, planID).Scan(&plan.ID, &plan.Name, &plan.Frequency, &plan.StartDate, &lastGen,
		&plan.Active, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	if lastGen.Valid {
		val := lastGen.Time
		plan.LastGenerateAt = &val
	}

	return &plan, nil
}

func GenerateTasksFromPlan(planID int64) ([]*models.InspectionTask, error) {
	plan, err := GetPlanByID(planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	if !plan.Active {
		return nil, errors.New("plan is not active")
	}

	now := time.Now()
	shouldGenerate := false
	var dueDate time.Time

	switch plan.Frequency {
	case "daily":
		if plan.LastGenerateAt == nil || plan.LastGenerateAt.Day() != now.Day() {
			shouldGenerate = true
			dueDate = now.Add(24 * time.Hour)
		}
	case "weekly":
		if plan.LastGenerateAt == nil || now.Sub(*plan.LastGenerateAt) >= 7*24*time.Hour {
			shouldGenerate = true
			dueDate = now.Add(7 * 24 * time.Hour)
		}
	case "monthly":
		if plan.LastGenerateAt == nil || now.Sub(*plan.LastGenerateAt) >= 30*24*time.Hour {
			shouldGenerate = true
			dueDate = now.Add(30 * 24 * time.Hour)
		}
	}

	if !shouldGenerate {
		return nil, nil
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT point_id, order_index FROM plan_points WHERE plan_id = ? ORDER BY order_index
	`, planID)
	if err != nil {
		return nil, err
	}

	var planPoints []*models.PlanPoint
	for rows.Next() {
		var pp models.PlanPoint
		if err = rows.Scan(&pp.PointID, &pp.OrderIndex); err != nil {
			rows.Close()
			return nil, err
		}
		planPoints = append(planPoints, &pp)
	}
	rows.Close()

	if len(planPoints) == 0 {
		return nil, errors.New("plan has no points")
	}

	taskCode := fmt.Sprintf("TASK-%s-%d", plan.Frequency, now.Unix())
	result, err := tx.Exec(`
		INSERT INTO inspection_tasks (plan_id, code, status, due_date, created_at, updated_at)
		VALUES (?, ?, 'pending', ?, ?, ?)
	`, planID, taskCode, dueDate, now, now)
	if err != nil {
		return nil, err
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, pp := range planPoints {
		_, err = tx.Exec(`
			INSERT INTO task_points (task_id, point_id, order_index, checked)
			VALUES (?, ?, ?, 0)
		`, taskID, pp.PointID, pp.OrderIndex)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(`
		UPDATE inspection_plans SET last_generate_at = ?, updated_at = ? WHERE id = ?
	`, now, now, planID)
	if err != nil {
		return nil, err
	}

	if err = UpdateStatistics(tx); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	return []*models.InspectionTask{task}, nil
}

func ListPlans(limit, offset int) ([]*models.InspectionPlan, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, frequency, start_date, last_generate_at, active, created_at, updated_at
		FROM inspection_plans ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.InspectionPlan
	for rows.Next() {
		var plan models.InspectionPlan
		var lastGen sql.NullTime
		err = rows.Scan(&plan.ID, &plan.Name, &plan.Frequency, &plan.StartDate, &lastGen,
			&plan.Active, &plan.CreatedAt, &plan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if lastGen.Valid {
			val := lastGen.Time
			plan.LastGenerateAt = &val
		}
		plans = append(plans, &plan)
	}
	return plans, nil
}
