package common

import "encoding/json"

type ParseRequest struct {
	Where string `json:"where"`
}

type ParseResponse struct {
	Success bool            `json:"success"`
	AST     json.RawMessage `json:"ast,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type MatchRequest struct {
	Where string                 `json:"where"`
	Data  map[string]interface{} `json:"data"`
}

type MatchResponse struct {
	Success bool   `json:"success"`
	Match   bool   `json:"match,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ExplainResponse struct {
	Success bool     `json:"success"`
	Steps   []string `json:"steps,omitempty"`
	Error   string   `json:"error,omitempty"`
}
