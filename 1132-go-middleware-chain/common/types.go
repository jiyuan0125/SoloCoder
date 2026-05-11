package common

type Request struct {
	ID       string            `json:"id"`
	Pipeline string            `json:"pipeline"`
	Header   map[string]string `json:"header"`
	Body     string            `json:"body"`
}

type Response struct {
	StatusCode int               `json:"status_code"`
	Header     map[string]string `json:"header"`
	Body       string            `json:"body"`
}

type ExecutionResult struct {
	RequestID    string                `json:"request_id"`
	Pipeline     string                `json:"pipeline"`
	Status       string                `json:"status"`
	StartTime    int64                 `json:"start_time"`
	EndTime      int64                 `json:"end_time"`
	DurationMs   int64                 `json:"duration_ms"`
	HandlerResults []HandlerResult     `json:"handler_results"`
	Response     *Response             `json:"response"`
	Error        string                `json:"error,omitempty"`
}

type HandlerResult struct {
	HandlerID   string `json:"handler_id"`
	HandlerName string `json:"handler_name"`
	Status      string `json:"status"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	DurationMs  int64  `json:"duration_ms"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	Panic       bool   `json:"panic,omitempty"`
}

type CreatePipelineRequest struct {
	Name string `json:"name"`
}

type AddHandlerRequest struct {
	Pipeline string `json:"pipeline"`
	HandlerID string `json:"handler_id"`
	Position int `json:"position"`
}

type RemoveHandlerRequest struct {
	Pipeline string `json:"pipeline"`
	HandlerID string `json:"handler_id"`
}

type ReorderHandlersRequest struct {
	Pipeline    string   `json:"pipeline"`
	HandlerIDs  []string `json:"handler_ids"`
}

type CreateHandlerRequest struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type CreateMiddlewareRequest struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Async    bool   `json:"async"`
	TimeoutMs int64  `json:"timeout_ms"`
}

type TraceQuery struct {
	RequestID string `json:"request_id"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type PipelineInfo struct {
	Name      string   `json:"name"`
	Handlers  []string `json:"handlers"`
}

type HandlerInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	PanicHistory []string `json:"panic_history"`
}

type MiddlewareInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Async     bool   `json:"async"`
	TimeoutMs int64  `json:"timeout_ms"`
}
