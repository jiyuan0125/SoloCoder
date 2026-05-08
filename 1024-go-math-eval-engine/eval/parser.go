package eval

import (
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	LOGICAL_OR
	LOGICAL_AND
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[TokenType]int{
	TOKEN_OR:       LOGICAL_OR,
	TOKEN_AND:      LOGICAL_AND,
	TOKEN_EQ:       EQUALS,
	TOKEN_NEQ:      EQUALS,
	TOKEN_LT:       LESSGREATER,
	TOKEN_GT:       LESSGREATER,
	TOKEN_LTE:      LESSGREATER,
	TOKEN_GTE:      LESSGREATER,
	TOKEN_PLUS:     SUM,
	TOKEN_MINUS:    SUM,
	TOKEN_DIVIDE:   PRODUCT,
	TOKEN_MULTIPLY: PRODUCT,
	TOKEN_LPAREN:   CALL,
}

type prefixParseFn func() Expression
type infixParseFn func(Expression) Expression

type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	errors    []string
	prefixFns map[TokenType]prefixParseFn
	infixFns  map[TokenType]infixParseFn
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []string{},
	}
	p.prefixFns = make(map[TokenType]prefixParseFn)
	p.prefixFns[TOKEN_NUMBER_INT] = p.parseIntegerLiteral
	p.prefixFns[TOKEN_NUMBER_FLOAT] = p.parseFloatLiteral
	p.prefixFns[TOKEN_TRUE] = p.parseBooleanLiteral
	p.prefixFns[TOKEN_FALSE] = p.parseBooleanLiteral
	p.prefixFns[TOKEN_IDENTIFIER] = p.parseIdentifier
	p.prefixFns[TOKEN_MINUS] = p.parsePrefixExpression
	p.prefixFns[TOKEN_PLUS] = p.parsePrefixExpression
	p.prefixFns[TOKEN_LPAREN] = p.parseGroupedExpression
	p.infixFns = make(map[TokenType]infixParseFn)
	p.infixFns[TOKEN_PLUS] = p.parseInfixExpression
	p.infixFns[TOKEN_MINUS] = p.parseInfixExpression
	p.infixFns[TOKEN_MULTIPLY] = p.parseInfixExpression
	p.infixFns[TOKEN_DIVIDE] = p.parseInfixExpression
	p.infixFns[TOKEN_EQ] = p.parseInfixExpression
	p.infixFns[TOKEN_NEQ] = p.parseInfixExpression
	p.infixFns[TOKEN_LT] = p.parseInfixExpression
	p.infixFns[TOKEN_GT] = p.parseInfixExpression
	p.infixFns[TOKEN_LTE] = p.parseInfixExpression
	p.infixFns[TOKEN_GTE] = p.parseInfixExpression
	p.infixFns[TOKEN_AND] = p.parseInfixExpression
	p.infixFns[TOKEN_OR] = p.parseInfixExpression
	p.infixFns[TOKEN_LPAREN] = p.parseCallExpression
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Expression = p.parseExpression(LOWEST)
	if !p.curTokenIs(TOKEN_EOF) {
		p.errors = append(p.errors, "unexpected token after expression")
	}
	return program
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	for !p.peekTokenIs(TOKEN_EOF) && precedence < p.peekPrecedence() {
		infix := p.infixFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseBooleanLiteral() Expression {
	return &BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(TOKEN_TRUE)}
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(TOKEN_RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}
	if p.peekTokenIs(TOKEN_RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(TOKEN_RPAREN) {
		return nil
	}
	return args
}
