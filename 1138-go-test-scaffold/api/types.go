package api

type GenerateRequest struct {
	FileName   string `json:"file_name"`
	SourceCode string `json:"source_code"`
}

type GenerateResponse struct {
	Success    bool   `json:"success"`
	FileName   string `json:"file_name,omitempty"`
	Code       string `json:"code,omitempty"`
	Error      string `json:"error,omitempty"`
}

type PreviewRequest struct {
	FileName   string `json:"file_name"`
	SourceCode string `json:"source_code"`
}

type PreviewResponse struct {
	Success   bool         `json:"success"`
	FileName  string       `json:"file_name,omitempty"`
	Structs   []string     `json:"structs,omitempty"`
	TestFuncs []TestFuncInfo `json:"test_functions,omitempty"`
	Error     string       `json:"error,omitempty"`
}

type TestFuncInfo struct {
	Name      string       `json:"name"`
	IsMethod  bool         `json:"is_method"`
	Receiver  string       `json:"receiver"`
	Cases     []TestCaseInfo `json:"cases"`
}

type TestCaseInfo struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type HistoryEntry struct {
	ID            string   `json:"id"`
	InputFileName string   `json:"input_file_name"`
	Structs       []string `json:"structs"`
	TestFunctions []string `json:"test_functions"`
	GeneratedAt   string   `json:"generated_at"`
}

type HistoryResponse struct {
	Success bool           `json:"success"`
	History []HistoryEntry `json:"history,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type BatchGenerateRequest struct {
	Files []GenerateRequest `json:"files"`
}

type BatchGenerateResponse struct {
	Success bool               `json:"success"`
	Results []GenerateResponse `json:"results,omitempty"`
	Error   string             `json:"error,omitempty"`
}
