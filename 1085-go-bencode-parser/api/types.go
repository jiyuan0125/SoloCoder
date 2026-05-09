package api

type DecodeRequest struct {
	Data string `json:"data"`
}

type DecodeResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

type EncodeRequest struct {
	Data interface{} `json:"data"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Result  string `json:"result,omitempty"`
}

type InfoRequest struct {
	Data string `json:"data"`
}

type InfoResponse struct {
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	OuterType     string `json:"outer_type,omitempty"`
	DictKeyCount  int    `json:"dict_key_count,omitempty"`
	ListElemCount int    `json:"list_elem_count,omitempty"`
	MaxDepth      int    `json:"max_depth,omitempty"`
}
