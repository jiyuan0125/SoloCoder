package common

type CreateRequest struct {
	Capacity int `json:"capacity"`
}

type CreateResponse struct {
	ID        string `json:"id"`
	Capacity  int    `json:"capacity"`
	Error     string `json:"error,omitempty"`
}

type WriteRequest struct {
	ID      string `json:"id"`
	Data    []byte `json:"data"`
	Timeout int64  `json:"timeout,omitempty"`
}

type WriteResponse struct {
	Written int    `json:"written"`
	Error   string `json:"error,omitempty"`
}

type ReadRequest struct {
	ID      string `json:"id"`
	Length  int    `json:"length"`
	Timeout int64  `json:"timeout,omitempty"`
}

type ReadResponse struct {
	Data  []byte `json:"data"`
	Read  int    `json:"read"`
	EOF   bool   `json:"eof,omitempty"`
	Error string `json:"error,omitempty"`
}

type StatusRequest struct {
	ID string `json:"id"`
}

type StatusResponse struct {
	ID       string `json:"id"`
	Capacity int    `json:"capacity"`
	Size     int    `json:"size"`
	Closed   bool   `json:"closed"`
	Error    string `json:"error,omitempty"`
}

type CloseRequest struct {
	ID string `json:"id"`
}

type CloseResponse struct {
	Error string `json:"error,omitempty"`
}

type ListResponse struct {
	Buffers []StatusResponse `json:"buffers"`
	Error   string           `json:"error,omitempty"`
}
