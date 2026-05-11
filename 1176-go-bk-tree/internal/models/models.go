package models

type AddRequest struct {
	Word string `json:"word"`
}

type AddResponse struct {
	Success bool `json:"success"`
}

type ImportRequest struct {
	Words []string `json:"words"`
}

type ImportResponse struct {
	Success  bool `json:"success"`
	Imported int  `json:"imported"`
}

type SearchRequest struct {
	Query       string `json:"query"`
	MaxDistance int    `json:"max_distance"`
}

type SearchMatch struct {
	Word     string `json:"word"`
	Distance int    `json:"distance"`
}

type SearchResponse struct {
	Success bool          `json:"success"`
	Matches []SearchMatch `json:"matches"`
}

type RemoveRequest struct {
	Word string `json:"word"`
}

type RemoveResponse struct {
	Success bool `json:"success"`
	Removed bool `json:"removed"`
}

type InfoResponse struct {
	Success bool `json:"success"`
	Size    int  `json:"size"`
	Depth   int  `json:"depth"`
}
