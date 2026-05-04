package proto

type ExportFormat string

const (
	FormatCSV       ExportFormat = "csv"
	FormatJSON      ExportFormat = "json"
	FormatJSONLines ExportFormat = "jsonl"
)

type ExportRequest struct {
	Data        []map[string]string `json:"data"`
	Format      ExportFormat        `json:"format"`
	Fields      []string            `json:"fields,omitempty"`
	FieldMapping map[string]string  `json:"field_mapping,omitempty"`
}

type ExportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}
