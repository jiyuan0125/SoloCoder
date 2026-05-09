package common

type ParseRequest struct {
	CSV        string `json:"csv"`
	WithHeader bool   `json:"with_header"`
}

type ParseResponse struct {
	Success bool              `json:"success"`
	Records [][]string        `json:"records,omitempty"`
	Headers []string          `json:"headers,omitempty"`
	Rows    []map[string]string `json:"rows,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type SerializeRequest struct {
	Data [][]string `json:"data"`
}

type SerializeResponse struct {
	Success bool   `json:"success"`
	CSV     string `json:"csv,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ValidateRequest struct {
	CSV        string `json:"csv"`
	WithHeader bool   `json:"with_header"`
}

type ValidateError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

type ValidateResponse struct {
	Success bool             `json:"success"`
	Valid   bool             `json:"valid"`
	Errors  []ValidateError  `json:"errors,omitempty"`
	Error   string           `json:"error,omitempty"`
}
