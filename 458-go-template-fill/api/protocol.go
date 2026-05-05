package api

type RenderRequest struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

type RenderResponse struct {
	Success  bool     `json:"success"`
	Content  string   `json:"content,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

type RenderFileRequest struct {
	TemplatePath string                 `json:"template_path"`
	Data         map[string]interface{} `json:"data"`
}

type RenderFileResponse struct {
	Success      bool     `json:"success"`
	Content      string   `json:"content,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	TemplatePath string   `json:"template_path,omitempty"`
	Error        string   `json:"error,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
