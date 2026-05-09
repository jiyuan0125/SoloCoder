package api

type CreateHLLRequest struct {
	Name      string `json:"name"`
	Precision int    `json:"precision"`
}

type CreateHLLResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type AddRequest struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type AddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type CountRequest struct {
	Name string `json:"name"`
}

type CountResponse struct {
	Success bool    `json:"success"`
	Count   float64 `json:"count,omitempty"`
	Message string  `json:"message,omitempty"`
}

type MergeRequest struct {
	NewName string   `json:"new_name"`
	Sources []string `json:"sources"`
}

type MergeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ListResponse struct {
	Success bool               `json:"success"`
	HLLs    []HLLInstanceInfo `json:"hlls,omitempty"`
	Message string             `json:"message,omitempty"`
}

type HLLInstanceInfo struct {
	Name      string  `json:"name"`
	Precision int     `json:"precision"`
	Count     float64 `json:"count"`
}

type ExportRequest struct {
	Name string `json:"name"`
}

type ExportResponse struct {
	Success bool   `json:"success"`
	Data    []byte `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type ImportRequest struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

type ImportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
