package xpathengine

import (
	"errors"
	"strconv"
)

type AxisType int

const (
	AxisChild AxisType = iota
	AxisDescendant
	AxisDescendantOrSelf
	AxisParent
	AxisSelf
	AxisAttribute
	AxisFollowingSibling
)

type NodeTest struct {
	Type    NodeTestType
	QName   QName
	IsWild  bool
}

type NodeTestType int

const (
	NodeTestName NodeTestType = iota
	NodeTestText
	NodeTestNode
	NodeTestAll
)

type Predicate struct {
	Expr Expr
}

type Expr interface{}

type LocationPath struct {
	Absolute bool
	Steps    []*Step
}

type Step struct {
	Axis       AxisType
	NodeTest   *NodeTest
	Predicates []*Predicate
}

type LiteralExpr struct {
	Value string
}

type NumberExpr struct {
	Value float64
}

type BinaryOpExpr struct {
	Op    TokenType
	Left  Expr
	Right Expr
}

type FunctionCallExpr struct {
	Name QName
	Args []Expr
}

type PathExpr struct {
	Filter   Expr
	Relative *LocationPath
}

type UnaryExpr struct {
	Op    TokenType
	Value Expr
}

type XPathParser struct {
	lexer *Lexer
	cur   *Token
	prev  *Token
}

func NewXPathParser(input string) *XPathParser {
	return &XPathParser{
		lexer: NewLexer(input),
	}
}

func (p *XPathParser) next() error {
	p.prev = p.cur
	tok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.cur = tok
	return nil
}

func (p *XPathParser) peek() (*Token, error) {
	backupPos := p.lexer.pos
	backupReadPos := p.lexer.readPos
	backupCh := p.lexer.ch
	backupLine := p.lexer.line
	
	tok, err := p.lexer.NextToken()
	
	p.lexer.pos = backupPos
	p.lexer.readPos = backupReadPos
	p.lexer.ch = backupCh
	p.lexer.line = backupLine
	
	return tok, err
}

func (p *XPathParser) Parse() (Expr, error) {
	if err := p.next(); err != nil {
		return nil, err
	}
	if p.cur.Type == TokenEOF {
		return nil, errors.New("empty XPath expression")
	}
	return p.parseExpr()
}

func (p *XPathParser) parseExpr() (Expr, error) {
	return p.parseOrExpr()
}

func (p *XPathParser) parseOrExpr() (Expr, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenOr {
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: TokenOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parseAndExpr() (Expr, error) {
	left, err := p.parseEqualityExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenAnd {
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseEqualityExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: TokenAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parseEqualityExpr() (Expr, error) {
	left, err := p.parseRelationalExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenEqual || p.cur.Type == TokenNotEqual {
		op := p.cur.Type
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseRelationalExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parseRelationalExpr() (Expr, error) {
	left, err := p.parseAdditiveExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenLess || p.cur.Type == TokenLessEqual ||
		p.cur.Type == TokenGreater || p.cur.Type == TokenGreaterEqual {
		op := p.cur.Type
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseAdditiveExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parseAdditiveExpr() (Expr, error) {
	left, err := p.parseMultiplicativeExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenPlus || p.cur.Type == TokenMinus {
		op := p.cur.Type
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseMultiplicativeExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parseMultiplicativeExpr() (Expr, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenAsterisk {
		op := p.cur.Type
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parseUnaryExpr() (Expr, error) {
	if p.cur.Type == TokenMinus {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: TokenMinus, Value: val}, nil
	}
	return p.parseUnionExpr()
}

func (p *XPathParser) parseUnionExpr() (Expr, error) {
	left, err := p.parsePathExpr()
	if err != nil {
		return nil, err
	}
	
	for p.cur.Type == TokenPipe {
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parsePathExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpExpr{Op: TokenPipe, Left: left, Right: right}
	}
	return left, nil
}

func (p *XPathParser) parsePathExpr() (Expr, error) {
	if p.cur.Type == TokenSlash {
		if err := p.next(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenEOF {
			return &LocationPath{Absolute: true, Steps: []*Step{}}, nil
		}
		steps, err := p.parseRelativeLocationPath()
		if err != nil {
			return nil, err
		}
		return &LocationPath{Absolute: true, Steps: steps}, nil
	}
	
	if p.cur.Type == TokenDoubleSlash {
		if err := p.next(); err != nil {
			return nil, err
		}
		steps, err := p.parseRelativeLocationPath()
		if err != nil {
			return nil, err
		}
		
		descendantStep := &Step{
			Axis:       AxisDescendantOrSelf,
			NodeTest:   &NodeTest{Type: NodeTestNode},
			Predicates: []*Predicate{},
		}
		newSteps := append([]*Step{descendantStep}, steps...)
		return &LocationPath{Absolute: true, Steps: newSteps}, nil
	}
	
	filter, err := p.parseFilterExpr()
	if err != nil {
		return nil, err
	}
	
	if p.cur.Type == TokenSlash || p.cur.Type == TokenDoubleSlash {
		isDoubleSlash := p.cur.Type == TokenDoubleSlash
		if err := p.next(); err != nil {
			return nil, err
		}
		steps, err := p.parseRelativeLocationPath()
		if err != nil {
			return nil, err
		}
		
		if isDoubleSlash {
			descendantStep := &Step{
				Axis:       AxisDescendantOrSelf,
				NodeTest:   &NodeTest{Type: NodeTestNode},
				Predicates: []*Predicate{},
			}
			steps = append([]*Step{descendantStep}, steps...)
		}
		
		return &PathExpr{Filter: filter, Relative: &LocationPath{Absolute: false, Steps: steps}}, nil
	}
	
	return filter, nil
}

func (p *XPathParser) parseFilterExpr() (Expr, error) {
	if p.cur.Type == TokenLParen {
		if err := p.next(); err != nil {
			return nil, err
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.cur.Type != TokenRParen {
			return nil, errors.New("expected )")
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		
		predicates, err := p.parsePredicates()
		if err != nil {
			return nil, err
		}
		
		if len(predicates) > 0 {
			return &PathExpr{
				Filter: expr,
				Relative: &LocationPath{
					Absolute: false,
					Steps: []*Step{{
						Axis:       AxisSelf,
						NodeTest:   &NodeTest{Type: NodeTestNode},
						Predicates: predicates,
					}},
				},
			}, nil
		}
		return expr, nil
	}
	
	if p.cur.Type == TokenFunction {
		return p.parseFunctionCall()
	}
	
	return p.parseLocationPath()
}

func (p *XPathParser) parseLocationPath() (Expr, error) {
	if p.isStepStart() {
		steps, err := p.parseRelativeLocationPath()
		if err != nil {
			return nil, err
		}
		return &LocationPath{Absolute: false, Steps: steps}, nil
	}
	
	if p.cur.Type == TokenDot {
		if err := p.next(); err != nil {
			return nil, err
		}
		predicates, err := p.parsePredicates()
		if err != nil {
			return nil, err
		}
		step := &Step{
			Axis:       AxisSelf,
			NodeTest:   &NodeTest{Type: NodeTestNode},
			Predicates: predicates,
		}
		
		steps := []*Step{step}
		if p.cur.Type == TokenSlash || p.cur.Type == TokenDoubleSlash {
			isDoubleSlash := p.cur.Type == TokenDoubleSlash
			if err := p.next(); err != nil {
				return nil, err
			}
			moreSteps, err := p.parseRelativeLocationPath()
			if err != nil {
				return nil, err
			}
			if isDoubleSlash {
				descendantStep := &Step{
					Axis:       AxisDescendantOrSelf,
					NodeTest:   &NodeTest{Type: NodeTestNode},
					Predicates: []*Predicate{},
				}
				steps = append(steps, descendantStep)
			}
			steps = append(steps, moreSteps...)
		}
		
		return &LocationPath{Absolute: false, Steps: steps}, nil
	}
	
	if p.cur.Type == TokenDoubleDot {
		if err := p.next(); err != nil {
			return nil, err
		}
		predicates, err := p.parsePredicates()
		if err != nil {
			return nil, err
		}
		step := &Step{
			Axis:       AxisParent,
			NodeTest:   &NodeTest{Type: NodeTestNode},
			Predicates: predicates,
		}
		
		steps := []*Step{step}
		if p.cur.Type == TokenSlash || p.cur.Type == TokenDoubleSlash {
			isDoubleSlash := p.cur.Type == TokenDoubleSlash
			if err := p.next(); err != nil {
				return nil, err
			}
			moreSteps, err := p.parseRelativeLocationPath()
			if err != nil {
				return nil, err
			}
			if isDoubleSlash {
				descendantStep := &Step{
					Axis:       AxisDescendantOrSelf,
					NodeTest:   &NodeTest{Type: NodeTestNode},
					Predicates: []*Predicate{},
				}
				steps = append(steps, descendantStep)
			}
			steps = append(steps, moreSteps...)
		}
		
		return &LocationPath{Absolute: false, Steps: steps}, nil
	}
	
	if p.cur.Type == TokenAt {
		step, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		return &LocationPath{Absolute: false, Steps: []*Step{step}}, nil
	}
	
	if p.cur.Type == TokenString {
		val := p.cur.Literal
		if err := p.next(); err != nil {
			return nil, err
		}
		return &LiteralExpr{Value: val}, nil
	}
	
	if p.cur.Type == TokenNumber {
		num, _ := strconv.ParseFloat(p.cur.Literal, 64)
		if err := p.next(); err != nil {
			return nil, err
		}
		return &NumberExpr{Value: num}, nil
	}
	
	return nil, errors.New("unexpected token: " + p.cur.Literal)
}

func (p *XPathParser) isStepStart() bool {
	return p.cur.Type == TokenName || p.cur.Type == TokenAsterisk ||
		p.cur.Type == TokenAt || p.cur.Type == TokenDot ||
		p.cur.Type == TokenDoubleDot || p.cur.Type == TokenFunction
}

func (p *XPathParser) parseRelativeLocationPath() ([]*Step, error) {
	var steps []*Step
	step, err := p.parseStep()
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)
	
	for p.cur.Type == TokenSlash || p.cur.Type == TokenDoubleSlash {
		isDoubleSlash := p.cur.Type == TokenDoubleSlash
		if err := p.next(); err != nil {
			return nil, err
		}
		
		if isDoubleSlash {
			descendantStep := &Step{
				Axis:       AxisDescendantOrSelf,
				NodeTest:   &NodeTest{Type: NodeTestNode},
				Predicates: []*Predicate{},
			}
			steps = append(steps, descendantStep)
		}
		
		step, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	
	return steps, nil
}

func (p *XPathParser) parseStep() (*Step, error) {
	axis := AxisChild
	
	if p.cur.Type == TokenDoubleDot {
		if err := p.next(); err != nil {
			return nil, err
		}
		predicates, err := p.parsePredicates()
		if err != nil {
			return nil, err
		}
		return &Step{
			Axis:       AxisParent,
			NodeTest:   &NodeTest{Type: NodeTestNode},
			Predicates: predicates,
		}, nil
	}
	
	if p.cur.Type == TokenDot {
		if err := p.next(); err != nil {
			return nil, err
		}
		predicates, err := p.parsePredicates()
		if err != nil {
			return nil, err
		}
		return &Step{
			Axis:       AxisSelf,
			NodeTest:   &NodeTest{Type: NodeTestNode},
			Predicates: predicates,
		}, nil
	}
	
	if p.cur.Type == TokenAt {
		axis = AxisAttribute
		if err := p.next(); err != nil {
			return nil, err
		}
	}
	
	nodeTest, err := p.parseNodeTest()
	if err != nil {
		return nil, err
	}
	
	predicates, err := p.parsePredicates()
	if err != nil {
		return nil, err
	}
	
	return &Step{
		Axis:       axis,
		NodeTest:   nodeTest,
		Predicates: predicates,
	}, nil
}

func (p *XPathParser) parseNodeTest() (*NodeTest, error) {
	if p.cur.Type == TokenFunction {
		funcName := p.cur.Literal
		if err := p.next(); err != nil {
			return nil, err
		}
		if p.cur.Type != TokenLParen {
			return nil, errors.New("expected ( after function name")
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		if p.cur.Type != TokenRParen {
			return nil, errors.New("expected )")
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		
		if funcName == "text" {
			return &NodeTest{Type: NodeTestText}, nil
		}
		if funcName == "node" {
			return &NodeTest{Type: NodeTestNode}, nil
		}
		return nil, errors.New("unsupported node test: " + funcName)
	}
	
	if p.cur.Type == TokenAsterisk {
		if err := p.next(); err != nil {
			return nil, err
		}
		return &NodeTest{Type: NodeTestAll, IsWild: true}, nil
	}
	
	if p.cur.Type == TokenName {
		qname := QName{Local: p.cur.Literal}
		if err := p.next(); err != nil {
			return nil, err
		}
		
		if p.cur.Type == TokenColon {
			if err := p.next(); err != nil {
				return nil, err
			}
			if p.cur.Type == TokenAsterisk {
				qname.Prefix = qname.Local
				qname.Local = "*"
				if err := p.next(); err != nil {
					return nil, err
				}
				return &NodeTest{Type: NodeTestName, QName: qname, IsWild: true}, nil
			}
			if p.cur.Type != TokenName {
				return nil, errors.New("expected name after colon")
			}
			qname.Prefix = qname.Local
			qname.Local = p.cur.Literal
			if err := p.next(); err != nil {
				return nil, err
			}
		}
		
		return &NodeTest{Type: NodeTestName, QName: qname}, nil
	}
	
	return nil, errors.New("expected node test")
}

func (p *XPathParser) parsePredicates() ([]*Predicate, error) {
	var predicates []*Predicate
	
	for p.cur.Type == TokenLBracket {
		if err := p.next(); err != nil {
			return nil, err
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.cur.Type != TokenRBracket {
			return nil, errors.New("expected ]")
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		predicates = append(predicates, &Predicate{Expr: expr})
	}
	
	return predicates, nil
}

func (p *XPathParser) parseFunctionCall() (Expr, error) {
	name := QName{Local: p.cur.Literal}
	if err := p.next(); err != nil {
		return nil, err
	}
	if p.cur.Type != TokenLParen {
		return nil, errors.New("expected ( after function name")
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	
	var args []Expr
	if p.cur.Type != TokenRParen {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		
		for p.cur.Type == TokenComma {
			if err := p.next(); err != nil {
				return nil, err
			}
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}
	
	if p.cur.Type != TokenRParen {
		return nil, errors.New("expected )")
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	
	return &FunctionCallExpr{Name: name, Args: args}, nil
}
