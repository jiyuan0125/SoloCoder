package api

type GenerateRequest struct {
	Source    string `json:"source"`
	Interface string `json:"interface"`
}

type GenerateResponse struct {
	Code string `json:"code"`
	Err  string `json:"err"`
}

type ParseResponse struct {
	Interfaces []string `json:"interfaces"`
	Err        string   `json:"err"`
}
