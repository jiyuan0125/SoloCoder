package api

type CreateRequest struct {
	Size int `json:"size"`
}

type CreateResponse struct {
	ID string `json:"id"`
}

type UpdateRequest struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Delta int64  `json:"delta"`
}

type UpdateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type QueryPrefixRequest struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
}

type QueryPrefixResponse struct {
	Result int64  `json:"result"`
	Error  string `json:"error,omitempty"`
}

type QueryRangeRequest struct {
	ID string `json:"id"`
	L  int    `json:"l"`
	R  int    `json:"r"`
}

type QueryRangeResponse struct {
	Result int64  `json:"result"`
	Error  string `json:"error,omitempty"`
}

type InversionsRequest struct {
	Array []int64 `json:"array"`
}

type InversionsResponse struct {
	Count int64  `json:"count"`
	Error string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
