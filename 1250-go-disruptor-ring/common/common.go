package common

const (
	MaxBufferSize = 1 << 30

	ErrBufferExists   = "buffer already exists"
	ErrBufferNotFound = "buffer not found"
	ErrBufferFull     = "buffer is full"
	ErrNoData         = "no data available"
	ErrInvalidSize    = "invalid buffer size"
)

type BufferInfo struct {
	Name           string            `json:"name"`
	Capacity       int64             `json:"capacity"`
	Used           int64             `json:"used"`
	ProducerCursor int64             `json:"producer_cursor"`
	Consumers      map[string]int64  `json:"consumers"`
}

type CreateBufferRequest struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type CreateBufferResponse struct {
	Name     string `json:"name"`
	Capacity int64  `json:"capacity"`
}

type DestroyBufferRequest struct {
	Name string `json:"name"`
}

type PublishRequest struct {
	BufferName string `json:"buffer_name"`
	Message    string `json:"message"`
}

type PublishResponse struct {
	Sequence int64 `json:"sequence"`
}

type BatchPublishRequest struct {
	BufferName string   `json:"buffer_name"`
	Messages   []string `json:"messages"`
}

type BatchPublishResponse struct {
	StartSequence int64 `json:"start_sequence"`
	Count         int   `json:"count"`
}

type ConsumeRequest struct {
	BufferName string `json:"buffer_name"`
	ConsumerID string `json:"consumer_id"`
}

type ConsumeResponse struct {
	Sequence int64  `json:"sequence"`
	Message  string `json:"message"`
}

type BatchConsumeRequest struct {
	BufferName string `json:"buffer_name"`
	ConsumerID string `json:"consumer_id"`
	MaxCount   int    `json:"max_count"`
}

type BatchConsumeResponse struct {
	Messages []ConsumeResponse `json:"messages"`
}

type BufferInfoResponse struct {
	Name           string            `json:"name"`
	Capacity       int64             `json:"capacity"`
	Used           int64             `json:"used"`
	ProducerCursor int64             `json:"producer_cursor"`
	Consumers      map[string]int64  `json:"consumers"`
}

type ListBufferResponse struct {
	Buffers []string `json:"buffers"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
