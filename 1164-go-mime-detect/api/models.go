package api

type DetectRequest struct {
	FilePath string `json:"file_path"`
}

type DetectResponse struct {
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
	FilePath       string `json:"file_path,omitempty"`
	ExtensionMIME  string `json:"extension_mime,omitempty"`
	MagicMIME      string `json:"magic_mime,omitempty"`
	FinalMIME      string `json:"final_mime,omitempty"`
	Warning        string `json:"warning,omitempty"`
	Confidence     string `json:"confidence,omitempty"`
	IsConsistent   bool   `json:"is_consistent,omitempty"`
	HasWarning     bool   `json:"has_warning,omitempty"`
}
