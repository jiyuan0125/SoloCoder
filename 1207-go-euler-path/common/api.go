package common

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EulerPathRequest struct {
	Nodes []string `json:"nodes"`
	Edges []Edge   `json:"edges"`
	Start string   `json:"start,omitempty"`
}

type EulerPathResponse struct {
	HasEulerPath    bool     `json:"has_euler_path"`
	HasEulerCircuit bool     `json:"has_euler_circuit"`
	Path            []string `json:"path,omitempty"`
	Error           string   `json:"error,omitempty"`
}
