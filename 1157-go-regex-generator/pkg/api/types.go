package api

type GenerateRequest struct {
	Examples      []string          `json:"examples"`
	ShortestMatch bool              `json:"shortest_match,omitempty"`
	AllowOptional bool              `json:"allow_optional,omitempty"`
	Options       map[string]string `json:"options,omitempty"`
}

type GenerateResponse struct {
	Regex       string   `json:"regex"`
	Explanation string   `json:"explanation"`
	Examples    []string `json:"examples_used,omitempty"`
}

type ValidateRequest struct {
	Regex    string `json:"regex"`
	TestStr  string `json:"test_str"`
}

type ValidateResponse struct {
	Matched bool `json:"matched"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
