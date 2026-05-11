package api

type Operation string

const (
	OperationAdd Operation = "add"
	OperationSub Operation = "sub"
	OperationMul Operation = "mul"
	OperationDiv Operation = "div"
	OperationDecimal Operation = "decimal"
)

type OperationRequest struct {
	Operation Operation `json:"operation"`
	Operand1  string    `json:"operand1"`
	Operand2  string    `json:"operand2,omitempty"`
	Precision int       `json:"precision,omitempty"`
}

type FractionResult struct {
	Fraction string `json:"fraction"`
	Integer  string `json:"integer,omitempty"`
	Mixed    string `json:"mixed,omitempty"`
	Decimal  string `json:"decimal,omitempty"`
}

type OperationResponse struct {
	Success bool           `json:"success"`
	Result  *FractionResult `json:"result,omitempty"`
	Error   string         `json:"error,omitempty"`
}
