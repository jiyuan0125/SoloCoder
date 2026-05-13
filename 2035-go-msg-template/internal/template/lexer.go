package template

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenText TokenType = iota
	TokenOpen
	TokenClose
	TokenVar
	TokenDefaultValue
	TokenIf
	TokenEndIf
	TokenPartial
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
	column  int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	for l.ch != 0 {
		if l.ch == '{' && l.peekChar() == '{' {
			tok.Line, tok.Column = l.line, l.column
			l.readChar()
			l.readChar()
			l.skipWhitespace()

			switch {
			case l.ch == '#':
				return l.readTag()
			case l.ch == '/':
				return l.readTag()
			case l.ch == '>':
				return l.readPartial()
			default:
				return l.readVariable()
			}
		}

		return l.readText()
	}

	tok.Type = TokenClose
	tok.Literal = ""
	tok.Line = l.line
	tok.Column = l.column
	return tok
}

func (l *Lexer) readText() Token {
	line, col := l.line, l.column
	var sb strings.Builder

	for l.ch != 0 && !(l.ch == '{' && l.peekChar() == '{') {
		sb.WriteByte(l.ch)
		l.readChar()
	}

	return Token{Type: TokenText, Literal: sb.String(), Line: line, Column: col}
}

func (l *Lexer) readVariable() Token {
	line, col := l.line, l.column
	var sb strings.Builder

	for l.ch != 0 && l.ch != '}' && l.ch != '|' {
		sb.WriteByte(l.ch)
		l.readChar()
	}

	varName := strings.TrimSpace(sb.String())
	var defaultValue string

	if l.ch == '|' {
		l.readChar()
		l.skipWhitespace()
		sb.Reset()
		for l.ch != 0 && l.ch != '}' {
			sb.WriteByte(l.ch)
			l.readChar()
		}
		defaultValue = strings.TrimSpace(sb.String())
	}

	l.skipWhitespace()
	if l.ch != '}' || l.peekChar() != '}' {
		return Token{Type: TokenClose, Literal: "", Line: line, Column: col}
	}
	l.readChar()
	l.readChar()

	tok := Token{Type: TokenVar, Literal: varName, Line: line, Column: col}
	if defaultValue != "" {
		tok.Literal = varName + "|" + defaultValue
	}
	return tok
}

func (l *Lexer) readTag() Token {
	line, col := l.line, l.column
	startCh := l.ch
	l.readChar()
	l.skipWhitespace()

	var sb strings.Builder
	for l.ch != 0 && l.ch != '}' {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	tagContent := strings.TrimSpace(sb.String())

	l.skipWhitespace()
	if l.ch != '}' || l.peekChar() != '}' {
		return Token{Type: TokenClose, Literal: "", Line: line, Column: col}
	}
	l.readChar()
	l.readChar()

	if startCh == '#' && strings.HasPrefix(tagContent, "if ") {
		condition := strings.TrimSpace(strings.TrimPrefix(tagContent, "if"))
		return Token{Type: TokenIf, Literal: condition, Line: line, Column: col}
	}
	if startCh == '/' && tagContent == "if" {
		return Token{Type: TokenEndIf, Literal: "", Line: line, Column: col}
	}

	return Token{Type: TokenClose, Literal: "", Line: line, Column: col}
}

func (l *Lexer) readPartial() Token {
	line, col := l.line, l.column
	l.readChar()
	l.skipWhitespace()

	var sb strings.Builder
	for l.ch != 0 && l.ch != '}' {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	partialName := strings.TrimSpace(sb.String())

	l.skipWhitespace()
	if l.ch != '}' || l.peekChar() != '}' {
		return Token{Type: TokenClose, Literal: "", Line: line, Column: col}
	}
	l.readChar()
	l.readChar()

	return Token{Type: TokenPartial, Literal: partialName, Line: line, Column: col}
}

func (l *Lexer) skipWhitespace() {
	for l.ch != 0 && unicode.IsSpace(rune(l.ch)) {
		l.readChar()
	}
}

type SyntaxError struct {
	File   string
	Line   int
	Column int
	Msg    string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Msg)
}
