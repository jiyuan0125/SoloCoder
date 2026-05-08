package protocol

type CompileRequest struct {
	Pattern       string `json:"pattern"`
	CompileTimeMs int    `json:"compile_time_ms,omitempty"`
}

type CompileResponse struct {
	RegexID string `json:"regex_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type MatchRequest struct {
	RegexID    string `json:"regex_id"`
	Text       string `json:"text"`
	MatchTimeMs int   `json:"match_time_ms,omitempty"`
}

type MatchResponse struct {
	Matches []MatchResult `json:"matches"`
	Success bool          `json:"success"`
	Timeout bool          `json:"timeout,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type MatchResult struct {
	Match   string   `json:"match"`
	Groups  []string `json:"groups"`
	Start   int      `json:"start"`
	End     int      `json:"end"`
}

type PurgeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Purged  int    `json:"purged,omitempty"`
}
