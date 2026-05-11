package common

type GenerateRequest struct {
	YAMLContent string `json:"yaml_content"`
	PackageName string `json:"package_name"`
}

type GenerateResponse struct {
	Success  bool     `json:"success"`
	Code     string   `json:"code,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}
