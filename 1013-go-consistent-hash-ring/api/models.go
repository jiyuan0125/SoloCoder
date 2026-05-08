package api

type AddNodeRequest struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
}

type AddNodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RemoveNodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type LookupResponse struct {
	Node   string `json:"node,omitempty"`
	Error  string `json:"error,omitempty"`
}

type StatsResponse struct {
	TotalVirtualNodes int               `json:"total_virtual_nodes"`
	Nodes             map[string]NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	Weight           int     `json:"weight"`
	VirtualNodeCount int     `json:"virtual_node_count"`
	Percentage       float64 `json:"percentage"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
