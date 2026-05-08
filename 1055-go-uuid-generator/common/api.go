package common

type GenerateRequest struct {
	Version   int    `json:"version"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Count     int    `json:"count,omitempty"`
}

type GenerateResponse struct {
	UUIDs []string `json:"uuids"`
	Error string   `json:"error,omitempty"`
}

type ParseRequest struct {
	UUID string `json:"uuid"`
}

type ParseResponse struct {
	Version int    `json:"version"`
	Variant int    `json:"variant"`
	Error   string `json:"error,omitempty"`
}
