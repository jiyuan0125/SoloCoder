package api

type CompressRequest struct {
	Data string `json:"data"`
}

type CompressResponse struct {
	Data string `json:"data"`
}

type DecompressRequest struct {
	Data string `json:"data"`
}

type DecompressResponse struct {
	Data string `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
