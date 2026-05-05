package api

type EvaluateRequest struct {
	Password string `json:"password"`
}

type EvaluateResponse struct {
	Success     bool     `json:"success"`
	Level       string   `json:"level"`
	Score       int      `json:"score"`
	Suggestions []string `json:"suggestions"`
	Error       string   `json:"error,omitempty"`
}
