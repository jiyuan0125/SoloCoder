package rpn

type ParseError struct {
	Msg string
}

func (e *ParseError) Error() string {
	return "parse error: " + e.Msg
}

type EvalError struct {
	Msg string
}

func (e *EvalError) Error() string {
	return "evaluation error: " + e.Msg
}

type FunctionError struct {
	Msg string
}

func (e *FunctionError) Error() string {
	return "function error: " + e.Msg
}

type VariableError struct {
	Msg string
}

func (e *VariableError) Error() string {
	return "variable error: " + e.Msg
}
