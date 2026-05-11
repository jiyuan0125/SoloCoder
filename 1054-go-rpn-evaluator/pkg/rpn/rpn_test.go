package rpn

import (
	"testing"
)

func TestFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{"abs", "abs(5)", "5 abs/1"},
		{"abs negative", "abs(-5)", "5 u- abs/1"},
		{"max", "max(3, 7)", "3 7 max/2"},
		{"min", "min(3, 7)", "3 7 min/2"},
		{"max three", "max(1, 5, 3)", "1 5 3 max/3"},
		{"basic paren", "(3+4)*2", "3 4 + 2 *"},
	}

	vars := NewVariableStore()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			terms, err := InfixToRPN(tc.expr)
			if err != nil {
				t.Fatalf("InfixToRPN(%q) error: %v", tc.expr, err)
			}

			got := RPNToString(terms)
			t.Logf("expr: %s, rpn: %s", tc.expr, got)

			result, err := EvaluateRPN(terms, vars)
			if err != nil {
				t.Logf("Evaluation error (may be expected): %v", err)
			} else {
				t.Logf("Result: %v, formatted: %s", result, FormatResult(result))
			}
		})
	}
}

func TestEvaluateFunctions(t *testing.T) {
	vars := NewVariableStore()

	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"abs positive", "abs(5)", 5},
		{"abs negative", "abs(-5)", 5},
		{"max", "max(3, 7)", 7},
		{"min", "min(3, 7)", 3},
		{"max three", "max(1, 5, 3)", 5},
		{"min three", "min(1, 5, 3)", 1},
		{"max with expr", "max(3+2, 7-1)", 6},
		{"nested abs", "abs(-3 + 5)", 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := EvaluateInfix(tc.expr, vars)
			if err != nil {
				t.Fatalf("EvaluateInfix(%q) error: %v", tc.expr, err)
			}
			if result != tc.expected {
				t.Errorf("EvaluateInfix(%q) = %v, want %v", tc.expr, result, tc.expected)
			}
		})
	}
}
