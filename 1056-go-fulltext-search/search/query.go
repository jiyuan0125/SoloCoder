package search

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokTerm TokenType = iota
	TokAnd
	TokOr
	TokNot
	TokLParen
	TokRParen
	TokEOF
)

type QueryToken struct {
	Type  TokenType
	Value string
}

type QueryLexer struct {
	input  []rune
	pos    int
	tokens []QueryToken
}

func NewQueryLexer(input string) *QueryLexer {
	return &QueryLexer{
		input: []rune(input),
		pos:   0,
	}
}

func (l *QueryLexer) Tokenize() ([]QueryToken, error) {
	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			break
		}

		r := l.input[l.pos]
		switch {
		case r == '(':
			l.tokens = append(l.tokens, QueryToken{Type: TokLParen, Value: "("})
			l.pos++
		case r == ')':
			l.tokens = append(l.tokens, QueryToken{Type: TokRParen, Value: ")"})
			l.pos++
		case isAlphaNum(r) || isChinese(r):
			term := l.readTerm()
			upper := strings.ToUpper(term)
			switch upper {
			case "AND":
				l.tokens = append(l.tokens, QueryToken{Type: TokAnd, Value: "AND"})
			case "OR":
				l.tokens = append(l.tokens, QueryToken{Type: TokOr, Value: "OR"})
			case "NOT":
				l.tokens = append(l.tokens, QueryToken{Type: TokNot, Value: "NOT"})
			default:
				l.tokens = append(l.tokens, QueryToken{Type: TokTerm, Value: term})
			}
		default:
			return nil, fmt.Errorf("unexpected character: %c at position %d", r, l.pos)
		}
	}
	l.tokens = append(l.tokens, QueryToken{Type: TokEOF, Value: ""})
	return l.tokens, nil
}

func (l *QueryLexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *QueryLexer) readTerm() string {
	start := l.pos
	for l.pos < len(l.input) {
		r := l.input[l.pos]
		if !isAlphaNum(r) && !isChinese(r) {
			break
		}
		l.pos++
	}
	return string(l.input[start:l.pos])
}

type AstNode interface{}

type TermNode struct {
	Value string
}

type AndNode struct {
	Left  AstNode
	Right AstNode
}

type OrNode struct {
	Left  AstNode
	Right AstNode
}

type NotNode struct {
	Operand AstNode
}

type QueryParser struct {
	tokens []QueryToken
	pos    int
}

func NewQueryParser(tokens []QueryToken) *QueryParser {
	return &QueryParser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *QueryParser) Parse() (AstNode, error) {
	node, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}
	if p.current().Type != TokEOF {
		return nil, fmt.Errorf("unexpected token at position %d", p.pos)
	}
	return node, nil
}

func (p *QueryParser) parseOrExpr() (AstNode, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokOr {
		p.advance()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &OrNode{Left: left, Right: right}
	}

	return left, nil
}

func (p *QueryParser) parseAndExpr() (AstNode, error) {
	left, err := p.parseNotExpr()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.current().Type
		if tok == TokAnd {
			p.advance()
			right, err := p.parseNotExpr()
			if err != nil {
				return nil, err
			}
			left = &AndNode{Left: left, Right: right}
		} else if tok == TokTerm || tok == TokLParen || tok == TokNot {
			right, err := p.parseNotExpr()
			if err != nil {
				return nil, err
			}
			left = &AndNode{Left: left, Right: right}
		} else {
			break
		}
	}

	return left, nil
}

func (p *QueryParser) parseNotExpr() (AstNode, error) {
	if p.current().Type == TokNot {
		p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &NotNode{Operand: operand}, nil
	}
	return p.parsePrimary()
}

func (p *QueryParser) parsePrimary() (AstNode, error) {
	tok := p.current()
	switch tok.Type {
	case TokTerm:
		p.advance()
		return &TermNode{Value: tok.Value}, nil
	case TokLParen:
		p.advance()
		node, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokRParen {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		p.advance()
		return node, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v", tok)
	}
}

func (p *QueryParser) current() QueryToken {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return QueryToken{Type: TokEOF}
}

func (p *QueryParser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func ParseQuery(input string) (AstNode, error) {
	lexer := NewQueryLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	if len(tokens) == 1 && tokens[0].Type == TokEOF {
		return nil, fmt.Errorf("empty query")
	}
	parser := NewQueryParser(tokens)
	return parser.Parse()
}
