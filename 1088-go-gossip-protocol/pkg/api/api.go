package api

type AddNodeRequest struct {
	NodeID string `json:"node_id"`
}

type RemoveNodeRequest struct {
	NodeID string `json:"node_id"`
}

type InjectDataRequest struct {
	NodeID string `json:"node_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type StepRequest struct {
	Steps int `json:"steps,omitempty"`
}

type StepResponse struct {
	Success   bool    `json:"success"`
	StepsDone int     `json:"steps_done"`
	Consistency float64 `json:"consistency,omitempty"`
}

type ConsistencyResponse struct {
	Consistency float64 `json:"consistency"`
}

type NodeStorageItem struct {
	Value   string `json:"value"`
	Version int64  `json:"version"`
}

type NodeState struct {
	Status  string                    `json:"status"`
	Storage map[string]NodeStorageItem `json:"storage"`
}

type StateResponse struct {
	Round int                    `json:"round"`
	Nodes map[string]NodeState    `json:"nodes"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SimulateFailureRequest struct {
	NodeID string `json:"node_id"`
}

type SimulateRecoveryRequest struct {
	NodeID string `json:"node_id"`
}
