package storage

import (
	"datacleanser/internal/common"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Storage struct {
	templates    map[string]*common.Template
	tasks        map[string]*common.AsyncTask
	errorSamples []common.ErrorSample
	mu           sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		templates: make(map[string]*common.Template),
		tasks:     make(map[string]*common.AsyncTask),
	}
}

func (s *Storage) CreateTemplate(template *common.Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.templates[template.Name]; exists {
		return &TemplateAlreadyExistsError{Name: template.Name}
	}

	now := time.Now()
	template.Version = 1
	template.CreatedAt = now
	template.UpdatedAt = now
	template.Versions = []common.TemplateVersion{
		{
			Version:   1,
			Rule:      template.Rule,
			CreatedAt: now,
		},
	}

	s.templates[template.Name] = template
	return nil
}

func (s *Storage) GetTemplate(name string) (*common.Template, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, exists := s.templates[name]
	return t, exists
}

func (s *Storage) ListTemplates() []common.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]common.Template, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, *t)
	}
	return templates
}

func (s *Storage) UpdateTemplate(name string, newRule common.CleanRule) (*common.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.templates[name]
	if !exists {
		return nil, &TemplateNotFoundError{Name: name}
	}

	newVersion := t.Version + 1
	now := time.Now()

	versionHistory := common.TemplateVersion{
		Version:   newVersion,
		Rule:      newRule,
		CreatedAt: now,
	}

	t.Version = newVersion
	t.Rule = newRule
	t.UpdatedAt = now
	t.Versions = append(t.Versions, versionHistory)

	return t, nil
}

func (s *Storage) RollbackTemplate(name string, targetVersion int) (*common.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.templates[name]
	if !exists {
		return nil, &TemplateNotFoundError{Name: name}
	}

	if targetVersion <= 0 || targetVersion > t.Version {
		return nil, &InvalidVersionError{Version: targetVersion, MaxVersion: t.Version}
	}

	var targetRule common.CleanRule
	for _, v := range t.Versions {
		if v.Version == targetVersion {
			targetRule = v.Rule
			break
		}
	}

	return s.updateTemplateInternal(t, targetRule)
}

func (s *Storage) updateTemplateInternal(t *common.Template, newRule common.CleanRule) (*common.Template, error) {
	newVersion := t.Version + 1
	now := time.Now()

	versionHistory := common.TemplateVersion{
		Version:   newVersion,
		Rule:      newRule,
		CreatedAt: now,
	}

	t.Version = newVersion
	t.Rule = newRule
	t.UpdatedAt = now
	t.Versions = append(t.Versions, versionHistory)

	return t, nil
}

func (s *Storage) CreateTask(request common.CleanRequest) *common.AsyncTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &common.AsyncTask{
		ID:        uuid.New().String(),
		Status:    common.TaskPending,
		Request:   request,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.tasks[task.ID] = task
	return task
}

func (s *Storage) GetTask(taskID string) (*common.AsyncTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, exists := s.tasks[taskID]
	return t, exists
}

func (s *Storage) UpdateTaskStatus(taskID string, status common.TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tasks[taskID]
	if !exists {
		return &TaskNotFoundError{TaskID: taskID}
	}

	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

func (s *Storage) UpdateTaskResult(taskID string, result *common.CleanResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tasks[taskID]
	if !exists {
		return &TaskNotFoundError{TaskID: taskID}
	}

	t.Status = common.TaskCompleted
	t.Result = result
	t.UpdatedAt = time.Now()
	return nil
}

func (s *Storage) UpdateTaskError(taskID string, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tasks[taskID]
	if !exists {
		return &TaskNotFoundError{TaskID: taskID}
	}

	t.Status = common.TaskFailed
	t.Error = errMsg
	t.UpdatedAt = time.Now()
	return nil
}

func (s *Storage) AddErrorSample(sample common.ErrorSample) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.errorSamples = append(s.errorSamples, sample)

	if len(s.errorSamples) > 1000 {
		s.errorSamples = s.errorSamples[len(s.errorSamples)-1000:]
	}
}

func (s *Storage) GetErrorSamples() []common.ErrorSample {
	s.mu.RLock()
	defer s.mu.RUnlock()

	samples := make([]common.ErrorSample, len(s.errorSamples))
	copy(samples, s.errorSamples)
	return samples
}

type TemplateAlreadyExistsError struct {
	Name string
}

func (e *TemplateAlreadyExistsError) Error() string {
	return "template already exists: " + e.Name
}

type TemplateNotFoundError struct {
	Name string
}

func (e *TemplateNotFoundError) Error() string {
	return "template not found: " + e.Name
}

type InvalidVersionError struct {
	Version    int
	MaxVersion int
}

func (e *InvalidVersionError) Error() string {
	return "invalid version: " + string(rune(e.Version)) + ", max version: " + string(rune(e.MaxVersion))
}

type TaskNotFoundError struct {
	TaskID string
}

func (e *TaskNotFoundError) Error() string {
	return "task not found: " + e.TaskID
}
