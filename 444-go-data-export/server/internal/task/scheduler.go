package task

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"data-export/pkg/common"
	"data-export/server/internal/data"
	"data-export/server/internal/export"
	"data-export/server/internal/template"
)

const (
	MaxConcurrentPerUser = 3
	FileRetentionDays    = 7
)

type TaskScheduler struct {
	templates      *template.TemplateManager
	dataSource     *data.DataSource
	formatter      *export.Formatter
	statsCollector StatsCollector

	tasks        map[string]*common.ExportTask
	userQueues   map[string][]*common.ExportTask
	userRunning  map[string]int
	files        map[string][]byte

	mu           sync.RWMutex
	workerCh     chan *common.ExportTask
	ctx          context.Context
	cancel       context.CancelFunc
}

type StatsCollector interface {
	RecordTask(task *common.ExportTask)
}

func generateTaskID() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return "task_" + hex.EncodeToString(bytes)
}

func NewTaskScheduler(templates *template.TemplateManager, dataSource *data.DataSource, stats StatsCollector) *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &TaskScheduler{
		templates:      templates,
		dataSource:     dataSource,
		formatter:      export.NewFormatter(),
		statsCollector: stats,
		tasks:          make(map[string]*common.ExportTask),
		userQueues:     make(map[string][]*common.ExportTask),
		userRunning:    make(map[string]int),
		files:          make(map[string][]byte),
		workerCh:       make(chan *common.ExportTask, 100),
		ctx:            ctx,
		cancel:         cancel,
	}

	go s.runWorkers()
	go s.cleanupExpiredFiles()

	return s
}

func (s *TaskScheduler) runWorkers() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.workerCh:
			s.executeTask(task)
		}
	}
}

func (s *TaskScheduler) cleanupExpiredFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.doCleanup()
		}
	}
}

func (s *TaskScheduler) doCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, task := range s.tasks {
		if task.Status == common.TaskStatusCompleted && !task.ExpireAt.IsZero() && now.After(task.ExpireAt) {
			delete(s.files, id)
			task.FileName = ""
			task.FileSize = 0
		}
	}
}

func (s *TaskScheduler) SubmitTask(req *common.CreateTaskRequest) (*common.ExportTask, error) {
	tpl, err := s.templates.GetByID(req.TemplateID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	running := s.userRunning[req.UserID]
	queueLen := len(s.userQueues[req.UserID])

	if running+queueLen >= MaxConcurrentPerUser*2 {
		return nil, common.NewAppError(common.ErrCodeUserQueueFull)
	}

	now := time.Now()
	task := &common.ExportTask{
		ID:           generateTaskID(),
		TemplateID:   req.TemplateID,
		TemplateName: tpl.Name,
		UserID:       req.UserID,
		Params:       req.Params,
		Status:       common.TaskStatusPending,
		Progress:     0,
		CreatedAt:    now,
	}

	s.tasks[task.ID] = task

	if running < MaxConcurrentPerUser {
		s.userRunning[req.UserID]++
		task.Status = common.TaskStatusRunning
		startedAt := now
		task.StartedAt = &startedAt
		go func() {
			s.workerCh <- task
		}()
	} else {
		s.userQueues[req.UserID] = append(s.userQueues[req.UserID], task)
	}

	return task, nil
}

func (s *TaskScheduler) executeTask(task *common.ExportTask) {
	defer s.finishTask(task)

	tpl, err := s.templates.GetByID(task.TemplateID)
	if err != nil {
		s.updateTaskStatus(task, common.TaskStatusFailed, err.Error())
		return
	}

	filters := make(map[string]interface{})
	if task.Params != nil {
		for k, v := range task.Params {
			filters[k] = v
		}
	}
	for _, cond := range tpl.QueryCondition {
		filters[cond.Field] = cond.Value
	}

	records, err := s.dataSource.Query(tpl.DataSource, filters)
	if err != nil {
		s.updateTaskStatus(task, common.TaskStatusFailed, err.Error())
		return
	}

	s.mu.Lock()
	task.TotalRecords = len(records)
	s.mu.Unlock()

	var processed int
	var batchSize = 100
	var allRecords []map[string]interface{}

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]
		allRecords = append(allRecords, batch...)

		processed = end
		progress := int(float64(processed) / float64(len(records)) * 100)

		s.mu.Lock()
		task.Processed = processed
		task.Progress = progress
		s.mu.Unlock()

		time.Sleep(10 * time.Millisecond)
	}

	data, err := s.formatter.Format(allRecords, tpl.FieldMappings, tpl.OutputFormat)
	if err != nil {
		s.updateTaskStatus(task, common.TaskStatusFailed, err.Error())
		return
	}

	ext := string(tpl.OutputFormat)
	if tpl.OutputFormat == common.FormatExcel {
		ext = "csv"
	}
	fileName := task.TemplateName + "_" + time.Now().Format("20060102150405") + "." + ext

	s.mu.Lock()
	s.files[task.ID] = bytes.Clone(data)
	task.FileName = fileName
	task.FileSize = int64(len(data))
	task.ExpireAt = time.Now().Add(FileRetentionDays * 24 * time.Hour)
	s.mu.Unlock()

	s.updateTaskStatus(task, common.TaskStatusCompleted, "")
}

func (s *TaskScheduler) updateTaskStatus(task *common.ExportTask, status common.TaskStatus, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.Status = status
	if errMsg != "" {
		task.ErrorMessage = errMsg
	}

	if status == common.TaskStatusCompleted || status == common.TaskStatusFailed {
		now := time.Now()
		task.CompletedAt = &now
		if task.StartedAt != nil {
			task.DurationMs = now.Sub(*task.StartedAt).Milliseconds()
		}
	}
}

func (s *TaskScheduler) finishTask(task *common.ExportTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userRunning[task.UserID]--
	if s.userRunning[task.UserID] <= 0 {
		delete(s.userRunning, task.UserID)
	}

	if s.statsCollector != nil {
		s.statsCollector.RecordTask(task)
	}

	queue := s.userQueues[task.UserID]
	if len(queue) > 0 {
		nextTask := queue[0]
		s.userQueues[task.UserID] = queue[1:]

		if s.userRunning[task.UserID] < MaxConcurrentPerUser {
			s.userRunning[task.UserID]++
			nextTask.Status = common.TaskStatusRunning
			now := time.Now()
			nextTask.StartedAt = &now
			go func() {
				s.workerCh <- nextTask
			}()
		}
	}

	if len(s.userQueues[task.UserID]) == 0 {
		delete(s.userQueues, task.UserID)
	}
}

func (s *TaskScheduler) GetTask(taskID string) (*common.ExportTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, common.NewAppError(common.ErrCodeTaskNotFound)
	}

	return task, nil
}

func (s *TaskScheduler) GetTaskProgress(taskID string) (*common.TaskProgressResponse, error) {
	task, err := s.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	resp := &common.TaskProgressResponse{
		TaskID:       task.ID,
		Status:       task.Status,
		Progress:     task.Progress,
		TotalRecords: task.TotalRecords,
		Processed:    task.Processed,
		ErrorMessage: task.ErrorMessage,
	}

	if task.Status == common.TaskStatusCompleted {
		resp.FileName = task.FileName
		resp.FileSize = task.FileSize
		if !task.ExpireAt.IsZero() {
			expireAt := task.ExpireAt
			resp.ExpireAt = &expireAt
		}
	}

	return resp, nil
}

func (s *TaskScheduler) GetTaskFile(taskID string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, "", common.NewAppError(common.ErrCodeTaskNotFound)
	}

	if task.Status != common.TaskStatusCompleted {
		if task.Status == common.TaskStatusFailed {
			return nil, "", common.NewAppError(common.ErrCodeTaskFailed)
		}
		return nil, "", common.NewAppError(common.ErrCodeTaskNotCompleted)
	}

	if !task.ExpireAt.IsZero() && time.Now().After(task.ExpireAt) {
		return nil, "", common.NewAppError(common.ErrCodeFileExpired)
	}

	fileData, exists := s.files[taskID]
	if !exists || len(fileData) == 0 {
		return nil, "", common.NewAppError(common.ErrCodeFileNotFound)
	}

	return fileData, task.FileName, nil
}

func (s *TaskScheduler) ListTasks(userID string) []*common.ExportTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.ExportTask
	for _, task := range s.tasks {
		if userID == "" || task.UserID == userID {
			result = append(result, task)
		}
	}
	return result
}

func (s *TaskScheduler) Stop() {
	s.cancel()
}
