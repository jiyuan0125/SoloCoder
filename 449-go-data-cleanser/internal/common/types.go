package common

import "time"

const MaxBatchSize = 10000

type CleanRule struct {
	DedupFields      []string          `json:"dedup_fields"`
	DefaultValues    map[string]string `json:"default_values"`
	PhoneFields      []string          `json:"phone_fields"`
	DateFields       []string          `json:"date_fields"`
	AmountFields     []string          `json:"amount_fields"`
	KeyFields        []string          `json:"key_fields"`
}

type Template struct {
	Name        string     `json:"name"`
	Version     int        `json:"version"`
	Rule        CleanRule  `json:"rule"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Versions    []TemplateVersion `json:"versions,omitempty"`
}

type TemplateVersion struct {
	Version   int       `json:"version"`
	Rule      CleanRule `json:"rule"`
	CreatedAt time.Time `json:"created_at"`
}

type CleanRequest struct {
	Records    []map[string]interface{} `json:"records"`
	Rule     *CleanRule                 `json:"rule,omitempty"`
	TemplateName string                 `json:"template_name,omitempty"`
	Async      bool                   `json:"async,omitempty"`
}

type CleanResponse struct {
	Success  bool                   `json:"success"`
	Message  string                 `json:"message,omitempty"`
	Data     *CleanResult           `json:"data,omitempty"`
	TaskID   string                 `json:"task_id,omitempty"`
}

type CleanResult struct {
	CleanedRecords  []map[string]interface{} `json:"cleaned_records"`
	DuplicateRecords []map[string]interface{} `json:"duplicate_records"`
	InvalidRecords   []map[string]interface{} `json:"invalid_records"`
	ErrorRecords     []map[string]interface{} `json:"error_records"`
	Report           CleanReport             `json:"report"`
}

type CleanReport struct {
	TotalRecords     int            `json:"total_records"`
	DuplicateCount    int            `json:"duplicate_count"`
	CorrectedCount    int            `json:"corrected_count"`
	InvalidCount      int            `json:"invalid_count"`
	RuleHits          map[string]int `json:"rule_hits"`
	QualityScore       float64        `json:"quality_score"`
	StartTime          time.Time      `json:"start_time"`
	EndTime            time.Time      `json:"end_time"`
}

type CreateTemplateRequest struct {
	Name string    `json:"name"`
	Rule CleanRule `json:"rule"`
}

type UpdateTemplateRequest struct {
	Name string    `json:"name"`
	Rule CleanRule `json:"rule"`
}

type TemplateResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Data    *Template `json:"data,omitempty"`
}

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

type AsyncTask struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	Request   CleanRequest `json:"request"`
	Result    *CleanResult `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type TaskResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    *AsyncTask `json:"data,omitempty"`
}

type ErrorSample struct {
	Record      map[string]interface{} `json:"record"`
	Error       string                 `json:"error"`
	Timestamp   time.Time              `json:"timestamp"`
	BatchID     string                 `json:"batch_id"`
}

type ListTemplatesResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    []Template  `json:"data,omitempty"`
}
