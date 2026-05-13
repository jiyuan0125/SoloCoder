package template

import (
	"strings"
)

type Node interface{}

type TextNode struct {
	Content string
}

type VarNode struct {
	Name         string
	DefaultValue string
	Line         int
	Column       int
}

type IfNode struct {
	Condition string
	Children  []Node
	Line      int
	Column    int
}

type PartialNode struct {
	Name   string
	Line   int
	Column int
}

type Parser struct {
	lexer    *Lexer
	curToken Token
	filename string
}

func NewParser(input, filename string) *Parser {
	lexer := NewLexer(input)
	return &Parser{lexer: lexer, filename: filename}
}

func (p *Parser) Parse() ([]Node, error) {
	p.advance()
	return p.parseNodes(nil)
}

func (p *Parser) parseNodes(stopAt func(Token) bool) ([]Node, error) {
	var nodes []Node

	for p.curToken.Type != TokenClose {
		if stopAt != nil && stopAt(p.curToken) {
			break
		}

		switch p.curToken.Type {
		case TokenText:
			nodes = append(nodes, &TextNode{Content: p.curToken.Literal})
			p.advance()
		case TokenVar:
			name, defaultValue := p.parseVarLiteral(p.curToken.Literal)
			nodes = append(nodes, &VarNode{
				Name:         name,
				DefaultValue: defaultValue,
				Line:         p.curToken.Line,
				Column:       p.curToken.Column,
			})
			p.advance()
		case TokenIf:
			cond := p.curToken.Literal
			line, col := p.curToken.Line, p.curToken.Column
			p.advance()
			children, err := p.parseNodes(func(t Token) bool {
				return t.Type == TokenEndIf
			})
			if err != nil {
				return nil, err
			}
			if p.curToken.Type != TokenEndIf {
				return nil, &SyntaxError{
					File:   p.filename,
					Line:   line,
					Column: col,
					Msg:    "unclosed if block, missing {{/if}}",
				}
			}
			nodes = append(nodes, &IfNode{
				Condition: cond,
				Children:  children,
				Line:      line,
				Column:    col,
			})
			p.advance()
		case TokenPartial:
			nodes = append(nodes, &PartialNode{
				Name:   p.curToken.Literal,
				Line:   p.curToken.Line,
				Column: p.curToken.Column,
			})
			p.advance()
		case TokenEndIf:
			return nil, &SyntaxError{
				File:   p.filename,
				Line:   p.curToken.Line,
				Column: p.curToken.Column,
				Msg:    "unexpected {{/if}}, no matching {{#if}}",
			}
		default:
			p.advance()
		}
	}

	return nodes, nil
}

func (p *Parser) parseVarLiteral(lit string) (string, string) {
	if idx := strings.Index(lit, "|"); idx != -1 {
		return strings.TrimSpace(lit[:idx]), strings.TrimSpace(lit[idx+1:])
	}
	return lit, ""
}

func (p *Parser) advance() {
	p.curToken = p.lexer.NextToken()
}
