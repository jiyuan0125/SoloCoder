package api

type DetectRequest struct {
	Content []byte `json:"content"`
}

type DetectResponse struct {
	Encoding string  `json:"encoding"`
	Confidence float64 `json:"confidence"`
	HasBOM   bool    `json:"has_bom"`
}

type ConvertRequest struct {
	Content    []byte `json:"content"`
	FromEncoding string `json:"from_encoding"`
	ToEncoding   string `json:"to_encoding"`
}

type ConvertResponse struct {
	Content []byte `json:"content"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
