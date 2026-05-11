package sexpr

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenLParen
	TokenRParen
	TokenQuote
	TokenInt
	TokenFloat
	TokenString
	TokenSymbol
)

type Token struct {
	Type   TokenType
	Value  string
	Int    int64
	Float  float64
	Pos    Position
}

func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenQuote:
		return "'"
	case TokenInt:
		return fmt.Sprintf("Int:%d", t.Int)
	case TokenFloat:
		return fmt.Sprintf("Float:%g", t.Float)
	case TokenString:
		return fmt.Sprintf("String:%q", t.Value)
	case TokenSymbol:
		return fmt.Sprintf("Symbol:%s", t.Value)
	default:
		return "Unknown"
	}
}

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) currentPos() Position {
	return Position{Line: l.line, Column: l.column}
}

func (l *Lexer) peek() (byte, bool) {
	if l.pos >= len(l.input) {
		return 0, false
	}
	return l.input[l.pos], true
}

func (l *Lexer) advance() {
	if l.pos >= len(l.input) {
		return
	}
	c := l.input[l.pos]
	l.pos++
	if c == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

func (l *Lexer) skipWhitespace() {
	for {
		c, ok := l.peek()
		if !ok {
			return
		}
		if unicode.IsSpace(rune(c)) {
			l.advance()
		} else {
			return
		}
	}
}

func (l *Lexer) skipComment() {
	for {
		c, ok := l.peek()
		if !ok || c == '\n' {
			return
		}
		l.advance()
	}
}

func (l *Lexer) NextToken() (Token, error) {
	for {
		l.skipWhitespace()

		c, ok := l.peek()
		if !ok {
			return Token{Type: TokenEOF, Pos: l.currentPos()}, nil
		}

		if c == ';' {
			l.skipComment()
			continue
		}

		pos := l.currentPos()

		switch {
		case c == '(':
			l.advance()
			return Token{Type: TokenLParen, Pos: pos}, nil
		case c == ')':
			l.advance()
			return Token{Type: TokenRParen, Pos: pos}, nil
		case c == '\'':
			l.advance()
			return Token{Type: TokenQuote, Pos: pos}, nil
		case c == '"':
			return l.readString(pos)
		case c == '-' || unicode.IsDigit(rune(c)) || c == '.':
			return l.readNumberOrSymbol(pos)
		case isSymbolStart(c):
			return l.readSymbol(pos)
		default:
			return Token{}, fmt.Errorf("%s: illegal character %q", pos, c)
		}
	}
}

func isSymbolStart(c byte) bool {
	return unicode.IsLetter(rune(c)) ||
		c == '-' || c == '!' || c == '?' || c == '*' ||
		c == '=' || c == '+' || c == '/' || c == '<'
}

func isSymbolChar(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) ||
		c == '-' || c == '!' || c == '?' || c == '*' ||
		c == '=' || c == '+' || c == '/' || c == '<'
}

func (l *Lexer) readString(startPos Position) (Token, error) {
	l.advance()
	var builder strings.Builder

	for {
		c, ok := l.peek()
		if !ok {
			return Token{}, fmt.Errorf("%s: unterminated string", startPos)
		}
		l.advance()

		switch c {
		case '"':
			return Token{
				Type:  TokenString,
				Value: builder.String(),
				Pos:   startPos,
			}, nil
		case '\\':
			esc, ok := l.peek()
			if !ok {
				return Token{}, fmt.Errorf("%s: unterminated string", startPos)
			}
			l.advance()
			switch esc {
			case 'n':
				builder.WriteByte('\n')
			case 't':
				builder.WriteByte('\t')
			case '\\':
				builder.WriteByte('\\')
			case '"':
				builder.WriteByte('"')
			case 'r':
				return Token{}, fmt.Errorf("%s: invalid escape sequence '\\r'", startPos)
			case '0':
				return Token{}, fmt.Errorf("%s: invalid escape sequence '\\0'", startPos)
			default:
				return Token{}, fmt.Errorf("%s: invalid escape sequence '\\%c'", startPos, esc)
			}
		case '\n':
			return Token{}, fmt.Errorf("%s: unterminated string", startPos)
		default:
			builder.WriteByte(c)
		}
	}
}

func (l *Lexer) readNumberOrSymbol(startPos Position) (Token, error) {
	start := l.pos
	c, _ := l.peek()

	hasDot := false
	hasDigit := false

	if c == '-' {
		l.advance()
		if next, ok := l.peek(); ok && !unicode.IsDigit(rune(next)) && next != '.' {
			return l.readSymbolFrom(start, startPos)
		}
	}

	for {
		c, ok := l.peek()
		if !ok {
			break
		}
		if unicode.IsDigit(rune(c)) {
			hasDigit = true
			l.advance()
		} else if c == '.' && !hasDot {
			hasDot = true
			l.advance()
			if next, ok := l.peek(); ok && !unicode.IsDigit(rune(next)) {
				return Token{}, fmt.Errorf("%s: invalid number: %s", startPos, l.input[start:l.pos])
			}
		} else if isSymbolChar(c) {
			return l.readSymbolFrom(start, startPos)
		} else {
			break
		}
	}

	numStr := l.input[start:l.pos]

	if numStr == "-" || (len(numStr) == 1 && numStr[0] == '.') {
		return l.readSymbolFrom(start, startPos)
	}

	if !hasDigit {
		return l.readSymbolFrom(start, startPos)
	}

	if hasDot {
		return Token{
			Type:  TokenFloat,
			Float: parseFloat(numStr),
			Pos:   startPos,
		}, nil
	}

	return Token{
		Type: TokenInt,
		Int:  parseInt(numStr),
		Pos:  startPos,
	}, nil
}

func (l *Lexer) readSymbol(startPos Position) (Token, error) {
	return l.readSymbolFrom(l.pos, startPos)
}

func (l *Lexer) readSymbolFrom(start int, startPos Position) (Token, error) {
	l.pos = start
	for {
		c, ok := l.peek()
		if !ok || !isSymbolChar(c) {
			break
		}
		l.advance()
		l.column++
	}
	return Token{
		Type:  TokenSymbol,
		Value: l.input[start:l.pos],
		Pos:   startPos,
	}, nil
}

func parseInt(s string) int64 {
	var n int64
	neg := false
	i := 0
	if len(s) > 0 && s[0] == '-' {
		neg = true
		i = 1
	}
	for ; i < len(s); i++ {
		n = n*10 + int64(s[i]-'0')
	}
	if neg {
		return -n
	}
	return n
}

func parseFloat(s string) float64 {
	var n float64
	var frac float64 = 0.1
	neg := false
	i := 0
	hasDot := false

	if len(s) > 0 && s[0] == '-' {
		neg = true
		i = 1
	}

	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			hasDot = true
			continue
		}
		if hasDot {
			n += float64(c-'0') * frac
			frac /= 10
		} else {
			n = n*10 + float64(c-'0')
		}
	}

	if neg {
		return -n
	}
	return n
}
