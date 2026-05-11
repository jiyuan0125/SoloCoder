package tokenizer

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
	column  int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 1}
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
		l.column = 1
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

	l.skipWhitespace()

	line, col := l.line, l.column

	switch l.ch {
	case '=':
		tok = Token{Type: TokenEq, Literal: "=", Line: line, Column: col}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenNeq, Literal: string(ch) + string(l.ch), Line: line, Column: col}
		} else {
			tok = Token{Type: TokenType("ILLEGAL"), Literal: string(l.ch), Line: line, Column: col}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenGte, Literal: string(ch) + string(l.ch), Line: line, Column: col}
		} else {
			tok = Token{Type: TokenGt, Literal: string(l.ch), Line: line, Column: col}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenLte, Literal: string(ch) + string(l.ch), Line: line, Column: col}
		} else {
			tok = Token{Type: TokenLt, Literal: string(l.ch), Line: line, Column: col}
		}
	case '(':
		tok = Token{Type: TokenLParen, Literal: "(", Line: line, Column: col}
	case ')':
		tok = Token{Type: TokenRParen, Literal: ")", Line: line, Column: col}
	case ',':
		tok = Token{Type: TokenComma, Literal: ",", Line: line, Column: col}
	case '\'':
		lit, err := l.readString()
		if err != nil {
			return Token{Type: TokenType("ILLEGAL"), Literal: err.Error(), Line: line, Column: col}
		}
		tok = Token{Type: TokenString, Literal: lit, Line: line, Column: col}
	case '`':
		lit, err := l.readQuotedIdent()
		if err != nil {
			return Token{Type: TokenType("ILLEGAL"), Literal: err.Error(), Line: line, Column: col}
		}
		tok = Token{Type: TokenQuotedIdent, Literal: lit, Line: line, Column: col}
	case 0:
		tok = Token{Type: TokenEOF, Literal: "", Line: line, Column: col}
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tt := LookupKeyword(strings.ToUpper(ident))
			if tt != TokenIdentifier {
				tok = Token{Type: tt, Literal: ident, Line: line, Column: col}
			} else {
				tok = Token{Type: TokenIdentifier, Literal: ident, Line: line, Column: col}
			}
			return tok
		} else if isDigit(l.ch) {
			num := l.readNumber()
			tok = Token{Type: TokenNumber, Literal: num, Line: line, Column: col}
			return tok
		} else {
			tok = Token{Type: TokenType("ILLEGAL"), Literal: string(l.ch), Line: line, Column: col}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readString() (string, error) {
	var buf strings.Builder
	l.readChar()
	for {
		if l.ch == 0 {
			return "", fmt.Errorf("unterminated string")
		}
		if l.ch == '\'' {
			if l.peekChar() == '\'' {
				l.readChar()
				buf.WriteByte('\'')
			} else {
				break
			}
		} else if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			default:
				buf.WriteByte(l.ch)
			}
		} else {
			buf.WriteByte(l.ch)
		}
		l.readChar()
	}
	return buf.String(), nil
}

func (l *Lexer) readQuotedIdent() (string, error) {
	l.readChar()
	pos := l.pos
	for l.ch != '`' {
		if l.ch == 0 {
			return "", fmt.Errorf("unterminated quoted identifier")
		}
		l.readChar()
	}
	return l.input[pos:l.pos], nil
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}
