package protocol

type ParseRequest struct {
	Format     string   `json:"format"`
	Fields     []string `json:"fields,omitempty"`
	LogContent string   `json:"log_content"`
	FilePath   string   `json:"file_path,omitempty"`
}

type LogEntryResponse struct {
	LineNum   int               `json:"line_num"`
	Fields    map[string]string `json:"fields"`
	Line      string            `json:"line,omitempty"`
	Timestamp string            `json:"timestamp,omitempty"`
}

type ParseResponse struct {
	Success    bool                `json:"success"`
	Entries    []LogEntryResponse  `json:"entries"`
	TotalCount int                 `json:"total_count"`
	Errors     []ParseErrorInfo    `json:"errors,omitempty"`
	Message    string              `json:"message,omitempty"`
}

type ParseErrorInfo struct {
	LineNum int    `json:"line_num"`
	Reason  string `json:"reason"`
}

type ListFormatsResponse struct {
	Formats []string `json:"formats"`
}

type FormatRegisterRequest struct {
	Name         string   `json:"name"`
	FormatString string   `json:"format_string"`
	Fields       []string `json:"fields"`
}

type FormatRegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}
