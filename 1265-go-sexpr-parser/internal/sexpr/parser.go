package sexpr

import (
	"fmt"
)

type Parser struct {
	lexer *Lexer
	token Token
	err   error
}

func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	parser := &Parser{lexer: lexer}
	parser.nextToken()
	return parser
}

func (p *Parser) nextToken() {
	p.token, p.err = p.lexer.NextToken()
}

func (p *Parser) Parse() ([]Expr, error) {
	var exprs []Expr

	for p.token.Type != TokenEOF {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	return exprs, nil
}

func (p *Parser) ParseOne() (Expr, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	if p.err != nil {
		return nil, p.err
	}

	token := p.token

	switch token.Type {
	case TokenQuote:
		return p.parseQuote()
	case TokenLParen:
		return p.parseList()
	case TokenRParen:
		return nil, fmt.Errorf("%s: unexpected ')'", token.Pos)
	case TokenInt:
		expr := NewInt(token.Int, token.Pos)
		p.nextToken()
		return expr, nil
	case TokenFloat:
		expr := NewFloat(token.Float, token.Pos)
		p.nextToken()
		return expr, nil
	case TokenString:
		expr := NewString(token.Value, token.Pos)
		p.nextToken()
		return expr, nil
	case TokenSymbol:
		expr := NewSymbol(token.Value, token.Pos)
		p.nextToken()
		return expr, nil
	case TokenEOF:
		return nil, fmt.Errorf("%s: unexpected EOF", token.Pos)
	default:
		return nil, fmt.Errorf("%s: unexpected token %v", token.Pos, token)
	}
}

func (p *Parser) parseQuote() (Expr, error) {
	quotePos := p.token.Pos
	p.nextToken()

	if p.token.Type == TokenEOF {
		return nil, fmt.Errorf("%s: quote expects an expression", quotePos)
	}
	if p.token.Type == TokenRParen {
		return nil, fmt.Errorf("%s: quote expects an expression, got ')'", quotePos)
	}

	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return NewList(
		[]Expr{
			NewSymbol("quote", quotePos),
			expr,
		},
		quotePos,
	), nil
}

func (p *Parser) parseList() (Expr, error) {
	listPos := p.token.Pos
	p.nextToken()

	var elements []Expr

	for p.token.Type != TokenRParen {
		if p.token.Type == TokenEOF {
			return nil, fmt.Errorf("%s: unterminated list", listPos)
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}

	p.nextToken()
	return NewList(elements, listPos), nil
}

func Parse(input string) ([]Expr, error) {
	parser := NewParser(input)
	return parser.Parse()
}

func ParseOne(input string) (Expr, error) {
	parser := NewParser(input)
	return parser.ParseOne()
}
