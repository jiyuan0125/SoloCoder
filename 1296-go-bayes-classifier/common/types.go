package common

type TrainingSample struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

type TrainRequest struct {
	Samples []TrainingSample `json:"samples"`
}

type TrainResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	TotalDocs int    `json:"total_docs,omitempty"`
	Classes   int    `json:"classes,omitempty"`
}

type ClassifyRequest struct {
	Text string `json:"text"`
}

type ClassifyResult struct {
	Label string  `json:"label"`
	Prob  float64 `json:"probability"`
}

type ClassifyResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message,omitempty"`
	BestLabel string            `json:"best_label,omitempty"`
	Results   []ClassifyResult  `json:"results,omitempty"`
}

type ExportModelResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	ModelFile string `json:"model_file,omitempty"`
}

type ImportModelRequest struct {
	ModelFile string `json:"model_file"`
}

type ImportModelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
