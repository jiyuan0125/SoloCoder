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

type runningTask struct {
	task     *models.ExportTask
	progress int
	status   string
}

var (
	taskMutex       sync.RWMutex
	userTasks       = make(map[string]struct{})
	runningTasksMap = make(map[uint]*runningTask)
	workerQueue     = make(chan *models.ExportTask, 100)
)

func InitManager() {
	go worker()
	go cleanupOldFiles()
	go periodicDBFlush()
}

func worker() {
	for task := range workerQueue {
		taskMutex.Lock()
		runningTasksMap[task.ID] = &runningTask{
			task:     task,
			progress: 0,
			status:   config.TaskStatusProcessing,
		}
		taskMutex.Unlock()

		database.DB.Model(task).Updates(map[string]interface{}{
			"status":   config.TaskStatusProcessing,
			"progress": 0,
		})

		ExecuteExport(task)

		taskMutex.Lock()
		delete(runningTasksMap, task.ID)
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

func UpdateProgress(taskID uint, progress int) {
	taskMutex.Lock()
	if rt, exists := runningTasksMap[taskID]; exists {
		rt.progress = progress
	}
	taskMutex.Unlock()
}

func UpdateTaskStatusDB(taskID uint, status string, progress int, filePath, fileName, errMsg string) {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}
	if filePath != "" {
		updates["file_path"] = filePath
	}
	if fileName != "" {
		updates["file_name"] = fileName
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}

	database.DB.Model(&models.ExportTask{}).Where("id = ?", taskID).Updates(updates)
}

func GetTaskProgress(taskID uint) (*models.ProgressResponse, error) {
	taskMutex.RLock()
	if rt, exists := runningTasksMap[taskID]; exists {
		resp := &models.ProgressResponse{
			ID:       taskID,
			Status:   rt.status,
			Progress: rt.progress,
		}
		taskMutex.RUnlock()
		return resp, nil
	}
	taskMutex.RUnlock()

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

func periodicDBFlush() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		taskMutex.RLock()
		snapshots := make(map[uint]int)
		for id, rt := range runningTasksMap {
			snapshots[id] = rt.progress
		}
		taskMutex.RUnlock()

		for id, progress := range snapshots {
			database.DB.Model(&models.ExportTask{}).
				Where("id = ? AND status = ?", id, config.TaskStatusProcessing).
				Update("progress", progress)
		}
	}
}

func cleanupOldFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().AddDate(0, 0, -config.FileRetentionDays)

		var oldTasks []models.ExportTask
		database.DB.Where("created_at < ? AND file_path IS NOT NULL AND file_path != ''", cutoff).Find(&oldTasks)

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
