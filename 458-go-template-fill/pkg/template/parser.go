package template

import (
	"fmt"
	"strings"
)

type NodeType int

const (
	NodeText NodeType = iota
	NodeVariable
	NodeIf
	NodeEach
	NodeRoot
)

type Node interface {
	Type() NodeType
}

type TextNode struct {
	Content string
}

func (n *TextNode) Type() NodeType { return NodeText }

type VariableNode struct {
	Name         string
	DefaultValue string
	Raw          string
}

func (n *VariableNode) Type() NodeType { return NodeVariable }

type IfNode struct {
	Condition string
	Then      []Node
	Else      []Node
}

func (n *IfNode) Type() NodeType { return NodeIf }

type EachNode struct {
	ArrayName string
	Body      []Node
}

func (n *EachNode) Type() NodeType { return NodeEach }

type RootNode struct {
	Children []Node
}

func (n *RootNode) Type() NodeType { return NodeRoot }

type Parser struct {
	tokens  []Token
	pos     int
	warnings []string
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:  tokens,
		pos:     0,
		warnings: []string{},
	}
}

func (p *Parser) Parse() (*RootNode, error) {
	root := &RootNode{Children: []Node{}}
	
	for p.pos < len(p.tokens) {
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			root.Children = append(root.Children, node)
		}
	}
	
	return root, nil
}

func (p *Parser) parseNode() (Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, nil
	}
	
	token := p.tokens[p.pos]
	
	switch token.Type {
	case TokenText:
		p.pos++
		return &TextNode{Content: token.Value}, nil
		
	case TokenVariable:
		p.pos++
		return &VariableNode{
			Name:         token.Value,
			DefaultValue: "",
			Raw:          token.Raw,
		}, nil
		
	case TokenDefaultValue:
		p.pos++
		parts := strings.SplitN(token.Value, "|", 2)
		varName := parts[0]
		defaultValue := ""
		if len(parts) > 1 {
			defaultValue = parts[1]
		}
		return &VariableNode{
			Name:         varName,
			DefaultValue: defaultValue,
			Raw:          token.Raw,
		}, nil
		
	case TokenIfStart:
		return p.parseIf()
		
	case TokenElse:
		return nil, fmt.Errorf("语法错误: 意外的else标签 (第%d行) - else必须在if块内部", token.Line)
		
	case TokenEnd:
		return nil, fmt.Errorf("语法错误: 意外的end标签 (第%d行) - 没有对应的if或each", token.Line)
		
	case TokenEachStart:
		return p.parseEach()
		
	default:
		p.pos++
		return nil, nil
	}
}

func (p *Parser) parseIf() (Node, error) {
	ifToken := p.tokens[p.pos]
	p.pos++
	
	ifNode := &IfNode{
		Condition: ifToken.Value,
		Then:      []Node{},
		Else:      []Node{},
	}
	
	inElse := false
	
	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]
		
		if token.Type == TokenElse {
			if inElse {
				return nil, fmt.Errorf("语法错误: 重复的else标签 (第%d行)", token.Line)
			}
			inElse = true
			p.pos++
			continue
		}
		
		if token.Type == TokenEnd {
			p.pos++
			return ifNode, nil
		}
		
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		
		if node != nil {
			if inElse {
				ifNode.Else = append(ifNode.Else, node)
			} else {
				ifNode.Then = append(ifNode.Then, node)
			}
		}
	}
	
	return nil, fmt.Errorf("语法错误: if标签未闭合 (第%d行: {{if %s}})", ifToken.Line, ifToken.Value)
}

func (p *Parser) parseEach() (Node, error) {
	eachToken := p.tokens[p.pos]
	p.pos++
	
	eachNode := &EachNode{
		ArrayName: eachToken.Value,
		Body:      []Node{},
	}
	
	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]
		
		if token.Type == TokenEnd {
			p.pos++
			return eachNode, nil
		}
		
		if token.Type == TokenElse {
			return nil, fmt.Errorf("语法错误: each块中不允许else标签 (第%d行)", token.Line)
		}
		
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		
		if node != nil {
			eachNode.Body = append(eachNode.Body, node)
		}
	}
	
	return nil, fmt.Errorf("语法错误: each标签未闭合 (第%d行: {{each %s}})", eachToken.Line, eachToken.Value)
}
