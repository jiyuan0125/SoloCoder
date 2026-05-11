package api

type ParseRequest struct {
	Input string `json:"input"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type EvalRequest struct {
	Input string `json:"input"`
}

type EvalResponse struct {
	Success bool        `json:"success"`
	Result  string      `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type FormatRequest struct {
	Input string `json:"input"`
}

type FormatResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
