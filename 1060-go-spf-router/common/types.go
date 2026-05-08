package common

type Node struct {
	ID string `json:"id"`
}

type Link struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Cost   uint32 `json:"cost"`
	SeqNum uint32 `json:"seq_num"`
}

type TopologyUpdateRequest struct {
	AddNodes    []Node `json:"add_nodes,omitempty"`
	RemoveNodes []string `json:"remove_nodes,omitempty"`
	AddLinks    []Link `json:"add_links,omitempty"`
	RemoveLinks []Link `json:"remove_links,omitempty"`
}

type TopologyUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SPFRunRequest struct {
	Source string `json:"source"`
}

type SPFRunResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ShortestPathQueryRequest struct {
	Source string `json:"source"`
}

type PathEntry struct {
	Destination string   `json:"destination"`
	Paths       [][]string `json:"paths"`
	TotalCost   uint32   `json:"total_cost"`
	Valid       bool     `json:"valid"`
}

type ShortestPathQueryResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Paths   []PathEntry `json:"paths,omitempty"`
}
