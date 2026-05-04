package protocol

type SubmitRequest struct {
	Data []string `json:"data"`
}

type SubmitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type FlushRequest struct {
}

type FlushResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsRequest struct {
}

type StatsResponse struct {
	Success    bool   `json:"success"`
	BufferSize int    `json:"buffer_size"`
	Message    string `json:"message,omitempty"`
}

type ShutdownRequest struct {
}

type ShutdownResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
