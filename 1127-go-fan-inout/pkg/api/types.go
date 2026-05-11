package api

type BackpressurePolicy string

const (
	PolicyBlock BackpressurePolicy = "block"
	PolicyDrop  BackpressurePolicy = "drop"
)

type AddSourceRequest struct {
	Name       string `json:"name"`
	IntervalMs int    `json:"interval_ms"`
	Payload    string `json:"payload"`
}

type AddSourceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RemoveSourceRequest struct {
	Name string `json:"name"`
}

type RemoveSourceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SetPolicyRequest struct {
	Policy BackpressurePolicy `json:"policy"`
}

type SetPolicyResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Policy  BackpressurePolicy `json:"policy,omitempty"`
}

type FlowStats struct {
	SourceName       string `json:"source_name"`
	TotalSent        int64  `json:"total_sent"`
	TotalReceived    int64  `json:"total_received"`
	Dropped          int64  `json:"dropped"`
	CurrentBacklog   int    `json:"current_backlog"`
	IsActive         bool   `json:"is_active"`
}

type FlowResponse struct {
	Sources []FlowStats `json:"sources"`
}

type ResultItem struct {
	Source    string `json:"source"`
	Payload   string `json:"payload"`
	Timestamp int64  `json:"timestamp"`
}

type ResultResponse struct {
	Items []ResultItem `json:"items"`
}
