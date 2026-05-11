package common

type InsertRequest struct {
	Key   int64 `json:"key"`
	Value int64 `json:"value"`
}

type SearchRequest struct {
	Key int64 `json:"key"`
}

type SearchResponse struct {
	Success bool  `json:"success"`
	Value   int64 `json:"value,omitempty"`
	Error   string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Key int64 `json:"key"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type SizeResponse struct {
	Size int `json:"size"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
