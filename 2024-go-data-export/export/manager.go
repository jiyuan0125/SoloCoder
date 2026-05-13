package export

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"data-export/config"
	"data-export/database"
	"data-export/models"
)

var (
	taskMutex   sync.Mutex
	userTasks   = make(map[string]struct{})
	workerQueue = make(chan *models.ExportTask, 10)
)

func InitManager() {
	go worker()
	go cleanupOldFiles()
}

func worker() {
	for task := range workerQueue {
		ExecuteExport(task)

		taskMutex.Lock()
		delete(userTasks, task.UserID)
		taskMutex.Unlock()
	}
}

func SubmitTask(req models.ExportRequest) (*models.ExportTask, error) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	if _, exists := userTasks[req.UserID]; exists {
		return nil, &DuplicateTaskError{}
	}

	var runningTasks int64
	database.DB.Model(&models.ExportTask{}).
		Where("user_id = ? AND status IN ?", req.UserID, []string{config.TaskStatusPending, config.TaskStatusProcessing}).
		Count(&runningTasks)

	if runningTasks > 0 {
		return nil, &DuplicateTaskError{}
	}

	task := &models.ExportTask{
		UserID:    req.UserID,
		TableName: req.TableName,
		Fields:    strings.Join(req.Fields, ","),
		Filter:    req.Filter.Conditions,
		Format:    req.Format,
		Status:    config.TaskStatusPending,
		Progress:  0,
	}

	if err := database.DB.Create(task).Error; err != nil {
		return nil, err
	}

	userTasks[req.UserID] = struct{}{}
	workerQueue <- task

	return task, nil
}

func GetTaskProgress(taskID uint) (*models.ProgressResponse, error) {
	var task models.ExportTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		return nil, err
	}

	resp := &models.ProgressResponse{
		ID:       task.ID,
		Status:   task.Status,
		Progress: task.Progress,
	}

	if task.Status == config.TaskStatusCompleted {
		resp.FileName = task.FileName
	}

	return resp, nil
}

func GetTaskFile(taskID uint) (*models.ExportTask, error) {
	var task models.ExportTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func cleanupOldFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().AddDate(0, 0, -config.FileRetentionDays)

		var oldTasks []models.ExportTask
		database.DB.Where("created_at < ? AND file_path IS NOT NULL", cutoff).Find(&oldTasks)

		for _, task := range oldTasks {
			if task.FilePath != "" {
				if err := os.Remove(task.FilePath); err != nil {
					log.Printf("Failed to remove old file: %v", err)
				}
			}
			database.DB.Model(&task).Updates(map[string]interface{}{
				"file_path": "",
				"file_name": "",
			})
		}

		log.Printf("Cleanup check completed, processed %d old tasks", len(oldTasks))
	}
}

type DuplicateTaskError struct{}

func (e *DuplicateTaskError) Error() string {
	return "已有导出任务进行中"
}
