package api

type CompressRequest struct {
	Data     []byte `json:"data"`
	Quality  int    `json:"quality"`
	IsBase64 bool   `json:"is_base64"`
}

type CompressResponse struct {
	Data         []byte `json:"data"`
	OriginalSize int    `json:"original_size"`
	CompressedSize int  `json:"compressed_size"`
	Quality      int    `json:"quality"`
}

type DecompressRequest struct {
	Data     []byte `json:"data"`
	IsBase64 bool   `json:"is_base64"`
}

type DecompressResponse struct {
	Data         []byte `json:"data"`
	OriginalSize int    `json:"original_size"`
	CompressedSize int  `json:"compressed_size"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ValidateQuality(quality int) int {
	if quality < 1 {
		return 1
	}
	if quality > 11 {
		return 11
	}
	return quality
}
