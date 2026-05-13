package template

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"email-template/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

var ErrTemplateExists = errors.New("template name already exists")
var ErrTemplateNotFound = errors.New("template not found")

type HTMLSyntaxError struct {
	Position string
	Err      error
}

func (e *HTMLSyntaxError) Error() string {
	return fmt.Sprintf("HTML syntax error at %s: %v", e.Position, e.Err)
}

type ValidateResult struct {
	Valid bool
	Error *HTMLSyntaxError
}

type CreateRequest struct {
	Name    string
	Subject string
	HTML    string
}

type UpdateRequest struct {
	ID      int64
	Subject string
	HTML    string
}

func (s *Service) ValidateTemplate(ctx context.Context, html string) *ValidateResult {
	if html == "" {
		return &ValidateResult{Valid: true}
	}

	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}

	tmpl, err := template.New("validate").Funcs(funcMap).Parse(html)
	if err != nil {
		pos := extractErrorPosition(err)
		return &ValidateResult{
			Valid: false,
			Error: &HTMLSyntaxError{Position: pos, Err: err},
		}
	}

	var buf bytes.Buffer
	testData := map[string]interface{}{}
	if err := tmpl.Execute(&buf, testData); err != nil {
		pos := extractErrorPosition(err)
		return &ValidateResult{
			Valid: false,
			Error: &HTMLSyntaxError{Position: pos, Err: err},
		}
	}

	return &ValidateResult{Valid: true}
}

func extractErrorPosition(err error) string {
	errStr := err.Error()
	if idx := strings.Index(errStr, ":"); idx > 0 {
		parts := strings.SplitN(errStr, ":", 2)
		if len(parts) >= 2 {
			return parts[0]
		}
	}
	return "unknown"
}

func (s *Service) RenderTemplate(ctx context.Context, html string, data map[string]interface{}) (string, error) {
	funcMap := template.FuncMap{
		"safeHTML": func(val interface{}) template.HTML {
			if val == nil {
				return ""
			}
			switch v := val.(type) {
			case string:
				return template.HTML(v)
			default:
				return template.HTML(fmt.Sprintf("%v", v))
			}
		},
	}

	tmpl, err := template.New("render").Funcs(funcMap).Parse(html)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	safeData := makeSafeData(data)
	if err := tmpl.Execute(&buf, safeData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func makeSafeData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	return result
}

func (s *Service) Create(ctx context.Context, req *CreateRequest) (*database.Template, error) {
	if req.Name == "" {
		return nil, errors.New("template name is required")
	}

	existing, err := s.db.GetTemplateByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTemplateExists
	}

	if result := s.ValidateTemplate(ctx, req.HTML); !result.Valid {
		return nil, result.Error
	}

	tpl := &database.Template{
		Name:    req.Name,
		Subject: req.Subject,
		HTML:    req.HTML,
	}

	id, err := s.db.CreateTemplate(ctx, tpl)
	if err != nil {
		return nil, err
	}

	tpl.ID = id
	return tpl, nil
}

func (s *Service) Get(ctx context.Context, name string) (*database.Template, error) {
	tpl, err := s.db.GetTemplateByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if tpl == nil {
		return nil, ErrTemplateNotFound
	}
	return tpl, nil
}

func (s *Service) Update(ctx context.Context, req *UpdateRequest) (*database.Template, error) {
	if result := s.ValidateTemplate(ctx, req.HTML); !result.Valid {
		return nil, result.Error
	}

	tpl := &database.Template{
		ID:      req.ID,
		Subject: req.Subject,
		HTML:    req.HTML,
	}

	if err := s.db.UpdateTemplate(ctx, tpl); err != nil {
		return nil, err
	}

	return tpl, nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	tpl, err := s.db.GetTemplateByName(ctx, name)
	if err != nil {
		return err
	}
	if tpl == nil {
		return ErrTemplateNotFound
	}

	return s.db.DeleteTemplate(ctx, tpl.ID)
}

func (s *Service) List(ctx context.Context) ([]*database.Template, error) {
	return s.db.ListTemplates(ctx)
}
