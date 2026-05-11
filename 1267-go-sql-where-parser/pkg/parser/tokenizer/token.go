package tokenizer

type TokenType string

const (
	TokenEOF TokenType = "EOF"

	TokenIdentifier TokenType = "IDENTIFIER"
	TokenQuotedIdent TokenType = "QUOTED_IDENT"
	TokenString     TokenType = "STRING"
	TokenNumber     TokenType = "NUMBER"
	TokenKeyword    TokenType = "KEYWORD"

	TokenAnd    TokenType = "AND"
	TokenOr     TokenType = "OR"
	TokenNot    TokenType = "NOT"
	TokenIn     TokenType = "IN"
	TokenLike   TokenType = "LIKE"
	TokenBetween TokenType = "BETWEEN"
	TokenIs     TokenType = "IS"
	TokenNull   TokenType = "NULL"
	TokenEscape TokenType = "ESCAPE"

	TokenEq  TokenType = "="
	TokenNeq TokenType = "!="
	TokenGt  TokenType = ">"
	TokenLt  TokenType = "<"
	TokenGte TokenType = ">="
	TokenLte TokenType = "<="

	TokenLParen TokenType = "("
	TokenRParen TokenType = ")"
	TokenComma  TokenType = ","
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"AND":    TokenAnd,
	"OR":     TokenOr,
	"NOT":    TokenNot,
	"IN":     TokenIn,
	"LIKE":   TokenLike,
	"BETWEEN": TokenBetween,
	"IS":     TokenIs,
	"NULL":   TokenNull,
	"ESCAPE": TokenEscape,
}

func LookupKeyword(ident string) TokenType {
	if tt, ok := keywords[ident]; ok {
		return tt
	}
	return TokenIdentifier
}
