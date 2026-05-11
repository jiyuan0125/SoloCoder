package regex

import (
	"fmt"
	"unicode/utf8"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenChar
	TokenPipe
	TokenStar
	TokenPlus
	TokenQuestion
	TokenLParen
	TokenRParen
)

type Token struct {
	Type  TokenType
	Value rune
	Pos   int
}

type Lexer struct {
	input string
	pos   int
	width int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) Lex() ([]Token, error) {
	var tokens []Token

	for {
		pos := l.pos
		r := l.next()

		if r == 0 {
			tokens = append(tokens, Token{Type: TokenEOF, Pos: pos})
			return tokens, nil
		}

		switch r {
		case '|':
			tokens = append(tokens, Token{Type: TokenPipe, Pos: pos})
		case '*':
			tokens = append(tokens, Token{Type: TokenStar, Pos: pos})
		case '+':
			tokens = append(tokens, Token{Type: TokenPlus, Pos: pos})
		case '?':
			tokens = append(tokens, Token{Type: TokenQuestion, Pos: pos})
		case '(':
			tokens = append(tokens, Token{Type: TokenLParen, Pos: pos})
		case ')':
			tokens = append(tokens, Token{Type: TokenRParen, Pos: pos})
		case '\\':
			n := l.next()
			if n == 0 {
				return nil, fmt.Errorf("pos %d: trailing backslash", pos)
			}
			tokens = append(tokens, Token{Type: TokenChar, Value: n, Pos: pos})
		default:
			if r < 32 || r == 127 {
				return nil, fmt.Errorf("pos %d: invalid control character", pos)
			}
			switch r {
			case '.', '[', ']', '{', '}', '^', '$':
				return nil, fmt.Errorf("pos %d: unsupported metacharacter '%c' (use '\\%c' to match literally)", pos, r, r)
			}
			tokens = append(tokens, Token{Type: TokenChar, Value: r, Pos: pos})
		}
	}
}
