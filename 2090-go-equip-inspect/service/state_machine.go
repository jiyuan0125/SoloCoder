package service

import (
	"errors"
	"fmt"
	"time"

	"equip-inspect/database"
	"equip-inspect/models"
)

var stateTransitions = map[models.TaskStatus][]models.TaskStatus{
	models.TaskStatusPending:       {models.TaskStatusInspecting},
	models.TaskStatusInspecting:    {models.TaskStatusNormal, models.TaskStatusAbnormal},
	models.TaskStatusNormal:        {models.TaskStatusPendingReview},
	models.TaskStatusAbnormal:      {models.TaskStatusPendingReview},
	models.TaskStatusPendingReview: {models.TaskStatusReviewPass, models.TaskStatusReviewFail},
	models.TaskStatusReviewFail:    {models.TaskStatusInspecting},
	models.TaskStatusReviewPass:    {models.TaskStatusClosed},
}

func CanTransition(from, to models.TaskStatus) bool {
	allowed, ok := stateTransitions[from]
	if !ok {
		return false
	}
	for _, allowedTo := range allowed {
		if allowedTo == to {
			return true
		}
	}
	return false
}

func GetNextValidStates(current models.TaskStatus) []models.TaskStatus {
	return stateTransitions[current]
}

func TransitionTask(taskID int64, newStatus models.TaskStatus, userID int64, comment string) (*models.InspectionTask, error) {
	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	if task.Status == newStatus {
		return task, nil
	}

	if !CanTransition(task.Status, newStatus) {
		return nil, fmt.Errorf("invalid state transition: %s -> %s. Current status: %s",
			task.Status, newStatus, task.Status)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	if newStatus == models.TaskStatusNormal || newStatus == models.TaskStatusAbnormal {
		_, err = tx.Exec(`
			UPDATE inspection_tasks 
			SET status = ?, completed_at = ?, updated_at = ?
			WHERE id = ?
		`, newStatus, now, now, taskID)
	} else if newStatus == models.TaskStatusReviewPass || newStatus == models.TaskStatusReviewFail {
		_, err = tx.Exec(`
			UPDATE inspection_tasks 
			SET status = ?, reviewer_id = ?, reviewed_at = ?, review_comment = ?, updated_at = ?
			WHERE id = ?
		`, newStatus, userID, now, comment, now, taskID)
	} else if newStatus == models.TaskStatusInspecting {
		_, err = tx.Exec(`
			UPDATE inspection_tasks 
			SET status = ?, updated_at = ?, review_comment = ?, reviewer_id = NULL, reviewed_at = NULL, completed_at = NULL
			WHERE id = ?
		`, newStatus, now, comment, taskID)
	} else {
		_, err = tx.Exec(`
			UPDATE inspection_tasks 
			SET status = ?, updated_at = ?
			WHERE id = ?
		`, newStatus, now, taskID)
	}

	if err != nil {
		return nil, err
	}

	if newStatus == models.TaskStatusAbnormal {
		if err = createRepairOrderForTask(tx, task); err != nil {
			return nil, err
		}
	}

	if err = UpdateStatistics(tx); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetTaskByID(taskID)
}
