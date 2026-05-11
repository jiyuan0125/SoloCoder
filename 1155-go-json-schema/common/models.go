package common

import "encoding/json"

type RegisterRequest struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ValidateRequest struct {
	SchemaName string          `json:"schemaName"`
	Data       json.RawMessage `json:"data"`
}

type ValidationError struct {
	Path       string      `json:"path"`
	Constraint string      `json:"constraint"`
	Actual     interface{} `json:"actual,omitempty"`
	Expected   interface{} `json:"expected,omitempty"`
}

type ValidateResponse struct {
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Message string            `json:"message,omitempty"`
}
