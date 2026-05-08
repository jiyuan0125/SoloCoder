package eval

import "strings"

type Result struct {
	Success bool
	Value   interface{}
	Type    string
	Errors  []string
}

func ParseAndEvaluate(expression string, vars Variables) *Result {
	lexer := NewLexer(expression)
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	if len(parser.Errors()) > 0 {
		return &Result{
			Success: false,
			Errors:  parser.Errors(),
		}
	}
	result := Eval(program, vars)
	if result == nil {
		return &Result{
			Success: false,
			Errors:  []string{"evaluation returned nil"},
		}
	}
	if result.Type == VALUE_ERROR {
		return &Result{
			Success: false,
			Errors:  []string{result.Msg},
		}
	}
	return buildSuccessResult(result)
}

func ParseAndEvaluateWithVariableCheck(expression string, vars Variables) *Result {
	undefinedVars := findUndefinedVariables(expression, vars)
	if len(undefinedVars) > 0 {
		return &Result{
			Success: false,
			Errors:  []string{"undefined variables: " + strings.Join(undefinedVars, ", ")},
		}
	}
	return ParseAndEvaluate(expression, vars)
}

func findUndefinedVariables(expression string, vars Variables) []string {
	lexer := NewLexer(expression)
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	if len(parser.Errors()) > 0 {
		return nil
	}
	identifiers := collectIdentifiers(program.Expression)
	undefined := []string{}
	for _, ident := range identifiers {
		if _, ok := vars[ident]; !ok {
			if !isBuiltinFunction(ident) {
				undefined = append(undefined, ident)
			}
		}
	}
	return undefined
}

func collectIdentifiers(expr Expression) []string {
	if expr == nil {
		return []string{}
	}
	idents := []string{}
	switch node := expr.(type) {
	case *Identifier:
		idents = append(idents, node.Value)
	case *PrefixExpression:
		idents = append(idents, collectIdentifiers(node.Right)...)
	case *InfixExpression:
		idents = append(idents, collectIdentifiers(node.Left)...)
		idents = append(idents, collectIdentifiers(node.Right)...)
	case *CallExpression:
		idents = append(idents, collectIdentifiers(node.Function)...)
		for _, arg := range node.Arguments {
			idents = append(idents, collectIdentifiers(arg)...)
		}
	}
	return idents
}

func buildSuccessResult(v *Value) *Result {
	result := &Result{Success: true}
	switch v.Type {
	case VALUE_INT:
		result.Value = v.Int
		result.Type = "int"
	case VALUE_FLOAT:
		result.Value = v.Float
		result.Type = "float"
	case VALUE_BOOL:
		result.Value = v.Bool
		result.Type = "bool"
	}
	return result
}
