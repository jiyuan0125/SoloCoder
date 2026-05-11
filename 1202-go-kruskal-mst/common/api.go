package common

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Cost int    `json:"cost"`
}

type TopologyRequest struct {
	Nodes []string `json:"nodes"`
	Edges []Edge   `json:"edges"`
}

type MSTResponse struct {
	Success     bool   `json:"success"`
	Edges       []Edge `json:"edges,omitempty"`
	TotalCost   int    `json:"total_cost,omitempty"`
	Message     string `json:"message,omitempty"`
	Isolated    []string `json:"isolated,omitempty"`
}
