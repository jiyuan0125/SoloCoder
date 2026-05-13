package service

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func ValidateExpression(expression string) error {
	if expression == "" {
		return nil
	}
	_, err := expr.Compile(expression, expr.AsBool())
	return err
}

func EvaluateExpression(expression string, data map[string]interface{}) (bool, error) {
	if expression == "" {
		return true, nil
	}

	var program *vm.Program
	var err error

	program, err = expr.Compile(expression, expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("expression compilation error: %w", err)
	}

	result, err := expr.Run(program, data)
	if err != nil {
		return false, fmt.Errorf("expression evaluation error: %w", err)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression must return boolean value")
	}

	return boolResult, nil
}
