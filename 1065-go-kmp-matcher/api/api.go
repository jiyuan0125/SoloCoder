package api

type SingleMatchRequest struct {
	Text            string `json:"text"`
	Pattern         string `json:"pattern"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty"`
}

type SingleMatchResponse struct {
	Positions []int `json:"positions"`
}

type MultiMatchRequest struct {
	Text            string   `json:"text"`
	Patterns        []string `json:"patterns"`
	CaseInsensitive bool     `json:"case_insensitive,omitempty"`
}

type MultiMatchResponse struct {
	Results map[string][]int `json:"results"`
}
