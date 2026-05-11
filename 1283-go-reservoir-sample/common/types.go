package common

type AddDataRequest struct {
	Data string `json:"data"`
}

type AddDataResponse struct {
	Success bool `json:"success"`
}

type SampleRequest struct {
	K *int `json:"k,omitempty"`
}

type SampleResponse struct {
	Success bool     `json:"success"`
	Samples []string `json:"samples"`
}

type SetKRequest struct {
	K int `json:"k"`
}

type SetKResponse struct {
	Success bool `json:"success"`
	K       int  `json:"k"`
}

type CountResponse struct {
	Success bool `json:"success"`
	Count   int  `json:"count"`
}

type ClearResponse struct {
	Success bool `json:"success"`
}
