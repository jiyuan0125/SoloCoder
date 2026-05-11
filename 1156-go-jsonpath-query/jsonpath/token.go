package jsonpath

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenRoot
	TokenDot
	TokenDotDot
	TokenBracketOpen
	TokenBracketClose
	TokenWildcard
	TokenIdentifier
	TokenNumber
	TokenString
	TokenAt
	TokenQuestion
	TokenParenOpen
	TokenParenClose
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenAnd
	TokenOr
	TokenColon
	TokenComma
	TokenBool
)

type Token struct {
	Type    TokenType
	Literal string
}
