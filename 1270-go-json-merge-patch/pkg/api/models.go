package api

import "encoding/json"

type ApplyRequest struct {
	Original json.RawMessage `json:"original"`
	Patch    json.RawMessage `json:"patch"`
}

type ApplyResponse struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error,omitempty"`
}

type DiffRequest struct {
	Source json.RawMessage `json:"source"`
	Target json.RawMessage `json:"target"`
}

type DiffResponse struct {
	Patch json.RawMessage `json:"patch"`
	Error string          `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
