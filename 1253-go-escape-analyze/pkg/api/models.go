package api

type EscapeReport struct {
	TotalEscapes int            `json:"total_escapes"`
	Files        []FileReport   `json:"files"`
	FilteredBy   *FilterOptions `json:"filtered_by,omitempty"`
}

type FileReport struct {
	FilePath      string          `json:"file_path"`
	EscapeCount   int             `json:"escape_count"`
	EscapeRecords []EscapeRecord  `json:"escape_records"`
}

type EscapeRecord struct {
	FileName        string `json:"file_name"`
	LineNumber      int    `json:"line_number"`
	VariableName    string `json:"variable_name"`
	EscapeReason    string `json:"escape_reason"`
	FunctionSignature string `json:"function_signature,omitempty"`
	IsGenerated     bool   `json:"is_generated"`
	IsReturnValue   bool   `json:"is_return_value"`
	IsClosureVar    bool   `json:"is_closure_var"`
}

type FilterOptions struct {
	FileName      string `json:"file_name,omitempty"`
	FunctionName  string `json:"function_name,omitempty"`
	EscapeReason  string `json:"escape_reason,omitempty"`
}

type AnalyzeRequest struct {
	OutputText string        `json:"output_text"`
	Filter     *FilterOptions `json:"filter,omitempty"`
}

type AnalyzeResponse struct {
	Success bool         `json:"success"`
	Report  *EscapeReport `json:"report,omitempty"`
	Error   string       `json:"error,omitempty"`
}
