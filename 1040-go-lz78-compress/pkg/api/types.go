package api

type CompressRequest struct {
	Text string `json:"text"`
}

type CompressResponse struct {
	Success bool   `json:"success"`
	Indexes []int  `json:"indexes"`
	Error   string `json:"error,omitempty"`
}

type DecompressRequest struct {
	Indexes []int `json:"indexes"`
}

type DecompressResponse struct {
	Success bool   `json:"success"`
	Text    string `json:"text"`
	Error   string `json:"error,omitempty"`
}

type DictStatusResponse struct {
	Success bool              `json:"success"`
	Count   int               `json:"count"`
	Entries map[int]string    `json:"entries"`
	Error   string            `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
