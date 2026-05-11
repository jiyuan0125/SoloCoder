package api

type CreateHeapRequest struct {
	HeapType string `json:"heap_type"`
}

type CreateHeapResponse struct {
	HeapID string `json:"heap_id"`
	Error  string `json:"error,omitempty"`
}

type InsertRequest struct {
	HeapID string `json:"heap_id"`
	Value  int64  `json:"value"`
}

type InsertResponse struct {
	Error string `json:"error,omitempty"`
}

type TopRequest struct {
	HeapID string `json:"heap_id"`
}

type TopResponse struct {
	Value *int64 `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type PopRequest struct {
	HeapID string `json:"heap_id"`
}

type PopResponse struct {
	Value *int64 `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type MergeRequest struct {
	HeapID1 string `json:"heap_id1"`
	HeapID2 string `json:"heap_id2"`
}

type MergeResponse struct {
	NewHeapID string `json:"new_heap_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
