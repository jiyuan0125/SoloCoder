package api

type ConvertRequest struct {
	Expression string `json:"expression"`
}

type ConvertResponse struct {
	Success bool   `json:"success"`
	RPN     string `json:"rpn"`
	Error   string `json:"error,omitempty"`
}

type EvaluateRequest struct {
	Expression string `json:"expression"`
}

type EvaluateResponse struct {
	Success bool    `json:"success"`
	RPN     string  `json:"rpn"`
	Result  float64 `json:"result"`
	ResultStr string `json:"result_str"`
	Error   string  `json:"error,omitempty"`
}

type SetVariableRequest struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type SetVariableResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetVariableRequest struct {
	Name string `json:"name"`
}

type GetVariableResponse struct {
	Success bool    `json:"success"`
	Value   float64 `json:"value"`
	Error   string  `json:"error,omitempty"`
}

type ListVariablesResponse struct {
	Success   bool              `json:"success"`
	Variables map[string]float64 `json:"variables"`
	Error     string            `json:"error,omitempty"`
}
