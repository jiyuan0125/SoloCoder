package api

type BinaryOpRequest struct {
	A []float64 `json:"a"`
	B []float64 `json:"b"`
}

type EvalRequest struct {
	Polynomial []float64 `json:"polynomial"`
	X          float64   `json:"x"`
}

type SinglePolynomialRequest struct {
	Polynomial []float64 `json:"polynomial"`
}

type PolynomialResponse struct {
	Success bool       `json:"success"`
	Result  []float64  `json:"result,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type DivResponse struct {
	Success   bool       `json:"success"`
	Quotient  []float64  `json:"quotient,omitempty"`
	Remainder []float64  `json:"remainder,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type EvalResponse struct {
	Success bool    `json:"success"`
	Result  float64 `json:"result,omitempty"`
	Error   string  `json:"error,omitempty"`
}

type FactorResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
