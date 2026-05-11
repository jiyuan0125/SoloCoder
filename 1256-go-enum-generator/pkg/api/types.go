package api

type GenerateRequest struct {
	SourceCode string `json:"source_code"`
	Filename   string `json:"filename"`
}

type GenerateResponse struct {
	GeneratedCode string `json:"generated_code"`
	Message       string `json:"message,omitempty"`
}
