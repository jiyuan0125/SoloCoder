package jsonpath

import (
	"strings"
	"unicode"
)

type Lexer struct {
	input string
	pos   int
	ch    byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, pos: -1}
	l.next()
	return l
}

func (l *Lexer) next() {
	if l.pos+1 >= len(l.input) {
		l.ch = 0
		l.pos = len(l.input)
		return
	}
	l.pos++
	l.ch = l.input[l.pos]
}

func (l *Lexer) peek() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) skipWhitespace() {
	for l.ch != 0 && unicode.IsSpace(rune(l.ch)) {
		l.next()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for l.ch != 0 && (unicode.IsLetter(rune(l.ch)) || unicode.IsDigit(rune(l.ch)) || l.ch == '_') {
		l.next()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readNumber() string {
	start := l.pos
	for l.ch != 0 && (unicode.IsDigit(rune(l.ch)) || l.ch == '.') {
		l.next()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readString() (string, error) {
	quote := l.ch
	l.next()
	var sb strings.Builder
	for l.ch != 0 && l.ch != quote {
		if l.ch == '\\' {
			l.next()
			switch l.ch {
			case '\'':
				sb.WriteByte('\'')
			case '\\':
				sb.WriteByte('\\')
			case '/':
				sb.WriteByte('/')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			default:
				sb.WriteByte(l.ch)
			}
		} else {
			sb.WriteByte(l.ch)
		}
		l.next()
	}
	if l.ch != quote {
		return "", ErrUnterminatedString
	}
	l.next()
	return sb.String(), nil
}

func (l *Lexer) NextToken() (Token, error) {
	l.skipWhitespace()
	switch l.ch {
	case 0:
		return Token{Type: TokenEOF}, nil
	case '$':
		l.next()
		return Token{Type: TokenRoot}, nil
	case '@':
		l.next()
		return Token{Type: TokenAt}, nil
	case '*':
		l.next()
		return Token{Type: TokenWildcard}, nil
	case '.':
		if l.peek() == '.' {
			l.next()
			l.next()
			return Token{Type: TokenDotDot}, nil
		}
		l.next()
		return Token{Type: TokenDot}, nil
	case '[':
		l.next()
		return Token{Type: TokenBracketOpen}, nil
	case ']':
		l.next()
		return Token{Type: TokenBracketClose}, nil
	case '(':
		l.next()
		return Token{Type: TokenParenOpen}, nil
	case ')':
		l.next()
		return Token{Type: TokenParenClose}, nil
	case '?':
		l.next()
		return Token{Type: TokenQuestion}, nil
	case ':':
		l.next()
		return Token{Type: TokenColon}, nil
	case ',':
		l.next()
		return Token{Type: TokenComma}, nil
	case '=':
		if l.peek() == '=' {
			l.next()
			l.next()
			return Token{Type: TokenEqual}, nil
		}
		return Token{}, ErrInvalidToken
	case '!':
		if l.peek() == '=' {
			l.next()
			l.next()
			return Token{Type: TokenNotEqual}, nil
		}
		return Token{}, ErrInvalidToken
	case '<':
		if l.peek() == '=' {
			l.next()
			l.next()
			return Token{Type: TokenLessEqual}, nil
		}
		l.next()
		return Token{Type: TokenLess}, nil
	case '>':
		if l.peek() == '=' {
			l.next()
			l.next()
			return Token{Type: TokenGreaterEqual}, nil
		}
		l.next()
		return Token{Type: TokenGreater}, nil
	case '&':
		if l.peek() == '&' {
			l.next()
			l.next()
			return Token{Type: TokenAnd}, nil
		}
		return Token{}, ErrInvalidToken
	case '|':
		if l.peek() == '|' {
			l.next()
			l.next()
			return Token{Type: TokenOr}, nil
		}
		return Token{}, ErrInvalidToken
	case '\'':
		s, err := l.readString()
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenString, Literal: s}, nil
	default:
		if unicode.IsLetter(rune(l.ch)) || l.ch == '_' {
			id := l.readIdentifier()
			if id == "true" || id == "false" {
				return Token{Type: TokenBool, Literal: id}, nil
			}
			return Token{Type: TokenIdentifier, Literal: id}, nil
		}
		if unicode.IsDigit(rune(l.ch)) || (l.ch == '-' && unicode.IsDigit(rune(l.peek()))) {
			var sign string
			if l.ch == '-' {
				sign = "-"
				l.next()
			}
			num := l.readNumber()
			return Token{Type: TokenNumber, Literal: sign + num}, nil
		}
		return Token{}, ErrInvalidToken
	}
}
