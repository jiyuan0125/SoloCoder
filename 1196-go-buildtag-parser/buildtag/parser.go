package buildtag

import (
	"fmt"
	"strings"
	"unicode"
)

type Expr interface {
	Evaluate(tags map[string]bool) bool
	String() string
}

type TagExpr struct {
	Tag string
}

func (e *TagExpr) Evaluate(tags map[string]bool) bool {
	return tags[strings.ToLower(e.Tag)]
}

func (e *TagExpr) String() string {
	return e.Tag
}

type NotExpr struct {
	Inner Expr
}

func (e *NotExpr) Evaluate(tags map[string]bool) bool {
	return !e.Inner.Evaluate(tags)
}

func (e *NotExpr) String() string {
	return fmt.Sprintf("!%s", e.Inner.String())
}

type AndExpr struct {
	Left, Right Expr
}

func (e *AndExpr) Evaluate(tags map[string]bool) bool {
	return e.Left.Evaluate(tags) && e.Right.Evaluate(tags)
}

func (e *AndExpr) String() string {
	return fmt.Sprintf("(%s && %s)", e.Left.String(), e.Right.String())
}

type OrExpr struct {
	Left, Right Expr
}

func (e *OrExpr) Evaluate(tags map[string]bool) bool {
	return e.Left.Evaluate(tags) || e.Right.Evaluate(tags)
}

func (e *OrExpr) String() string {
	return fmt.Sprintf("(%s || %s)", e.Left.String(), e.Right.String())
}

type Parser struct {
	input string
	pos   int
}

func NewParser(input string) *Parser {
	return &Parser{input: strings.TrimSpace(input), pos: 0}
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *Parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) consume() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	b := p.input[p.pos]
	p.pos++
	return b
}

func (p *Parser) match(s string) bool {
	if p.pos+len(s) > len(p.input) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if p.input[p.pos+i] != s[i] {
			return false
		}
	}
	p.pos += len(s)
	return true
}

func isTagChar(b byte) bool {
	return b == '_' || b == '.' || unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b))
}

func (p *Parser) parseTag() Expr {
	start := p.pos
	for p.pos < len(p.input) && isTagChar(p.input[p.pos]) {
		p.pos++
	}
	return &TagExpr{Tag: strings.ToLower(p.input[start:p.pos])}
}

func (p *Parser) parsePrimary() (Expr, error) {
	p.skipWhitespace()
	if p.peek() == '(' {
		p.consume()
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.peek() != ')' {
			return nil, fmt.Errorf("expected ')'")
		}
		p.consume()
		return expr, nil
	}
	if p.peek() == '!' {
		p.consume()
		inner, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Inner: inner}, nil
	}
	return p.parseTag(), nil
}

func (p *Parser) parseAnd() (Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.match("&&") {
			right, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			expr = &AndExpr{Left: expr, Right: right}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) parseOr() (Expr, error) {
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.match("||") {
			right, err := p.parseAnd()
			if err != nil {
				return nil, err
			}
			expr = &OrExpr{Left: expr, Right: right}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) Parse() (Expr, error) {
	if strings.TrimSpace(p.input) == "" {
		return nil, nil
	}
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return nil, fmt.Errorf("unexpected character at position %d", p.pos)
	}
	return expr, nil
}
