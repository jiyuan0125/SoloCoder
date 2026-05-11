package lexer

type TokenType string

const (
	TokenEOF           TokenType = "EOF"
	TokenIllegal       TokenType = "ILLEGAL"
	TokenName          TokenType = "NAME"
	TokenString        TokenType = "STRING"
	TokenInt           TokenType = "INT"
	TokenFloat         TokenType = "FLOAT"
	TokenBraceL        TokenType = "{"
	TokenBraceR        TokenType = "}"
	TokenParenL        TokenType = "("
	TokenParenR        TokenType = ")"
	TokenBracketL      TokenType = "["
	TokenBracketR      TokenType = "]"
	TokenColon         TokenType = ":"
	TokenComma         TokenType = ","
	TokenEquals        TokenType = "="
	TokenBang          TokenType = "!"
	TokenDollar        TokenType = "$"
	TokenDot           TokenType = "."
	TokenEllipsis      TokenType = "..."
	TokenQuery         TokenType = "query"
	TokenMutation      TokenType = "mutation"
	TokenSubscription  TokenType = "subscription"
	TokenFragment      TokenType = "fragment"
	TokenOn            TokenType = "on"
	TokenTrue          TokenType = "true"
	TokenFalse         TokenType = "false"
	TokenNull          TokenType = "null"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"query":        TokenQuery,
	"mutation":     TokenMutation,
	"subscription": TokenSubscription,
	"fragment":     TokenFragment,
	"on":           TokenOn,
	"true":         TokenTrue,
	"false":        TokenFalse,
	"null":         TokenNull,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenName
}
