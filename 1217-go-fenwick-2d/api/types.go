package api

type CreateRequest struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

type CreateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type UpdateRequest struct {
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Delta int64 `json:"delta"`
}

type UpdateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type SumRequest struct {
	Lx int `json:"lx"`
	Ly int `json:"ly"`
	Rx int `json:"rx"`
	Ry int `json:"ry"`
}

type SumResponse struct {
	Success bool   `json:"success"`
	Sum     int64  `json:"sum,omitempty"`
	Error   string `json:"error,omitempty"`
}

type InitRequest struct {
	Matrix [][]int64 `json:"matrix"`
}

type InitResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
