package api

type CreateRequest struct {
	Method   string `json:"method"`
	Capacity int    `json:"capacity"`
}

type CreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GetRequest struct {
	Key string `json:"key"`
}

type GetResponse struct {
	Success bool   `json:"success"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

type RemoveRequest struct {
	Key string `json:"key"`
}

type RemoveResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsResponse struct {
	Success     bool   `json:"success"`
	Size        int    `json:"size,omitempty"`
	Tombstones  int    `json:"tombstones,omitempty"`
	Capacity    int    `json:"capacity,omitempty"`
	MaxProbeLen int    `json:"maxProbeLen,omitempty"`
	Method      string `json:"method,omitempty"`
	Message     string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
