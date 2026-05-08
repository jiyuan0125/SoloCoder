package rpn

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenNumber TokenType = iota
	TokenOperator
	TokenLeftParen
	TokenRightParen
	TokenIdentifier
	TokenComma
)

type Token struct {
	Type    TokenType
	Value   string
	IsUnary bool
}

func Tokenize(expr string) ([]Token, error) {
	var tokens []Token
	runes := []rune(expr)
	n := len(runes)
	i := 0

	expectUnary := true

	for i < n {
		r := runes[i]

		if unicode.IsSpace(r) {
			i++
			continue
		}

		if unicode.IsDigit(r) || r == '.' {
			start := i
			hasDot := false
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				if runes[i] == '.' {
					if hasDot {
						return nil, &ParseError{Msg: "invalid number format"}
					}
					hasDot = true
				}
				i++
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: string(runes[start:i])})
			expectUnary = false
			continue
		}

		if unicode.IsLetter(r) || r == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenIdentifier, Value: string(runes[start:i])})
			expectUnary = false
			continue
		}

		switch r {
		case '+', '-', '*', '/':
			isUnary := false
			if r == '-' && expectUnary {
				isUnary = true
			}
			tokens = append(tokens, Token{Type: TokenOperator, Value: string(r), IsUnary: isUnary})
			expectUnary = true
			i++
			continue
		case '(':
			tokens = append(tokens, Token{Type: TokenLeftParen, Value: "("})
			expectUnary = true
			i++
			continue
		case ')':
			tokens = append(tokens, Token{Type: TokenRightParen, Value: ")"})
			expectUnary = false
			i++
			continue
		case ',':
			tokens = append(tokens, Token{Type: TokenComma, Value: ","})
			expectUnary = true
			i++
			continue
		default:
			return nil, &ParseError{Msg: "unexpected character: " + string(r)}
		}
	}

	return tokens, nil
}

func (t Token) String() string {
	switch t.Type {
	case TokenNumber:
		return "NUMBER(" + t.Value + ")"
	case TokenOperator:
		if t.IsUnary {
			return "UNARY(" + t.Value + ")"
		}
		return "OP(" + t.Value + ")"
	case TokenLeftParen:
		return "LPAREN"
	case TokenRightParen:
		return "RPAREN"
	case TokenIdentifier:
		return "IDENT(" + t.Value + ")"
	case TokenComma:
		return "COMMA"
	}
	return "UNKNOWN(" + t.Value + ")"
}

func TokensToString(tokens []Token) string {
	var parts []string
	for _, t := range tokens {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, " ")
}
