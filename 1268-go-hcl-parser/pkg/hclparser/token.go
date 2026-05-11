package hclparser

import (
	"fmt"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenString
	TokenHeredoc
	TokenNumber
	TokenBoolean
	TokenNull
	TokenEquals
	TokenComma
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenDot
	TokenQuestion
	TokenColon
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenAmpersand
	TokenPipe
	TokenExclamation
	TokenDoubleAmp
	TokenDoublePipe
	TokenEqualEqual
	TokenNotEqual
	TokenLess
	TokenGreater
	TokenLessEqual
	TokenGreaterEqual
	TokenLeftArrow
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenString:
		return "STRING"
	case TokenHeredoc:
		return "HEREDOC"
	case TokenNumber:
		return "NUMBER"
	case TokenBoolean:
		return "BOOLEAN"
	case TokenNull:
		return "NULL"
	case TokenEquals:
		return "="
	case TokenComma:
		return ","
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenDot:
		return "."
	case TokenQuestion:
		return "?"
	case TokenColon:
		return ":"
	case TokenPlus:
		return "+"
	case TokenMinus:
		return "-"
	case TokenStar:
		return "*"
	case TokenSlash:
		return "/"
	case TokenPercent:
		return "%"
	case TokenAmpersand:
		return "&"
	case TokenPipe:
		return "|"
	case TokenExclamation:
		return "!"
	case TokenDoubleAmp:
		return "&&"
	case TokenDoublePipe:
		return "||"
	case TokenEqualEqual:
		return "=="
	case TokenNotEqual:
		return "!="
	case TokenLess:
		return "<"
	case TokenGreater:
		return ">"
	case TokenLessEqual:
		return "<="
	case TokenGreaterEqual:
		return ">="
	case TokenLeftArrow:
		return "<<-"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

type Token struct {
	Type     TokenType
	Literal  string
	Line     int
	Column   int
	RawValue interface{}
}

func NewToken(typ TokenType, literal string, line, column int, rawValue interface{}) Token {
	return Token{
		Type:     typ,
		Literal:  literal,
		Line:     line,
		Column:   column,
		RawValue: rawValue,
	}
}
