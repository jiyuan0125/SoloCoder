package eval

type TokenType string

const (
	TOKEN_EOF           TokenType = "EOF"
	TOKEN_NUMBER_INT    TokenType = "NUMBER_INT"
	TOKEN_NUMBER_FLOAT  TokenType = "NUMBER_FLOAT"
	TOKEN_IDENTIFIER    TokenType = "IDENTIFIER"
	TOKEN_PLUS          TokenType = "PLUS"
	TOKEN_MINUS         TokenType = "MINUS"
	TOKEN_MULTIPLY      TokenType = "MULTIPLY"
	TOKEN_DIVIDE        TokenType = "DIVIDE"
	TOKEN_LPAREN        TokenType = "LPAREN"
	TOKEN_RPAREN        TokenType = "RPAREN"
	TOKEN_COMMA         TokenType = "COMMA"
	TOKEN_EQ            TokenType = "EQ"
	TOKEN_NEQ           TokenType = "NEQ"
	TOKEN_GT            TokenType = "GT"
	TOKEN_GTE           TokenType = "GTE"
	TOKEN_LT            TokenType = "LT"
	TOKEN_LTE           TokenType = "LTE"
	TOKEN_AND           TokenType = "AND"
	TOKEN_OR            TokenType = "OR"
	TOKEN_TRUE          TokenType = "TRUE"
	TOKEN_FALSE         TokenType = "FALSE"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"and":   TOKEN_AND,
	"or":    TOKEN_OR,
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
}

func LookupIdentifier(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENTIFIER
}
