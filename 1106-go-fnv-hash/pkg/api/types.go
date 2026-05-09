package api

type HashRequest struct {
	Algorithm string `json:"algorithm"`
	Bits      int    `json:"bits"`
	Input     string `json:"input"`
}

type HashResponse struct {
	Success bool   `json:"success"`
	Hash    string `json:"hash,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ConsistentRequest struct {
	Nodes []string `json:"nodes"`
	Key   string   `json:"key"`
}

type ConsistentResponse struct {
	Success bool   `json:"success"`
	Node    string `json:"node,omitempty"`
	Error   string `json:"error,omitempty"`
}

type NodeOperationRequest struct {
	Node string `json:"node"`
}

type NodeOperationResponse struct {
	Success bool     `json:"success"`
	Nodes   []string `json:"nodes,omitempty"`
	Error   string   `json:"error,omitempty"`
}
