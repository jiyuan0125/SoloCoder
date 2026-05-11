package parser

import (
	"fmt"
	"strconv"
	"strings"

	"sqlparser/pkg/parser/ast"
	"sqlparser/pkg/parser/tokenizer"
)

type Parser struct {
	lexer     *tokenizer.Lexer
	curToken  tokenizer.Token
	peekToken tokenizer.Token
	explain   bool
	steps     []string
}

func NewParser(input string) *Parser {
	p := &Parser{lexer: tokenizer.NewLexer(input)}
	p.nextToken()
	p.nextToken()
	return p
}

func NewParserWithExplain(input string) *Parser {
	p := NewParser(input)
	p.explain = true
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) addStep(msg string) {
	if p.explain {
		p.steps = append(p.steps, fmt.Sprintf("[Token=%s Literal=%s] %s", p.curToken.Type, p.curToken.Literal, msg))
	}
}

func (p *Parser) GetSteps() []string {
	return p.steps
}

func (p *Parser) curTokenIs(t tokenizer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t tokenizer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t tokenizer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) expectCur(t tokenizer.TokenType) bool {
	if p.curTokenIs(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) Parse() (ast.Node, error) {
	p.addStep("开始解析")
	
	if p.curTokenIs(tokenizer.TokenIdentifier) && strings.EqualFold(p.curToken.Literal, "WHERE") {
		p.addStep("跳过WHERE关键字")
		p.nextToken()
	}

	if p.curTokenIs(tokenizer.TokenEOF) {
		return nil, fmt.Errorf("empty WHERE clause")
	}

	node, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	if !p.curTokenIs(tokenizer.TokenEOF) {
		return nil, fmt.Errorf("unexpected token at end: %s (%s)", p.curToken.Type, p.curToken.Literal)
	}

	p.addStep("解析完成")
	return node, nil
}

func (p *Parser) parseOrExpr() (ast.Node, error) {
	p.addStep("解析OR表达式")
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.curTokenIs(tokenizer.TokenOr) {
		p.addStep("遇到OR运算符")
		op := ast.OpOr
		p.nextToken()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.LogicalNode{
			Operator: op,
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseAndExpr() (ast.Node, error) {
	p.addStep("解析AND表达式")
	left, err := p.parseNotExpr()
	if err != nil {
		return nil, err
	}

	for p.curTokenIs(tokenizer.TokenAnd) {
		p.addStep("遇到AND运算符")
		op := ast.OpAnd
		p.nextToken()
		right, err := p.parseNotExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.LogicalNode{
			Operator: op,
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseNotExpr() (ast.Node, error) {
	p.addStep("解析NOT表达式")
	
	if p.curTokenIs(tokenizer.TokenNot) {
		p.addStep("遇到NOT运算符")
		p.nextToken()
		node, err := p.parseNotExpr()
		if err != nil {
			return nil, err
		}
		return &ast.NotNode{Operand: node}, nil
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Node, error) {
	p.addStep("解析基本表达式")

	if p.curTokenIs(tokenizer.TokenLParen) {
		p.addStep("遇到左括号，解析括号内表达式")
		p.nextToken()
		node, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if !p.expectCur(tokenizer.TokenRParen) {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		return node, nil
	}

	return p.parseCondition()
}

func (p *Parser) parseCondition() (ast.Node, error) {
	p.addStep("解析条件表达式")

	column, err := p.parseColumn()
	if err != nil {
		return nil, err
	}

	p.addStep(fmt.Sprintf("列名: %s", column))

	if p.curTokenIs(tokenizer.TokenBetween) {
		return p.parseBetween(column, false)
	}

	if p.curTokenIs(tokenizer.TokenIn) {
		return p.parseIn(column, false)
	}

	if p.curTokenIs(tokenizer.TokenNot) {
		p.nextToken()
		if p.curTokenIs(tokenizer.TokenIn) {
			return p.parseIn(column, true)
		}
		if p.curTokenIs(tokenizer.TokenLike) {
			return p.parseLike(column, true)
		}
		if p.curTokenIs(tokenizer.TokenBetween) {
			return p.parseBetween(column, true)
		}
		if p.curTokenIs(tokenizer.TokenNull) {
			return p.parseNullCheck(column, true)
		}
		return nil, fmt.Errorf("unexpected NOT")
	}

	if p.curTokenIs(tokenizer.TokenLike) {
		return p.parseLike(column, false)
	}

	if p.curTokenIs(tokenizer.TokenIs) {
		return p.parseIs(column)
	}

	return p.parseComparison(column)
}

func (p *Parser) parseColumn() (string, error) {
	if p.curTokenIs(tokenizer.TokenIdentifier) {
		col := p.curToken.Literal
		p.nextToken()
		return col, nil
	}
	if p.curTokenIs(tokenizer.TokenQuotedIdent) {
		col := p.curToken.Literal
		p.nextToken()
		return col, nil
	}
	return "", fmt.Errorf("expected column name, got %s", p.curToken.Type)
}

func (p *Parser) parseComparison(column string) (ast.Node, error) {
	p.addStep("解析比较表达式")

	var op ast.ComparisonOperator
	switch p.curToken.Type {
	case tokenizer.TokenEq:
		op = ast.OpEq
	case tokenizer.TokenNeq:
		op = ast.OpNeq
	case tokenizer.TokenGt:
		op = ast.OpGt
	case tokenizer.TokenLt:
		op = ast.OpLt
	case tokenizer.TokenGte:
		op = ast.OpGte
	case tokenizer.TokenLte:
		op = ast.OpLte
	default:
		return nil, fmt.Errorf("expected comparison operator, got %s", p.curToken.Type)
	}

	p.nextToken()
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.addStep(fmt.Sprintf("比较: %s %s %s", column, op, value))

	return &ast.ComparisonNode{
		Column:   column,
		Operator: op,
		Value:    value,
	}, nil
}

func (p *Parser) parseBetween(column string, not bool) (ast.Node, error) {
	p.addStep(fmt.Sprintf("解析BETWEEN表达式 (not=%v)", not))
	p.nextToken()

	lower, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	if !p.expectCur(tokenizer.TokenAnd) {
		return nil, fmt.Errorf("expected AND in BETWEEN clause")
	}
	p.addStep("BETWEEN中的AND（非逻辑运算符）")

	upper, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.addStep(fmt.Sprintf("BETWEEN: %s BETWEEN %s AND %s (not=%v)", column, lower, upper, not))

	return &ast.BetweenNode{
		Column: column,
		Not:    not,
		Lower:  lower,
		Upper:  upper,
	}, nil
}

func (p *Parser) parseIn(column string, not bool) (ast.Node, error) {
	p.addStep(fmt.Sprintf("解析IN表达式 (not=%v)", not))
	p.nextToken()

	if !p.expectCur(tokenizer.TokenLParen) {
		return nil, fmt.Errorf("expected ( after IN")
	}

	values := []ast.Value{}

	if p.curTokenIs(tokenizer.TokenRParen) {
		return nil, fmt.Errorf("empty IN list is not allowed")
	}

	for {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, val)

		if p.curTokenIs(tokenizer.TokenComma) {
			p.nextToken()
		} else {
			break
		}
	}

	if !p.expectCur(tokenizer.TokenRParen) {
		return nil, fmt.Errorf("expected ) after IN list")
	}

	p.addStep(fmt.Sprintf("IN: %s IN (%d values) (not=%v)", column, len(values), not))

	return &ast.InNode{
		Column: column,
		Not:    not,
		Values: values,
	}, nil
}

func (p *Parser) parseLike(column string, not bool) (ast.Node, error) {
	p.addStep(fmt.Sprintf("解析LIKE表达式 (not=%v)", not))
	p.nextToken()

	if !p.curTokenIs(tokenizer.TokenString) {
		return nil, fmt.Errorf("expected string pattern for LIKE")
	}

	pattern := p.curToken.Literal
	escape := ""
	p.nextToken()

	if p.curTokenIs(tokenizer.TokenEscape) {
		p.addStep("解析ESCAPE子句")
		p.nextToken()
		if !p.curTokenIs(tokenizer.TokenString) {
			return nil, fmt.Errorf("expected string for ESCAPE")
		}
		escape = p.curToken.Literal
		p.nextToken()
	}

	p.addStep(fmt.Sprintf("LIKE: %s LIKE '%s' (escape='%s', not=%v)", column, pattern, escape, not))

	return &ast.LikeNode{
		Column:  column,
		Not:     not,
		Pattern: pattern,
		Escape:  escape,
	}, nil
}

func (p *Parser) parseIs(column string) (ast.Node, error) {
	p.addStep("解析IS表达式")
	p.nextToken()

	not := false
	if p.curTokenIs(tokenizer.TokenNot) {
		not = true
		p.addStep("IS NOT")
		p.nextToken()
	}

	if !p.curTokenIs(tokenizer.TokenNull) {
		return nil, fmt.Errorf("expected NULL after IS")
	}
	p.nextToken()

	p.addStep(fmt.Sprintf("IS NULL: %s IS NULL (not=%v)", column, not))

	return &ast.NullCheckNode{
		Column: column,
		Not:    not,
	}, nil
}

func (p *Parser) parseNullCheck(column string, not bool) (ast.Node, error) {
	p.addStep(fmt.Sprintf("解析NULL检查 (not=%v)", not))
	p.nextToken()

	p.addStep(fmt.Sprintf("IS NULL: %s IS NULL (not=%v)", column, not))

	return &ast.NullCheckNode{
		Column: column,
		Not:    not,
	}, nil
}

func (p *Parser) parseValue() (ast.Value, error) {
	p.addStep("解析值")

	if p.curTokenIs(tokenizer.TokenNumber) {
		num, err := strconv.ParseFloat(p.curToken.Literal, 64)
		if err != nil {
			return ast.Value{}, fmt.Errorf("invalid number: %s", p.curToken.Literal)
		}
		p.nextToken()
		return ast.NewNumberValue(num), nil
	}

	if p.curTokenIs(tokenizer.TokenString) {
		str := p.curToken.Literal
		p.nextToken()
		return ast.NewStringValue(str), nil
	}

	if p.curTokenIs(tokenizer.TokenNull) {
		p.nextToken()
		return ast.NewNullValue(), nil
	}

	return ast.Value{}, fmt.Errorf("unexpected value type: %s", p.curToken.Type)
}

func Parse(input string) (ast.Node, error) {
	p := NewParser(input)
	return p.Parse()
}

func ParseWithExplain(input string) (ast.Node, []string, error) {
	p := NewParserWithExplain(input)
	node, err := p.Parse()
	return node, p.GetSteps(), err
}
