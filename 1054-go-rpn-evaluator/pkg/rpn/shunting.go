package rpn

import (
	"strings"
)

type RPNTermType int

const (
	TermNumber RPNTermType = iota
	TermOperator
	TermIdentifier
	TermFunction
)

type RPNTerm struct {
	Type      RPNTermType
	Value     string
	IsUnary   bool
	ArgCount  int
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		if op == "-" {
			return 1
		}
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func isLeftAssoc(op string) bool {
	return op == "+" || op == "-" || op == "*" || op == "/"
}

func InfixToRPN(expr string) ([]RPNTerm, error) {
	tokens, err := Tokenize(expr)
	if err != nil {
		return nil, err
	}

	var output []RPNTerm
	var stack []Token
	argCounts := make(map[int]int)
	parenDepth := 0

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		switch tok.Type {
		case TokenNumber:
			output = append(output, RPNTerm{Type: TermNumber, Value: tok.Value})

		case TokenIdentifier:
			if i+1 < len(tokens) && tokens[i+1].Type == TokenLeftParen {
				stack = append(stack, Token{Type: TokenIdentifier, Value: tok.Value})
				parenDepth++
				argCounts[parenDepth] = 1
				i++
			} else {
				output = append(output, RPNTerm{Type: TermIdentifier, Value: tok.Value})
			}

		case TokenOperator:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type == TokenOperator {
					shouldPop := false
					if tok.IsUnary {
						break
					}
					if isLeftAssoc(tok.Value) {
						if precedence(top.Value) >= precedence(tok.Value) {
							shouldPop = true
						}
					} else {
						if precedence(top.Value) > precedence(tok.Value) {
							shouldPop = true
						}
					}
					if shouldPop {
						output = append(output, RPNTerm{Type: TermOperator, Value: top.Value, IsUnary: top.IsUnary})
						stack = stack[:len(stack)-1]
					} else {
						break
					}
				} else {
					break
				}
			}
			stack = append(stack, tok)

		case TokenLeftParen:
			stack = append(stack, tok)

		case TokenComma:
			foundParen := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type == TokenLeftParen {
					foundParen = true
					break
				}
				output = append(output, RPNTerm{Type: TermOperator, Value: top.Value, IsUnary: top.IsUnary})
				stack = stack[:len(stack)-1]
			}
			if !foundParen {
				return nil, &ParseError{Msg: "misplaced comma or mismatched parentheses"}
			}
			argCounts[parenDepth]++

		case TokenRightParen:
			foundParen := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type == TokenLeftParen {
					foundParen = true
					stack = stack[:len(stack)-1]
					break
				}
				output = append(output, RPNTerm{Type: TermOperator, Value: top.Value, IsUnary: top.IsUnary})
				stack = stack[:len(stack)-1]
			}
			if !foundParen {
				return nil, &ParseError{Msg: "mismatched parentheses"}
			}

			if len(stack) > 0 && stack[len(stack)-1].Type == TokenIdentifier {
				funcName := stack[len(stack)-1].Value
				argCount := argCounts[parenDepth]
				output = append(output, RPNTerm{Type: TermFunction, Value: funcName, ArgCount: argCount})
				stack = stack[:len(stack)-1]
				parenDepth--
			}
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top.Type == TokenLeftParen || top.Type == TokenRightParen {
			return nil, &ParseError{Msg: "mismatched parentheses"}
		}
		output = append(output, RPNTerm{Type: TermOperator, Value: top.Value, IsUnary: top.IsUnary})
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func RPNToString(terms []RPNTerm) string {
	var parts []string
	for _, t := range terms {
		switch t.Type {
		case TermNumber:
			parts = append(parts, t.Value)
		case TermOperator:
			if t.IsUnary {
				parts = append(parts, "u"+t.Value)
			} else {
				parts = append(parts, t.Value)
			}
		case TermIdentifier:
			parts = append(parts, t.Value)
		case TermFunction:
			parts = append(parts, t.Value+"/"+string(rune('0'+t.ArgCount)))
		}
	}
	return strings.Join(parts, " ")
}
