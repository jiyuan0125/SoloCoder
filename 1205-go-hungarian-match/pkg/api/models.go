package api

type MatchRequest struct {
	Interns   []string `json:"interns"`
	Positions []string `json:"positions"`
	Edges     []Edge   `json:"edges"`
}

type Edge struct {
	Intern   string `json:"intern"`
	Position string `json:"position"`
}

type MatchResponse struct {
	MaxMatch    int          `json:"max_match"`
	Assignments []Assignment `json:"assignments"`
}

type Assignment struct {
	Intern   string `json:"intern"`
	Position string `json:"position"`
}

type QueryRequest struct {
	Intern string `json:"intern"`
}

type QueryResponse struct {
	Matched  bool   `json:"matched"`
	Position string `json:"position,omitempty"`
}
