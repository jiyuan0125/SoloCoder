package common

type GenerateRequest struct {
	TomlContent string `json:"toml_content"`
	PackageName string `json:"package_name"`
}

type GenerateResponse struct {
	Code    string `json:"code"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}
