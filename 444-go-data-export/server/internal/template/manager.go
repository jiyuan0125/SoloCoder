package template

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"data-export/pkg/common"
)

type TemplateManager struct {
	templates map[string]*common.Template
	nameIndex map[string]string
	mu        sync.RWMutex
}

func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*common.Template),
		nameIndex: make(map[string]string),
	}
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "tpl_" + hex.EncodeToString(bytes)
}

func (m *TemplateManager) Create(req *common.CreateTemplateRequest) (*common.Template, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nameIndex[req.Name]; exists {
		return nil, common.NewAppError(common.ErrCodeTemplateNameExists)
	}

	if err := m.validateCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now()
	tpl := &common.Template{
		ID:             generateID(),
		Name:           req.Name,
		Description:    req.Description,
		DataSource:     req.DataSource,
		QueryCondition: req.QueryCondition,
		OutputFormat:   req.OutputFormat,
		FieldMappings:  req.FieldMappings,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	m.templates[tpl.ID] = tpl
	m.nameIndex[tpl.Name] = tpl.ID

	return tpl, nil
}

func (m *TemplateManager) validateCreateRequest(req *common.CreateTemplateRequest) error {
	if req.Name == "" {
		return common.NewInvalidRequestError("template name is required")
	}
	if req.DataSource == "" {
		return common.NewInvalidRequestError("data source is required")
	}
	if req.OutputFormat != common.FormatCSV &&
		req.OutputFormat != common.FormatJSON &&
		req.OutputFormat != common.FormatExcel {
		return common.NewInvalidRequestError("invalid output format, must be csv, json, or excel")
	}
	if len(req.FieldMappings) == 0 {
		return common.NewInvalidRequestError("at least one field mapping is required")
	}
	for i, fm := range req.FieldMappings {
		if fm.SourceField == "" {
			return common.NewInvalidRequestError("field mapping source_field is required at index " + string(rune(i)))
		}
		if fm.TargetField == "" {
			return common.NewInvalidRequestError("field mapping target_field is required at index " + string(rune(i)))
		}
	}
	return nil
}

func (m *TemplateManager) Update(id string, req *common.UpdateTemplateRequest) (*common.Template, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tpl, exists := m.templates[id]
	if !exists {
		return nil, common.NewAppError(common.ErrCodeTemplateNotFound)
	}

	if req.Name != "" && req.Name != tpl.Name {
		if existingID, exists := m.nameIndex[req.Name]; exists && existingID != id {
			return nil, common.NewAppError(common.ErrCodeTemplateNameExists)
		}
		delete(m.nameIndex, tpl.Name)
		tpl.Name = req.Name
		m.nameIndex[tpl.Name] = tpl.ID
	}

	if req.Description != "" {
		tpl.Description = req.Description
	}
	if req.DataSource != "" {
		tpl.DataSource = req.DataSource
	}
	if req.QueryCondition != nil {
		tpl.QueryCondition = req.QueryCondition
	}
	if req.OutputFormat != "" {
		if req.OutputFormat != common.FormatCSV &&
			req.OutputFormat != common.FormatJSON &&
			req.OutputFormat != common.FormatExcel {
			return nil, common.NewInvalidRequestError("invalid output format")
		}
		tpl.OutputFormat = req.OutputFormat
	}
	if req.FieldMappings != nil {
		if len(req.FieldMappings) == 0 {
			return nil, common.NewInvalidRequestError("field mappings cannot be empty")
		}
		tpl.FieldMappings = req.FieldMappings
	}

	tpl.UpdatedAt = time.Now()
	return tpl, nil
}

func (m *TemplateManager) GetByID(id string) (*common.Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tpl, exists := m.templates[id]
	if !exists {
		return nil, common.NewAppError(common.ErrCodeTemplateNotFound)
	}
	return tpl, nil
}

func (m *TemplateManager) GetByName(name string) (*common.Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, exists := m.nameIndex[name]
	if !exists {
		return nil, common.NewAppError(common.ErrCodeTemplateNotFound)
	}
	return m.templates[id], nil
}

func (m *TemplateManager) List() []*common.Template {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*common.Template, 0, len(m.templates))
	for _, tpl := range m.templates {
		result = append(result, tpl)
	}
	return result
}

func (m *TemplateManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tpl, exists := m.templates[id]
	if !exists {
		return common.NewAppError(common.ErrCodeTemplateNotFound)
	}

	delete(m.nameIndex, tpl.Name)
	delete(m.templates, id)
	return nil
}
