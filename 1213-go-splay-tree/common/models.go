package common

type PutRequest struct {
	Key   int64 `json:"key"`
	Value int64 `json:"value"`
}

type PutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GetRequest struct {
	Key int64 `json:"key"`
}

type GetResponse struct {
	Success bool   `json:"success"`
	Value   int64  `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

type DeleteRequest struct {
	Key int64 `json:"key"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RangeQueryRequest struct {
	Left  int64 `json:"left"`
	Right int64 `json:"right"`
}

type RangeQueryResponse struct {
	Success bool     `json:"success"`
	Sum     int64    `json:"sum,omitempty"`
	Keys    []int64  `json:"keys,omitempty"`
	Values  []int64  `json:"values,omitempty"`
	Message string   `json:"message,omitempty"`
}

type RangeAddRequest struct {
	Left  int64 `json:"left"`
	Right int64 `json:"right"`
	Delta int64 `json:"delta"`
}

type RangeAddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RangeSetRequest struct {
	Left  int64 `json:"left"`
	Right int64 `json:"right"`
	Value int64 `json:"value"`
}

type RangeSetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
