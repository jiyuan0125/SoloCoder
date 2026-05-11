package api

type ParseRequest struct {
	Content string `json:"content"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	Body    interface{} `json:"body,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ValidateRequest struct {
	Content string      `json:"content"`
	Schema  interface{} `json:"schema"`
}

type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ValidateResponse struct {
	Success bool              `json:"success"`
	Valid   bool              `json:"valid,omitempty"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type FormatRequest struct {
	Content string `json:"content"`
}

type FormatResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type GetRequest struct {
	Content string `json:"content"`
	Path    string `json:"path"`
}

type GetResponse struct {
	Success bool        `json:"success"`
	Value   interface{} `json:"value,omitempty"`
	Error   string      `json:"error,omitempty"`
}
