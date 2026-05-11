package ldapfilter

import (
	"errors"
	"fmt"
)

type Parser struct {
	lexer *Lexer
	tok   Token
	err   error
}

func Parse(filter string) (Node, error) {
	trimmed := filter
	if len(trimmed) == 0 {
		return nil, errors.New("empty filter")
	}
	p := &Parser{lexer: NewLexer(trimmed)}
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	node, err := p.parseFilter()
	if err != nil {
		return nil, err
	}
	if p.tok.Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token %q after filter at position %d", p.tok.Literal, p.tok.Pos)
	}
	return node, nil
}

func (p *Parser) next() {
	tok, err := p.lexer.NextToken()
	if err != nil {
		p.err = err
		return
	}
	p.tok = tok
}

func (p *Parser) parseFilter() (Node, error) {
	if p.tok.Type != TokenLParen {
		return nil, fmt.Errorf("expected '(' at position %d, got %q", p.tok.Pos, p.tok.Literal)
	}
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	node, err := p.parseFilterComp()
	if err != nil {
		return nil, err
	}
	if p.tok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ')' at position %d, got %q", p.tok.Pos, p.tok.Literal)
	}
	p.next()
	return node, p.err
}

func (p *Parser) parseFilterComp() (Node, error) {
	switch p.tok.Type {
	case TokenAnd:
		return p.parseAnd()
	case TokenOr:
		return p.parseOr()
	case TokenNot:
		return p.parseNot()
	case TokenAttr:
		return p.parseItem()
	default:
		return nil, fmt.Errorf("expected filter component at position %d, got %q", p.tok.Pos, p.tok.Literal)
	}
}

func (p *Parser) parseAnd() (Node, error) {
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	children, err := p.parseFilterList()
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, errors.New("'&' requires at least one child filter")
	}
	return &AndNode{Children: children}, nil
}

func (p *Parser) parseOr() (Node, error) {
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	children, err := p.parseFilterList()
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, errors.New("'|' requires at least one child filter")
	}
	return &OrNode{Children: children}, nil
}

func (p *Parser) parseNot() (Node, error) {
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	child, err := p.parseFilter()
	if err != nil {
		return nil, err
	}
	return &NotNode{Child: child}, nil
}

func (p *Parser) parseFilterList() ([]Node, error) {
	var children []Node
	for {
		if p.tok.Type == TokenRParen {
			return children, nil
		}
		if p.tok.Type == TokenEOF {
			return nil, errors.New("unexpected EOF in filter list")
		}
		child, err := p.parseFilter()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
}

func (p *Parser) parseItem() (Node, error) {
	attr := p.tok.Literal
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	switch p.tok.Type {
	case TokenOpEqual:
		return p.parseEqual(attr)
	case TokenOpApprox:
		return p.parseSimple(attr, OpApprox)
	case TokenOpGreaterEqual:
		return p.parseSimple(attr, OpGreaterEqual)
	case TokenOpLessEqual:
		return p.parseSimple(attr, OpLessEqual)
	default:
		return nil, fmt.Errorf("expected operator after attribute %q at position %d", attr, p.tok.Pos)
	}
}

func (p *Parser) parseSimple(attr string, op Operator) (Node, error) {
	val, err := p.lexer.ReadValue()
	if err != nil {
		return nil, err
	}
	p.next()
	if p.err != nil {
		return nil, p.err
	}
	if p.tok.Type == TokenStar {
		return nil, fmt.Errorf("unexpected '*' in simple value for operator %q", op.String())
	}
	if p.tok.Type != TokenRParen && p.tok.Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token %q after value", p.tok.Literal)
	}
	return &SimpleNode{Attribute: attr, Operator: op, Value: val}, nil
}

func (p *Parser) parseEqual(attr string) (Node, error) {
	var hasStar bool
	var initial string
	var any []string
	var final string

	val, err := p.lexer.ReadValue()
	if err != nil {
		return nil, err
	}

	if p.lexer.peek() == '*' {
		initial = val
		hasStar = true
		p.next()
		for {
			seg, err := p.lexer.ReadValue()
			if err != nil {
				return nil, err
			}
			if p.lexer.peek() == '*' {
				if seg != "" || len(any) > 0 {
					any = append(any, seg)
				}
				p.next()
			} else {
				final = seg
				break
			}
		}
		p.next()
	} else {
		final = val
		p.next()
	}

	if p.tok.Type == TokenStar {
		return nil, errors.New("unexpected token after value")
	}

	if hasStar {
		if initial == "" && len(any) == 0 && final == "" {
			return &PresentNode{Attribute: attr}, nil
		}
		return &SubstringNode{
			Attribute: attr,
			Initial:   initial,
			Any:       any,
			Final:     final,
		}, nil
	}

	return &SimpleNode{
		Attribute: attr,
		Operator:  OpEqual,
		Value:     final,
	}, nil
}
