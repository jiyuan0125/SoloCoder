package api

type AddPatternRequest struct {
	Pattern string `json:"pattern"`
}

type AddPatternResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type AddPatternsRequest struct {
	Patterns []string `json:"patterns"`
}

type AddPatternsResponse struct {
	Success      bool   `json:"success"`
	AddedCount   int    `json:"added_count"`
	SkippedCount int    `json:"skipped_count"`
	Message      string `json:"message,omitempty"`
}

type SearchRequest struct {
	Text string `json:"text"`
}

type Match struct {
	Pattern string `json:"pattern"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
}

type SearchResponse struct {
	Success bool    `json:"success"`
	Matches []Match `json:"matches"`
	Message string  `json:"message,omitempty"`
}

type DictStatusResponse struct {
	Success bool   `json:"success"`
	Count   int    `json:"count"`
	Message string `json:"message,omitempty"`
}

type ClearResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
