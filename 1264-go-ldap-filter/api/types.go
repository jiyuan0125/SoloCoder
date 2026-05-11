package api

type ParseRequest struct {
	Filter string `json:"filter"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	AST     interface{} `json:"ast,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type MatchRequest struct {
	Filter     string              `json:"filter"`
	Attributes map[string][]string `json:"attributes"`
}

type MatchResponse struct {
	Success bool   `json:"success"`
	Matched bool   `json:"matched"`
	Error   string `json:"error,omitempty"`
}
