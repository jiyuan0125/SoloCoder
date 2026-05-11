package proto

type ParseRequest struct {
	FormatID string `json:"format_id"`
	Data     []byte `json:"data"`
}

type ParseResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type SerializeRequest struct {
	FormatID string                 `json:"format_id"`
	Data     map[string]interface{} `json:"data"`
}

type SerializeResponse struct {
	Success bool   `json:"success"`
	Data    []byte `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ValidateRequest struct {
	DSL string `json:"dsl"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type RegisterFormatRequest struct {
	FormatID string `json:"format_id"`
	DSL      string `json:"dsl"`
}

type RegisterFormatResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
