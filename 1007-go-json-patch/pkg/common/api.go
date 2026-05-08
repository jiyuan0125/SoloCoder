package common

import (
	"encoding/json"
)

type ApplyRequest struct {
	Document json.RawMessage `json:"document"`
	Patch    json.RawMessage `json:"patch"`
}

type ApplyResponse struct {
	Success  bool            `json:"success"`
	Document json.RawMessage `json:"document,omitempty"`
	Error    string          `json:"error,omitempty"`
}

type ValidateRequest struct {
	Patch json.RawMessage `json:"patch"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid,omitempty"`
	Error   string `json:"error,omitempty"`
}
