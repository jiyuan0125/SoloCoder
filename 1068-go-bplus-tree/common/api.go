package common

type InsertRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type InsertResponse struct {
	Success bool `json:"success"`
}

type SearchRequest struct {
	Key string `json:"key"`
}

type SearchResponse struct {
	Found bool        `json:"found"`
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}

type DeleteResponse struct {
	Success bool `json:"success"`
}

type RangeRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type RangeResponse struct {
	Count int           `json:"count"`
	Items []*KeyValue   `json:"items"`
}

type ScanRequest struct {
	Direction   string `json:"direction"`
	BatchSize   int    `json:"batch_size"`
	StartOffset int    `json:"start_offset"`
}

type ScanResponse struct {
	Count int           `json:"count"`
	Items []*KeyValue   `json:"items"`
	HasMore bool        `json:"has_more"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
