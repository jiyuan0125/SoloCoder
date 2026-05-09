package api

const (
	EndpointCompress = "/compress"
	EndpointDecompress = "/decompress"
)

type CompressRequest struct {
	Data        []byte `json:"data"`
	MinCodeSize int    `json:"minCodeSize"`
}

type CompressResponse struct {
	Data        []byte `json:"data"`
	MinCodeSize int    `json:"minCodeSize"`
	Error       string `json:"error,omitempty"`
}

type DecompressRequest struct {
	Data        []byte `json:"data"`
	MinCodeSize int    `json:"minCodeSize"`
}

type DecompressResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error,omitempty"`
}
