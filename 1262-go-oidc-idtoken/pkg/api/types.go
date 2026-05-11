package api

type VerifyRequest struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Valid  bool                   `json:"valid"`
	Claims map[string]interface{} `json:"claims"`
	Errors []string               `json:"errors,omitempty"`
}

type ConfigRequest struct {
	Issuer   string `json:"issuer"`
	ClientID string `json:"client_id"`
	Secret   string `json:"secret"`
}

type ConfigResponse struct {
	Success bool `json:"success"`
}

type ParseRequest struct {
	Token string `json:"token"`
}

type ParseResponse struct {
	Header map[string]interface{} `json:"header"`
	Claims map[string]interface{} `json:"claims"`
}

type SetConfigRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetConfigResponse struct {
	Success bool `json:"success"`
}
