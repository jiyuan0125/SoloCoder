package regex

import (
	"fmt"
)

type ASTNode interface {
	String() string
}

type CharNode struct {
	Value rune
}

func (n *CharNode) String() string {
	return fmt.Sprintf("Char(%q)", n.Value)
}

type ConcatNode struct {
	Left  ASTNode
	Right ASTNode
}

func (n *ConcatNode) String() string {
	return fmt.Sprintf("Concat(%s, %s)", n.Left, n.Right)
}

type AltNode struct {
	Left  ASTNode
	Right ASTNode
}

func (n *AltNode) String() string {
	return fmt.Sprintf("Alt(%s, %s)", n.Left, n.Right)
}

type StarNode struct {
	Child ASTNode
}

func (n *StarNode) String() string {
	return fmt.Sprintf("Star(%s)", n.Child)
}

type PlusNode struct {
	Child ASTNode
}

func (n *PlusNode) String() string {
	return fmt.Sprintf("Plus(%s)", n.Child)
}

type QuestionNode struct {
	Child ASTNode
}

func (n *QuestionNode) String() string {
	return fmt.Sprintf("Question(%s)", n.Child)
}

type EmptyNode struct{}

func (n *EmptyNode) String() string {
	return "Empty"
}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek(offset int) Token {
	if p.pos+offset >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos+offset]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) expect(t TokenType) error {
	if p.current().Type != t {
		return fmt.Errorf("pos %d: expected token %v, got %v", p.current().Pos, t, p.current().Type)
	}
	p.advance()
	return nil
}

func (p *Parser) Parse() (ASTNode, error) {
	node, err := p.parseAlt()
	if err != nil {
		return nil, err
	}
	if p.current().Type != TokenEOF {
		return nil, fmt.Errorf("pos %d: unexpected token %v", p.current().Pos, p.current().Type)
	}
	if node == nil {
		return &EmptyNode{}, nil
	}
	return node, nil
}

func (p *Parser) parseAlt() (ASTNode, error) {
	left, err := p.parseConcat()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenPipe {
		p.advance()
		right, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, fmt.Errorf("pos %d: '|' requires right operand", p.current().Pos)
		}
		left = &AltNode{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseConcat() (ASTNode, error) {
	var left ASTNode

	for {
		node, err := p.parsePostfix()
		if err != nil {
			return nil, err
		}
		if node == nil {
			break
		}

		if left == nil {
			left = node
		} else {
			left = &ConcatNode{Left: left, Right: node}
		}
	}

	return left, nil
}

func (p *Parser) parsePostfix() (ASTNode, error) {
	child, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	if child == nil {
		return nil, nil
	}

	switch p.current().Type {
	case TokenStar:
		p.advance()
		return &StarNode{Child: child}, nil
	case TokenPlus:
		p.advance()
		return &PlusNode{Child: child}, nil
	case TokenQuestion:
		p.advance()
		return &QuestionNode{Child: child}, nil
	default:
		return child, nil
	}
}

func (p *Parser) parseAtom() (ASTNode, error) {
	tok := p.current()

	switch tok.Type {
	case TokenChar:
		p.advance()
		return &CharNode{Value: tok.Value}, nil
	case TokenLParen:
		p.advance()
		node, err := p.parseAlt()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		if node == nil {
			return &EmptyNode{}, nil
		}
		return node, nil
	case TokenPipe, TokenStar, TokenPlus, TokenQuestion, TokenRParen, TokenEOF:
		return nil, nil
	default:
		return nil, fmt.Errorf("pos %d: unexpected token %v", tok.Pos, tok.Type)
	}
}
