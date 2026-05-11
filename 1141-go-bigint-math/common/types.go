package common

type Operation string

const (
	OpAdd      Operation = "add"
	OpSubtract Operation = "subtract"
	OpMultiply Operation = "multiply"
	OpDivide   Operation = "divide"
)

type Request struct {
	Operation Operation `json:"operation"`
	A         string    `json:"a"`
	B         string    `json:"b"`
}

type Response struct {
	Success   bool   `json:"success"`
	Result    string `json:"result,omitempty"`
	Quotient  string `json:"quotient,omitempty"`
	Remainder string `json:"remainder,omitempty"`
	Error     string `json:"error,omitempty"`
}
