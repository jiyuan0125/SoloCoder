package queue

import (
	"encoding/json"
	"errors"
	"task-queue/internal/database"
	"task-queue/internal/models"
	"time"

	"gorm.io/gorm"
)

type SubmitTaskRequest struct {
	Type           string          `json:"type"`
	Priority       int             `json:"priority"`
	Payload        json.RawMessage `json:"payload"`
	TimeoutSeconds int             `json:"timeout_seconds"`
	MaxRetries     int             `json:"max_retries"`
}

func ValidateTask(req *SubmitTaskRequest) error {
	if req.Type == "" {
		return errors.New("task type is required")
	}
	if req.Priority < 1 || req.Priority > 10 {
		return errors.New("priority must be between 1 and 10")
	}
	if !json.Valid(req.Payload) {
		return errors.New("payload is not valid JSON")
	}
	return nil
}

func SubmitTask(req *SubmitTaskRequest) (*models.Task, error) {
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 1800
	}
	if req.MaxRetries <= 0 {
		req.MaxRetries = 3
	}

	task := &models.Task{
		Type:           req.Type,
		Priority:       req.Priority,
		Payload:        string(req.Payload),
		Status:         models.StatusPending,
		RetryCount:     0,
		TimeoutCount:   0,
		MaxRetries:     req.MaxRetries,
		TimeoutSeconds: req.TimeoutSeconds,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result := database.DB.Create(task)
	if result.Error != nil {
		return nil, result.Error
	}

	return task, nil
}

func FetchTasks(limit int) ([]*models.Task, error) {
	tasks := make([]*models.Task, 0)

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	result := tx.Raw(`
		SELECT * FROM tasks 
		WHERE status = ? 
		ORDER BY priority DESC, id ASC 
		LIMIT ? 
	`, models.StatusPending, limit).Scan(&tasks)

	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	for _, task := range tasks {
		task.Status = models.StatusProcessing
		task.ProcessingAt = &now
		task.UpdatedAt = now

		updateResult := tx.Model(&models.Task{}).
			Where("id = ? AND status = ?", task.ID, models.StatusPending).
			Updates(map[string]interface{}{
				"status":         models.StatusProcessing,
				"processing_at":  now,
				"updated_at":     now,
			})

		if updateResult.RowsAffected == 0 {
			task.Status = models.StatusPending
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	actualTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.Status == models.StatusProcessing {
			actualTasks = append(actualTasks, task)
		}
	}

	return actualTasks, nil
}

type CompleteTaskRequest struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error"`
}

func CompleteTask(taskID int64, req *CompleteTaskRequest) (*models.Task, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var task models.Task
	result := tx.First(&task, taskID)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}

	now := time.Now()

	if req.Success {
		task.Status = models.StatusSuccess
		task.Result = req.Result
		task.CompletedAt = &now
		task.UpdatedAt = now
	} else {
		task.RetryCount++
		task.Error = req.Error
		task.UpdatedAt = now

		if task.RetryCount >= task.MaxRetries {
			task.Status = models.StatusFailed
			task.CompletedAt = &now
		} else {
			task.Status = models.StatusPending
			task.ProcessingAt = nil
		}
	}

	updateResult := tx.Save(&task)
	if updateResult.Error != nil {
		tx.Rollback()
		return nil, updateResult.Error
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &task, nil
}

func GetTask(taskID int64) (*models.Task, error) {
	var task models.Task
	result := database.DB.First(&task, taskID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

func ProcessTimeouts() error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	result := tx.Exec(`
		UPDATE tasks 
		SET status = ?, 
		    timeout_count = timeout_count + 1,
		    processing_at = NULL,
		    updated_at = ?
		WHERE status = ? 
		AND processing_at IS NOT NULL 
		AND strftime('%s', ?) - strftime('%s', processing_at) > timeout_seconds
		AND retry_count < max_retries
	`, models.StatusPending, now, models.StatusProcessing, now)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	result = tx.Exec(`
		UPDATE tasks 
		SET status = ?, 
		    timeout_count = timeout_count + 1,
		    completed_at = ?,
		    updated_at = ?
		WHERE status = ? 
		AND processing_at IS NOT NULL 
		AND strftime('%s', ?) - strftime('%s', processing_at) > timeout_seconds
		AND retry_count >= max_retries
	`, models.StatusFailed, now, now, models.StatusProcessing, now)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func CleanupOldTasks(days int) error {
	cutoffTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	result := database.DB.Where(
		"status IN ? AND completed_at < ?",
		[]models.TaskStatus{models.StatusSuccess, models.StatusFailed},
		cutoffTime,
	).Delete(&models.Task{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
