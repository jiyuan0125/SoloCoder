package api

type CreateRequest struct {
	Size int `json:"size"`
}

type CreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RangeAddRequest struct {
	L   int   `json:"l"`
	R   int   `json:"r"`
	Val int64 `json:"val"`
}

type RangeAddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RangeSumRequest struct {
	L int `json:"l"`
	R int `json:"r"`
}

type RangeSumResponse struct {
	Success bool   `json:"success"`
	Sum     int64  `json:"sum"`
	Message string `json:"message"`
}

type SetRequest struct {
	Index int   `json:"index"`
	Val   int64 `json:"val"`
}

type SetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetRequest struct {
	Index int `json:"index"`
}

type GetResponse struct {
	Success bool   `json:"success"`
	Value   int64  `json:"value"`
	Message string `json:"message"`
}

type DumpRequest struct {
}

type DumpResponse struct {
	Success bool    `json:"success"`
	Data    []int64 `json:"data"`
	Message string  `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
