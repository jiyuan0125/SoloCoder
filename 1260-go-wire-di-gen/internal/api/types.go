package api

type GenerateRequest struct {
	Files map[string]string `json:"files"`
}

type GenerateResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
	CyclePath []string `json:"cycle_path,omitempty"`
}

type ErrorType string

const (
	ErrorTypeCircularDependency ErrorType = "circular_dependency"
	ErrorTypeMultipleProviders  ErrorType = "multiple_providers"
	ErrorTypeMissingProvider    ErrorType = "missing_provider"
	ErrorTypeUnknown            ErrorType = "unknown"
)

func (e ErrorType) String() string {
	return string(e)
}
