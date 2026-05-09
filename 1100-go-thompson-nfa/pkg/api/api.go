package api

type CompileRequest struct {
	Pattern string `json:"pattern"`
}

type CompileResponse struct {
	Success      bool   `json:"success"`
	StateCount   int    `json:"state_count"`
	TransitionCount int `json:"transition_count"`
	Error        string `json:"error,omitempty"`
}

type MatchRequest struct {
	Pattern string `json:"pattern"`
	Input   string `json:"input"`
}

type MatchResponse struct {
	Success bool   `json:"success"`
	Match   bool   `json:"match"`
	Error   string `json:"error,omitempty"`
}

type FindAllRequest struct {
	Pattern string `json:"pattern"`
	Input   string `json:"input"`
}

type MatchInfo struct {
	Match  string `json:"match"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
}

type FindAllResponse struct {
	Success bool        `json:"success"`
	Matches []MatchInfo `json:"matches"`
	Error   string      `json:"error,omitempty"`
}

type VisualizeRequest struct {
	Pattern string `json:"pattern"`
}

type VisualizeResponse struct {
	Success bool   `json:"success"`
	Graph   string `json:"graph"`
	Error   string `json:"error,omitempty"`
}

type TestRequest struct {
	Pattern string   `json:"pattern"`
	Inputs  []string `json:"inputs"`
}

type TestResult struct {
	Input    string `json:"input"`
	Match    bool   `json:"match"`
	Expected bool   `json:"expected,omitempty"`
	Passed   bool   `json:"passed,omitempty"`
}

type TestResponse struct {
	Success bool         `json:"success"`
	Results []TestResult `json:"results"`
	Error   string       `json:"error,omitempty"`
}
