package common

type PluginInfo struct {
	Name         string   `json:"name"`
	APIVersion   string   `json:"api_version"`
	Capabilities []string `json:"capabilities"`
	Loaded       bool     `json:"loaded"`
}

type ListResponse struct {
	Plugins []PluginInfo `json:"plugins"`
}

type ExecuteRequest struct {
	Name  string `json:"name"`
	Input []byte `json:"input"`
}

type ExecuteResponse struct {
	Name   string `json:"name"`
	Output []byte `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ScanResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
