package ldapfilter

import (
	"fmt"
	"strings"
)

type TokenType int

const (
	TokenLParen TokenType = iota
	TokenRParen
	TokenAnd
	TokenOr
	TokenNot
	TokenAttr
	TokenOpEqual
	TokenOpApprox
	TokenOpGreaterEqual
	TokenOpLessEqual
	TokenValue
	TokenEOF
	TokenStar
)

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

type Lexer struct {
	input string
	pos   int
	width int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) next() byte {
	if l.pos >= len(l.input) {
		l.width = 0
		return 0
	}
	b := l.input[l.pos]
	l.width = 1
	l.pos++
	return b
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func isAttrChar(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= '0' && b <= '9':
		return true
	case b == '-', b == '_', b == '.', b == ';':
		return true
	default:
		return false
	}
}

func isHexChar(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'a' && b <= 'f') ||
		(b >= 'A' && b <= 'F')
}

func hexToByte(b byte) byte {
	if b >= '0' && b <= '9' {
		return b - '0'
	}
	if b >= 'a' && b <= 'f' {
		return b - 'a' + 10
	}
	return b - 'A' + 10
}

func isPlainValueChar(b byte) bool {
	if b == 0 {
		return false
	}
	if b == '*' || b == '(' || b == ')' || b == '\\' {
		return false
	}
	return true
}

func (l *Lexer) NextToken() (Token, error) {
	for {
		b := l.next()
		switch b {
		case 0:
			return Token{Type: TokenEOF, Pos: l.pos}, nil
		case ' ', '\t', '\n', '\r':
			continue
		case '(':
			return Token{Type: TokenLParen, Literal: "(", Pos: l.pos - 1}, nil
		case ')':
			return Token{Type: TokenRParen, Literal: ")", Pos: l.pos - 1}, nil
		case '&':
			return Token{Type: TokenAnd, Literal: "&", Pos: l.pos - 1}, nil
		case '|':
			return Token{Type: TokenOr, Literal: "|", Pos: l.pos - 1}, nil
		case '!':
			return Token{Type: TokenNot, Literal: "!", Pos: l.pos - 1}, nil
		case '*':
			return Token{Type: TokenStar, Literal: "*", Pos: l.pos - 1}, nil
		case '~':
			if l.peek() == '=' {
				l.next()
				return Token{Type: TokenOpApprox, Literal: "~=", Pos: l.pos - 2}, nil
			}
			return Token{}, fmt.Errorf("unexpected character '~' at position %d", l.pos-1)
		case '>':
			if l.peek() == '=' {
				l.next()
				return Token{Type: TokenOpGreaterEqual, Literal: ">=", Pos: l.pos - 2}, nil
			}
			return Token{}, fmt.Errorf("unexpected character '>' at position %d", l.pos-1)
		case '<':
			if l.peek() == '=' {
				l.next()
				return Token{Type: TokenOpLessEqual, Literal: "<=", Pos: l.pos - 2}, nil
			}
			return Token{}, fmt.Errorf("unexpected character '<' at position %d", l.pos-1)
		case '=':
			return Token{Type: TokenOpEqual, Literal: "=", Pos: l.pos - 1}, nil
		case '\\':
			return Token{}, fmt.Errorf("unexpected escape character at position %d, not in value context", l.pos-1)
		default:
			if isAttrChar(b) {
				l.backup()
				return l.readAttr()
			}
			return Token{}, fmt.Errorf("unexpected character %q at position %d", b, l.pos-1)
		}
	}
}

func (l *Lexer) readAttr() (Token, error) {
	start := l.pos
	for isAttrChar(l.peek()) {
		l.next()
	}
	if start == l.pos {
		return Token{}, fmt.Errorf("expected attribute at position %d", start)
	}
	return Token{Type: TokenAttr, Literal: l.input[start:l.pos], Pos: start}, nil
}

func (l *Lexer) ReadValue() (string, error) {
	var sb strings.Builder
	for {
		b := l.peek()
		switch b {
		case 0:
			return sb.String(), nil
		case ')':
			return sb.String(), nil
		case '*':
			return sb.String(), nil
		case '\\':
			l.next()
			h1 := l.next()
			if h1 == 0 {
				return "", fmt.Errorf("incomplete escape sequence at position %d", l.pos-2)
			}
			if !isHexChar(h1) {
				return "", fmt.Errorf("invalid hex character %q in escape sequence at position %d", h1, l.pos-2)
			}
			h2 := l.next()
			if h2 == 0 {
				return "", fmt.Errorf("incomplete escape sequence at position %d", l.pos-3)
			}
			if !isHexChar(h2) {
				return "", fmt.Errorf("invalid hex character %q in escape sequence at position %d", h2, l.pos-2)
			}
			val := hexToByte(h1)<<4 | hexToByte(h2)
			sb.WriteByte(val)
		default:
			if isPlainValueChar(b) {
				l.next()
				sb.WriteByte(b)
			} else {
				return "", fmt.Errorf("unexpected character %q in value at position %d", b, l.pos)
			}
		}
	}
}

func (l *Lexer) Pos() int {
	return l.pos
}
