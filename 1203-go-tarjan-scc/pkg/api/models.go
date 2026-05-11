package api

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GraphRequest struct {
	Nodes []string `json:"nodes"`
	Edges []Edge   `json:"edges"`
}

type SCCResponseItem struct {
	Nodes   []string `json:"nodes"`
	IsCycle bool     `json:"is_cycle"`
}

type GraphResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	SCCs    []SCCResponseItem `json:"sccs,omitempty"`
}
