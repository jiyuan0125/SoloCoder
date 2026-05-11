package api

type ConvertRequest struct {
	Value     float64 `json:"value"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	TempDelta bool    `json:"temp_delta,omitempty"`
}

type ConvertResponse struct {
	Value    float64 `json:"value"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Original float64 `json:"original"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
