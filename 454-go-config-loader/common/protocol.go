// Package common provides shared communication protocols and types
// between the config server and client.
package common

import (
	"encoding/json"
	"fmt"
)

// ConfigRequest represents a request to get or set configuration values.
type ConfigRequest struct {
	Operation string                 `json:"operation"`
	Path      string                 `json:"path,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// ConfigResponse represents a response from the config server.
type ConfigResponse struct {
	Success bool                   `json:"success"`
	Value   interface{}            `json:"value,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConfigInfo represents configuration information for listing.
type ConfigInfo struct {
	Path        string      `json:"path"`
	Value       interface{} `json:"value"`
	Source      string      `json:"source"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status      string `json:"status"`
	Uptime      string `json:"uptime"`
	Version     string `json:"version"`
	ConfigCount int    `json:"config_count"`
}

// Error codes
const (
	ErrCodeNotFound    = "NOT_FOUND"
	ErrCodeInvalidPath = "INVALID_PATH"
	ErrCodeTypeMismatch = "TYPE_MISMATCH"
	ErrCodeInternal    = "INTERNAL_ERROR"
	ErrCodePermission  = "PERMISSION_DENIED"
)

// API endpoints
const (
	EndpointHealth   = "/health"
	EndpointConfig   = "/config"
	EndpointConfigByPath = "/config/"
	EndpointReload   = "/reload"
	EndpointWatch    = "/watch"
)

// MarshalConfigRequest marshals a ConfigRequest to JSON.
func MarshalConfigRequest(req *ConfigRequest) ([]byte, error) {
	return json.Marshal(req)
}

// UnmarshalConfigRequest unmarshals JSON to a ConfigRequest.
func UnmarshalConfigRequest(data []byte) (*ConfigRequest, error) {
	var req ConfigRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config request: %w", err)
	}
	return &req, nil
}

// MarshalConfigResponse marshals a ConfigResponse to JSON.
func MarshalConfigResponse(resp *ConfigResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// UnmarshalConfigResponse unmarshals JSON to a ConfigResponse.
func UnmarshalConfigResponse(data []byte) (*ConfigResponse, error) {
	var resp ConfigResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config response: %w", err)
	}
	return &resp, nil
}

// NewSuccessResponse creates a successful response with a value.
func NewSuccessResponse(value interface{}) *ConfigResponse {
	return &ConfigResponse{
		Success: true,
		Value:   value,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(errCode, message string) *ConfigResponse {
	return &ConfigResponse{
		Success: false,
		Error:   fmt.Sprintf("%s: %s", errCode, message),
		Metadata: map[string]interface{}{
			"error_code": errCode,
		},
	}
}
