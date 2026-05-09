package regex

import "fmt"

type parser struct {
	pattern string
	pos     int
}

func newParser(pattern string) *parser {
	return &parser{pattern: pattern, pos: 0}
}

func (p *parser) peek() byte {
	if p.pos >= len(p.pattern) {
		return 0
	}
	return p.pattern[p.pos]
}

func (p *parser) consume() byte {
	ch := p.peek()
	p.pos++
	return ch
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.pattern)
}

func (p *parser) error(msg string) error {
	return fmt.Errorf("syntax error at position %d: %s", p.pos, msg)
}

func (p *parser) parse() (*Node, error) {
	node, err := p.alternate()
	if err != nil {
		return nil, err
	}
	if !p.atEnd() {
		return nil, p.error("unexpected character '" + string(p.peek()) + "'")
	}
	return node, nil
}

func (p *parser) alternate() (*Node, error) {
	node, err := p.concat()
	if err != nil {
		return nil, err
	}
	
	if node == nil {
		node = newEmptyNode()
	}
	
	for p.peek() == '|' {
		p.consume()
		right, err := p.concat()
		if err != nil {
			return nil, err
		}
		if right == nil {
			right = newEmptyNode()
		}
		node = newAlternateNode(node, right)
	}
	
	return node, nil
}

func (p *parser) concat() (*Node, error) {
	var nodes []*Node
	
	for !p.atEnd() && p.peek() != '|' && p.peek() != ')' {
		node, err := p.repeat()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	
	if len(nodes) == 0 {
		return nil, nil
	}
	
	node := nodes[0]
	for i := 1; i < len(nodes); i++ {
		node = newConcatNode(node, nodes[i])
	}
	
	return node, nil
}

func (p *parser) repeat() (*Node, error) {
	node, err := p.atom()
	if err != nil {
		return nil, err
	}
	
	for p.peek() == '*' {
		p.consume()
		if node == nil {
			return nil, p.error("nothing to repeat")
		}
		node = newStarNode(node)
	}
	
	return node, nil
}

func (p *parser) atom() (*Node, error) {
	ch := p.peek()
	
	if ch == '(' {
		p.consume()
		if p.peek() == ')' {
			p.consume()
			return newEmptyNode(), nil
		}
		
		node, err := p.alternate()
		if err != nil {
			return nil, err
		}
		
		if p.peek() != ')' {
			return nil, p.error("unclosed parenthesis")
		}
		p.consume()
		
		if node == nil {
			return newEmptyNode(), nil
		}
		return node, nil
	}
	
	if ch == ')' {
		return nil, p.error("unmatched closing parenthesis")
	}
	
	if ch == '*' {
		return nil, p.error("nothing to repeat")
	}
	
	if ch == '|' {
		return newEmptyNode(), nil
	}
	
	if ch == '.' {
		p.consume()
		return newDotNode(), nil
	}
	
	if ch == 0 {
		return nil, nil
	}
	
	p.consume()
	return newLiteralNode(ch), nil
}
