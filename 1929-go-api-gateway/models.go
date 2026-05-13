package main

type Backend struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type FallbackPolicy string

const (
	FallbackReturnDefault FallbackPolicy = "return_default"
	FallbackIgnore        FallbackPolicy = "ignore"
	FallbackFailAll       FallbackPolicy = "fail_all"
)

type ExecutionMode string

const (
	ExecutionModeParallel ExecutionMode = "parallel"
	ExecutionModeSerial   ExecutionMode = "serial"
)

type ServiceCall struct {
	Name           string         `json:"name"`
	URL            string         `json:"url"`
	Method         string         `json:"method"`
	Mode           ExecutionMode  `json:"mode"`
	FallbackPolicy FallbackPolicy `json:"fallback_policy"`
	DefaultValue   interface{}    `json:"default_value,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Body           interface{}    `json:"body,omitempty"`
}

type Plan struct {
	ID       string            `json:"id"`
	Services []ServiceCall     `json:"services"`
	Template interface{}       `json:"template"`
}

type ServiceResult struct {
	Name     string      `json:"name"`
	Success  bool        `json:"success"`
	Error    string      `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`
	Status   int         `json:"status,omitempty"`
}

type ExecutionError struct {
	ServiceName string
	Message     string
}

func (e *ExecutionError) Error() string {
	return e.Message
}
