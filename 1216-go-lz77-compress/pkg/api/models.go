package api

type CompressRequest struct {
	Text string `json:"text"`
}

type CompressResponse struct {
	CompressedData string  `json:"compressed_data"`
	OriginalSize   int     `json:"original_size"`
	CompressedSize int     `json:"compressed_size"`
	Ratio          float64 `json:"ratio"`
	Success        bool    `json:"success"`
	Message        string  `json:"message,omitempty"`
}

type DecompressRequest struct {
	CompressedData string `json:"compressed_data"`
}

type DecompressResponse struct {
	Text    string `json:"text"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
