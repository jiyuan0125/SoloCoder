package common

type AddRequest struct {
	Item string `json:"item"`
}

type AddBatchRequest struct {
	Items []string `json:"items"`
}

type CheckResponse struct {
	Exists bool `json:"exists"`
}

type StatsResponse struct {
	Count      uint64  `json:"count"`
	Capacity   uint64  `json:"capacity"`
	CurrentFPR float64 `json:"current_fpr"`
	BitUsage   float64 `json:"bit_usage"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
