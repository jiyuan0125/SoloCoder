package common

import (
	"time"
)

type ExportFormat string

const (
	FormatCSV   ExportFormat = "csv"
	FormatJSON  ExportFormat = "json"
	FormatExcel ExportFormat = "excel"
)

type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeInteger FieldType = "integer"
	FieldTypeFloat   FieldType = "float"
	FieldTypeDate    FieldType = "date"
	FieldTypePhone   FieldType = "phone"
	FieldTypeIDCard  FieldType = "idcard"
)

type FieldMapping struct {
	SourceField   string    `json:"source_field"`
	TargetField   string    `json:"target_field"`
	FieldType     FieldType `json:"field_type"`
	FormatPattern string    `json:"format_pattern,omitempty"`
	Sensitive     bool      `json:"sensitive,omitempty"`
}

type QueryCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type Template struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	DataSource     string            `json:"data_source"`
	QueryCondition []QueryCondition  `json:"query_conditions"`
	OutputFormat   ExportFormat      `json:"output_format"`
	FieldMappings  []FieldMapping    `json:"field_mappings"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type ExportTask struct {
	ID           string                 `json:"id"`
	TemplateID   string                 `json:"template_id"`
	TemplateName string                 `json:"template_name"`
	UserID       string                 `json:"user_id"`
	Params       map[string]interface{} `json:"params,omitempty"`
	Status       TaskStatus             `json:"status"`
	Progress     int                    `json:"progress"`
	TotalRecords int                    `json:"total_records,omitempty"`
	Processed    int                    `json:"processed,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	FileName     string                 `json:"file_name,omitempty"`
	FileSize     int64                  `json:"file_size,omitempty"`
	ExpireAt     time.Time              `json:"expire_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	DurationMs   int64                  `json:"duration_ms,omitempty"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateTemplateRequest struct {
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	DataSource     string           `json:"data_source"`
	QueryCondition []QueryCondition `json:"query_conditions"`
	OutputFormat   ExportFormat     `json:"output_format"`
	FieldMappings  []FieldMapping   `json:"field_mappings"`
}

type UpdateTemplateRequest struct {
	Name           string           `json:"name,omitempty"`
	Description    string           `json:"description,omitempty"`
	DataSource     string           `json:"data_source,omitempty"`
	QueryCondition []QueryCondition `json:"query_conditions,omitempty"`
	OutputFormat   ExportFormat     `json:"output_format,omitempty"`
	FieldMappings  []FieldMapping   `json:"field_mappings,omitempty"`
}

type CreateTaskRequest struct {
	TemplateID string                 `json:"template_id"`
	UserID     string                 `json:"user_id"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

type TaskProgressResponse struct {
	TaskID       string     `json:"task_id"`
	Status       TaskStatus `json:"status"`
	Progress     int        `json:"progress"`
	TotalRecords int        `json:"total_records,omitempty"`
	Processed    int        `json:"processed,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	FileSize     int64      `json:"file_size,omitempty"`
	ExpireAt     *time.Time `json:"expire_at,omitempty"`
}

type DailyStats struct {
	Date        string `json:"date"`
	TotalTasks  int    `json:"total_tasks"`
	SuccessTasks int   `json:"success_tasks"`
	FailedTasks  int   `json:"failed_tasks"`
	AvgDurationMs int64 `json:"avg_duration_ms"`
}

type TemplateStats struct {
	TemplateID   string `json:"template_id"`
	TemplateName string `json:"template_name"`
	UsageCount   int    `json:"usage_count"`
	AvgDurationMs int64 `json:"avg_duration_ms"`
}

type ExportStatsResponse struct {
	DailyStats     []DailyStats     `json:"daily_stats"`
	TopTemplates   []TemplateStats  `json:"top_templates"`
	OverallStats   struct {
		TotalTasks     int   `json:"total_tasks"`
		SuccessRate    string `json:"success_rate"`
		AvgDurationMs  int64  `json:"avg_duration_ms"`
	} `json:"overall_stats"`
}
