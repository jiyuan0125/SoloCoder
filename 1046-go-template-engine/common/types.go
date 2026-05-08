package common

type RenderRequest struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
	Options  *RenderOptions         `json:"options,omitempty"`
}

type RenderResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type TemplateRegisterRequest struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

type TemplateRegisterResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type TemplateRenderByNameRequest struct {
	Name    string                 `json:"name"`
	Data    map[string]interface{} `json:"data"`
	Options *RenderOptions         `json:"options,omitempty"`
}

type TemplateListResponse struct {
	Success bool              `json:"success"`
	Templates []TemplateInfo  `json:"templates,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type TemplateInfo struct {
	Name string `json:"name"`
}

type RenderOptions struct {
	StrictMissing bool `json:"strict_missing,omitempty"`
}
