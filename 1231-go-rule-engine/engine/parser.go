package engine

import (
	"fmt"
	"strings"
	"unicode"
)

type token struct {
	typ   tokenType
	value string
}

type tokenType int

const (
	tokInvalid tokenType = iota
	tokEOF
	tokLiteral
	tokCompGT
	tokCompGTE
	tokCompLT
	tokCompLTE
	tokCompEQ
	tokCompNEQ
	tokLogicAND
	tokLogicOR
	tokLParen
	tokRParen
)

type lexer struct {
	input string
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: input, pos: 0}
}

func (l *lexer) skipSpace() {
	for l.pos < len(l.input) && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

func (l *lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) advance() {
	l.pos++
}

func (l *lexer) collectIdentifier() string {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '>' || ch == '<' || ch == '=' || ch == '!' ||
			ch == '(' || ch == ')' || unicode.IsSpace(rune(ch)) {
			break
		}
		l.pos++
	}
	return l.input[start:l.pos]
}

func (l *lexer) Next() token {
	l.skipSpace()
	if l.pos >= len(l.input) {
		return token{typ: tokEOF}
	}

	ch := l.peek()

	switch ch {
	case '(':
		l.advance()
		return token{typ: tokLParen}
	case ')':
		l.advance()
		return token{typ: tokRParen}
	case '>':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return token{typ: tokCompGTE, value: ">="}
		}
		return token{typ: tokCompGT, value: ">"}
	case '<':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return token{typ: tokCompLTE, value: "<="}
		}
		return token{typ: tokCompLT, value: "<"}
	case '=':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return token{typ: tokCompEQ, value: "=="}
		}
		return token{typ: tokInvalid, value: string(ch)}
	case '!':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return token{typ: tokCompNEQ, value: "!="}
		}
		return token{typ: tokInvalid, value: string(ch)}
	}

	ident := l.collectIdentifier()
	if ident == "" {
		return token{typ: tokInvalid}
	}

	upper := strings.ToUpper(ident)
	if upper == "AND" {
		return token{typ: tokLogicAND, value: "AND"}
	}
	if upper == "OR" {
		return token{typ: tokLogicOR, value: "OR"}
	}
	return token{typ: tokLiteral, value: ident}
}

type parser struct {
	lexer   *lexer
	current token
}

func ParseExpression(expr string) (*Node, error) {
	p := &parser{lexer: newLexer(expr)}
	p.current = p.lexer.Next()
	n, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.current.typ != tokEOF {
		return nil, fmt.Errorf("unexpected token at position %d", p.lexer.pos)
	}
	return n, nil
}

func (p *parser) advance() {
	p.current = p.lexer.Next()
}

func (p *parser) parseOr() (*Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.current.typ == tokLogicOR {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &Node{
			Type:    NodeTypeLogic,
			LogicOp: LogicOpOR,
			Children: []*Node{left, right},
		}
	}
	return left, nil
}

func (p *parser) parseAnd() (*Node, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.current.typ == tokLogicAND {
		p.advance()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &Node{
			Type:    NodeTypeLogic,
			LogicOp: LogicOpAND,
			Children: []*Node{left, right},
		}
	}
	return left, nil
}

func (p *parser) parseFactor() (*Node, error) {
	if p.current.typ == tokLParen {
		p.advance()
		n, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.current.typ != tokRParen {
			return nil, fmt.Errorf("expected )")
		}
		p.advance()
		return n, nil
	}
	return p.parseComparison()
}

func (p *parser) parseComparison() (*Node, error) {
	if p.current.typ != tokLiteral {
		return nil, fmt.Errorf("expected variable")
	}
	leftVar := p.current.value
	p.advance()

	var op CompOp
	switch p.current.typ {
	case tokCompGT:
		op = CompOpGT
	case tokCompGTE:
		op = CompOpGTE
	case tokCompLT:
		op = CompOpLT
	case tokCompLTE:
		op = CompOpLTE
	case tokCompEQ:
		op = CompOpEQ
	case tokCompNEQ:
		op = CompOpNEQ
	default:
		return nil, fmt.Errorf("expected comparison operator")
	}
	p.advance()

	if p.current.typ != tokLiteral {
		return nil, fmt.Errorf("expected value")
	}
	rightVal := p.current.value
	p.advance()

	return &Node{
		Type:       NodeTypeComparison,
		CompOp:     op,
		LeftVar:    leftVar,
		RightValue: rightVal,
	}, nil
}
