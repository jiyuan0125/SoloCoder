package buildtag

import (
	"fmt"
	"strings"
	"unicode"
)

type OldParser struct {
	input string
	pos   int
}

func NewOldParser(input string) *OldParser {
	return &OldParser{input: input, pos: 0}
}

func (p *OldParser) skipWhitespace() {
	for p.pos < len(p.input) && (unicode.IsSpace(rune(p.input[p.pos])) && p.input[p.pos] != ',') {
		p.pos++
	}
}

func (p *OldParser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *OldParser) consume() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	b := p.input[p.pos]
	p.pos++
	return b
}

func (p *OldParser) parseTag() Expr {
	start := p.pos
	for p.pos < len(p.input) && isTagChar(p.input[p.pos]) {
		p.pos++
	}
	return &TagExpr{Tag: strings.ToLower(p.input[start:p.pos])}
}

func (p *OldParser) parsePrimary() (Expr, error) {
	p.skipWhitespace()
	if p.peek() == '!' {
		p.consume()
		inner, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Inner: inner}, nil
	}
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of input")
	}
	return p.parseTag(), nil
}

func (p *OldParser) parseOr() (Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.peek() == ',' {
			p.consume()
			right, err := p.parsePrimary()
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

func (p *OldParser) parseAnd() (Expr, error) {
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.peek() == 0 || p.peek() == ',' {
			break
		}
		right, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		expr = &AndExpr{Left: expr, Right: right}
	}
	return expr, nil
}

func (p *OldParser) Parse() (Expr, error) {
	if strings.TrimSpace(p.input) == "" {
		return nil, nil
	}
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return nil, fmt.Errorf("unexpected character at position %d: '%c'", p.pos, p.input[p.pos])
	}
	return expr, nil
}

func ParseOldLines(lines []string) (Expr, error) {
	var result Expr
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p := NewOldParser(line)
		expr, err := p.Parse()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}
		if result == nil {
			result = expr
		} else {
			result = &AndExpr{Left: result, Right: expr}
		}
	}
	return result, nil
}
