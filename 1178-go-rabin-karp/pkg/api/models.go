package api

type AddRequest struct {
	Pattern string `json:"pattern"`
}

type AddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ImportRequest struct {
	Patterns []string `json:"patterns"`
}

type ImportResponse struct {
	Added   int    `json:"added"`
	Skipped int    `json:"skipped"`
	Message string `json:"message,omitempty"`
}

type SearchRequest struct {
	Text string `json:"text"`
}

type SearchMatch struct {
	Pattern string `json:"pattern"`
	Index   int    `json:"index"`
}

type SearchResponse struct {
	Matches []SearchMatch `json:"matches"`
	Count   int           `json:"count"`
}

type Pattern struct {
	Pattern string `json:"pattern"`
	Length  int    `json:"length"`
}

type PatternsResponse struct {
	Patterns []Pattern       `json:"patterns"`
	ByLength map[int][]string `json:"by_length"`
	Total    int             `json:"total"`
}

type CollisionStatsResponse struct {
	TotalChecks    int            `json:"total_checks"`
	FalsePositives int            `json:"false_positives"`
	HashBuckets    map[string]int `json:"hash_buckets"`
}

type StatusResponse struct {
	Version  string                  `json:"version"`
	Patterns PatternsResponse        `json:"patterns"`
	Stats    CollisionStatsResponse  `json:"stats"`
}

type HashRequest struct {
	Text   string `json:"text"`
	Length int    `json:"length,omitempty"`
}

type HashResponse struct {
	Hash    string `json:"hash"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
