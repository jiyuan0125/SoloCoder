package shared

type ErrorResponse struct {
	Error string `json:"error"`
}

type RegisterPrototypeRequest struct {
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`
}

type RegisterPrototypeResponse struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
}

type ListPrototypesResponse struct {
	Prototypes []string `json:"prototypes"`
}

type CloneRequest struct {
	Name string `json:"name"`
}

type CloneResponse struct {
	Success bool                   `json:"success"`
	Name    string                 `json:"name,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type DeepCopyRequest struct {
	Data map[string]interface{} `json:"data"`
}

type DeepCopyResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type CompareRequest struct {
	A    map[string]interface{} `json:"a"`
	B    map[string]interface{} `json:"b"`
	Mode string                 `json:"mode"`
}

type CompareResponse struct {
	Success bool   `json:"success"`
	Equal   bool   `json:"equal"`
	Mode    string `json:"mode"`
	Error   string `json:"error,omitempty"`
}
