package api

const (
	MaxBatchCount    = 1000
	ReservedNodeID   = 0
	MaxNodeID        = 1023
	ClockBackoffMax  = 5
	EpochYear        = 2024
	EpochMonth       = 1
	EpochDay         = 1
	DefaultPort      = 8080
)

type RegisterNodeRequest struct {
	Name    string `json:"name"`
	Remark  string `json:"remark,omitempty"`
}

type RegisterNodeResponse struct {
	Success bool   `json:"success"`
	NodeID  uint16 `json:"node_id"`
	Message string `json:"message,omitempty"`
}

type UnregisterNodeRequest struct {
	NodeID uint16 `json:"node_id"`
}

type UnregisterNodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GenerateIDRequest struct {
	NodeID uint16 `json:"node_id"`
}

type GenerateIDResponse struct {
	Success bool   `json:"success"`
	ID      int64  `json:"id"`
	Message string `json:"message,omitempty"`
}

type BatchGenerateRequest struct {
	NodeID uint16 `json:"node_id"`
	Count  int    `json:"count"`
}

type BatchGenerateResponse struct {
	Success bool    `json:"success"`
	IDs     []int64 `json:"ids,omitempty"`
	Message string  `json:"message,omitempty"`
}

type NodeInfo struct {
	NodeID     uint16 `json:"node_id"`
	Name       string `json:"name"`
	Remark     string `json:"remark,omitempty"`
	RegisteredAt int64 `json:"registered_at"`
}

type ListNodesResponse struct {
	Success bool        `json:"success"`
	Nodes   []*NodeInfo `json:"nodes"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
