package api

type NormalizeRequest struct {
	Text string `json:"text"`
	Form string `json:"form"`
}

type NormalizeResponse struct {
	Success        bool   `json:"success"`
	NormalizedText string `json:"normalized_text"`
	OriginalText   string `json:"original_text"`
	Form           string `json:"form"`
	UnicodeVersion string `json:"unicode_version"`
	Error          string `json:"error,omitempty"`
}

type AnalyzeRequest struct {
	Text string `json:"text"`
	Form string `json:"form"`
}

type CharChangeInfo struct {
	Original      string `json:"original"`
	OriginalCode  string `json:"original_code"`
	Normalized    string `json:"normalized"`
	NormalizedCodes []string `json:"normalized_codes"`
}

type AnalyzeResponse struct {
	Success        bool             `json:"success"`
	NormalizedText string           `json:"normalized_text"`
	OriginalText   string           `json:"original_text"`
	Form           string           `json:"form"`
	UnicodeVersion string           `json:"unicode_version"`
	Changes        []CharChangeInfo `json:"changes"`
	ChangeCount    int              `json:"change_count"`
	Error          string           `json:"error,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
