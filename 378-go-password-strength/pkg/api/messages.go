package api

type EvaluateRequest struct {
	Password string `json:"password"`
}

type EvaluateResponse struct {
	Success     bool     `json:"success"`
	Level       string   `json:"level,omitempty"`
	Score       int      `json:"score,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
	Error       string   `json:"error,omitempty"`
}
