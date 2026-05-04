package api

type FormatRequest struct {
	Phone      string `json:"phone"`
	FormatType string `json:"format_type"`
}

type FormatResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ValidateRequest struct {
	Phone string `json:"phone"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
}

type ExtractRequest struct {
	Text string `json:"text"`
}

type ExtractResponse struct {
	Success bool     `json:"success"`
	Results []string `json:"results,omitempty"`
	Error   string   `json:"error,omitempty"`
}
