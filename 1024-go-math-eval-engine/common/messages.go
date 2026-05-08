package common

type EvaluateRequest struct {
	Expression string                 `json:"expression"`
	Variables  map[string]interface{} `json:"variables"`
}

type EvaluateResponse struct {
	Success bool        `json:"success"`
	Value   interface{} `json:"value,omitempty"`
	Type    string      `json:"type,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}
