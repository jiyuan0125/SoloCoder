package common

type GraphData struct {
	Nodes []string  `json:"nodes"`
	Edges []Edge    `json:"edges"`
}

type Edge struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Weight float64 `json:"weight"`
}

type ImportGraphResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ShortestPathRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ShortestPathResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message,omitempty"`
	TotalTime   float64  `json:"total_time,omitempty"`
	Path        []string `json:"path,omitempty"`
	EdgeCount   int      `json:"edge_count,omitempty"`
}

type AllShortestPathsRequest struct {
	Start string `json:"start"`
}

type AllShortestPathsResponse struct {
	Success     bool                `json:"success"`
	Message     string              `json:"message,omitempty"`
	Distances   map[string]float64  `json:"distances,omitempty"`
}
