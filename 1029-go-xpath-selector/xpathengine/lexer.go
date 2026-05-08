package xpathengine

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenSlash
	TokenDoubleSlash
	TokenDot
	TokenDoubleDot
	TokenAt
	TokenLBracket
	TokenRBracket
	TokenLParen
	TokenRParen
	TokenColon
	TokenAsterisk
	TokenPlus
	TokenMinus
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenAnd
	TokenOr
	TokenString
	TokenNumber
	TokenName
	TokenFunction
	TokenComma
	TokenPipe
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
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
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
		}
		l.readChar()
	}
}

func (l *Lexer) isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func (l *Lexer) isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func (l *Lexer) readIdentifier() string {
	startPos := l.pos
	for l.isLetter(l.ch) || l.isDigit(l.ch) || l.ch == '-' || l.ch == '_' {
		l.readChar()
	}
	return l.input[startPos:l.pos]
}

func (l *Lexer) readNumber() string {
	startPos := l.pos
	for l.isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[startPos:l.pos]
}

func (l *Lexer) readString() (string, error) {
	quote := l.ch
	l.readChar()
	var builder strings.Builder
	for l.ch != quote && l.ch != 0 {
		builder.WriteByte(l.ch)
		l.readChar()
	}
	if l.ch == 0 {
		return "", errors.New("unterminated string literal")
	}
	l.readChar()
	return builder.String(), nil
}

func (l *Lexer) NextToken() (*Token, error) {
	var tok *Token
	l.skipWhitespace()
	
	switch l.ch {
	case '/':
		if l.peekChar() == '/' {
			tok = &Token{Type: TokenDoubleSlash, Literal: "//", Line: l.line}
			l.readChar()
		} else {
			tok = &Token{Type: TokenSlash, Literal: "/", Line: l.line}
		}
	case '.':
		if l.peekChar() == '.' {
			tok = &Token{Type: TokenDoubleDot, Literal: "..", Line: l.line}
			l.readChar()
		} else {
			tok = &Token{Type: TokenDot, Literal: ".", Line: l.line}
		}
	case '@':
		tok = &Token{Type: TokenAt, Literal: "@", Line: l.line}
	case '[':
		tok = &Token{Type: TokenLBracket, Literal: "[", Line: l.line}
	case ']':
		tok = &Token{Type: TokenRBracket, Literal: "]", Line: l.line}
	case '(':
		tok = &Token{Type: TokenLParen, Literal: "(", Line: l.line}
	case ')':
		tok = &Token{Type: TokenRParen, Literal: ")", Line: l.line}
	case ':':
		tok = &Token{Type: TokenColon, Literal: ":", Line: l.line}
	case '*':
		tok = &Token{Type: TokenAsterisk, Literal: "*", Line: l.line}
	case '+':
		tok = &Token{Type: TokenPlus, Literal: "+", Line: l.line}
	case '-':
		tok = &Token{Type: TokenMinus, Literal: "-", Line: l.line}
	case '=':
		tok = &Token{Type: TokenEqual, Literal: "=", Line: l.line}
	case '!':
		if l.peekChar() == '=' {
			tok = &Token{Type: TokenNotEqual, Literal: "!=", Line: l.line}
			l.readChar()
		} else {
			return nil, errors.New("unexpected character: !")
		}
	case '<':
		if l.peekChar() == '=' {
			tok = &Token{Type: TokenLessEqual, Literal: "<=", Line: l.line}
			l.readChar()
		} else {
			tok = &Token{Type: TokenLess, Literal: "<", Line: l.line}
		}
	case '>':
		if l.peekChar() == '=' {
			tok = &Token{Type: TokenGreaterEqual, Literal: ">=", Line: l.line}
			l.readChar()
		} else {
			tok = &Token{Type: TokenGreater, Literal: ">", Line: l.line}
		}
	case ',':
		tok = &Token{Type: TokenComma, Literal: ",", Line: l.line}
	case '|':
		tok = &Token{Type: TokenPipe, Literal: "|", Line: l.line}
	case '\'', '"':
		str, err := l.readString()
		if err != nil {
			return nil, err
		}
		tok = &Token{Type: TokenString, Literal: str, Line: l.line}
	case 0:
		tok = &Token{Type: TokenEOF, Literal: "", Line: l.line}
	default:
		if l.isLetter(l.ch) {
			ident := l.readIdentifier()
			if l.ch == '(' {
				return &Token{Type: TokenFunction, Literal: ident, Line: l.line}, nil
			}
			if ident == "and" {
				return &Token{Type: TokenAnd, Literal: "and", Line: l.line}, nil
			}
			if ident == "or" {
				return &Token{Type: TokenOr, Literal: "or", Line: l.line}, nil
			}
			return &Token{Type: TokenName, Literal: ident, Line: l.line}, nil
		} else if l.isDigit(l.ch) {
			numStr := l.readNumber()
			if _, err := strconv.ParseFloat(numStr, 64); err != nil {
				return nil, errors.New("invalid number: " + numStr)
			}
			return &Token{Type: TokenNumber, Literal: numStr, Line: l.line}, nil
		} else {
			return nil, errors.New("unexpected character: " + string(l.ch))
		}
	}
	
	l.readChar()
	return tok, nil
}
