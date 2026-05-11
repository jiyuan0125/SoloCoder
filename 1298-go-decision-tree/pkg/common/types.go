package common

type TrainRequest struct {
	CSVData       string `json:"csv_data"`
	MaxDepth      int    `json:"max_depth"`
	MinSamples    int    `json:"min_samples"`
	HandleMissing string `json:"handle_missing"`
}

type TrainResponse struct {
	Success bool   `json:"success"`
	TreeID  string `json:"tree_id"`
	Message string `json:"message"`
}

type PredictRequest struct {
	TreeID   string            `json:"tree_id"`
	Features []string          `json:"features"`
	Data     [][]string        `json:"data"`
}

type PredictResponse struct {
	Success bool     `json:"success"`
	Labels  []string `json:"labels"`
	Message string   `json:"message"`
}

type TreeExportRequest struct {
	TreeID string `json:"tree_id"`
}

type TreeExportResponse struct {
	Success   bool   `json:"success"`
	JSONTree  string `json:"json_tree"`
	TextTree  string `json:"text_tree"`
	Message   string `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
