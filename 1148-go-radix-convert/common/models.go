package common

type ConvertRequest struct {
	Input     string `json:"input"`
	FromBase  int    `json:"from_base"`
	ToBase    int    `json:"to_base"`
	Precision int    `json:"precision"`
}

type ConvertResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
