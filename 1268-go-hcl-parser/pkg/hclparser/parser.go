package hclparser

import (
	"fmt"
)

type Parser struct {
	lexer   *Lexer
	curToken Token
	peekToken Token
	errors  []string
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []string{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	tok, err := p.lexer.NextToken()
	if err != nil {
		p.errors = append(p.errors, err.Error())
		p.peekToken = NewToken(TokenEOF, "", 0, 0, nil)
		return
	}
	p.peekToken = tok
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at line %d, column %d",
		t.String(), p.peekToken.Type.String(), p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Parse() (*Body, error) {
	body := &Body{
		Statements: []Statement{},
	}

	for p.curToken.Type != TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
		p.nextToken()
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}

	return body, nil
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case TokenIdentifier:
		return p.parseIdentifierStatement()
	default:
		return nil
	}
}

func (p *Parser) parseIdentifierStatement() Statement {
	firstIdent := p.curToken.Literal

	if p.peekToken.Type == TokenLBrace {
		return p.parseBlock(firstIdent, []string{})
	}

	if p.peekToken.Type == TokenString || p.peekToken.Type == TokenIdentifier {
		labels := []string{}
		if p.peekToken.Type == TokenString {
			p.nextToken()
			labels = append(labels, p.curToken.RawValue.(string))
		} else {
			p.nextToken()
			labels = append(labels, p.curToken.Literal)
		}

		for p.peekToken.Type == TokenString || p.peekToken.Type == TokenIdentifier {
			if p.peekToken.Type == TokenString {
				p.nextToken()
				labels = append(labels, p.curToken.RawValue.(string))
			} else {
				p.nextToken()
				labels = append(labels, p.curToken.Literal)
			}
		}

		if p.peekToken.Type == TokenLBrace {
			return p.parseBlock(firstIdent, labels)
		}
	}

	if p.peekToken.Type == TokenEquals {
		p.nextToken()
		p.nextToken()
		value := p.parseExpression(lowestPrec)
		if value == nil {
			return nil
		}
		return &Attribute{
			Key:   firstIdent,
			Value: value,
		}
	}

	return nil
}

func (p *Parser) parseBlock(typ string, labels []string) Statement {
	if !p.expectPeek(TokenLBrace) {
		return nil
	}

	block := &Block{
		Type:   typ,
		Labels: labels,
		Body:   &Body{Statements: []Statement{}},
	}

	p.nextToken()

	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Body.Statements = append(block.Body.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

const (
	lowestPrec      = 1
	equalsPrec      = 2
	lessgreaterPrec = 3
	sumPrec         = 4
	productPrec     = 5
	prefixPrec      = 6
	callPrec        = 7
	indexPrec       = 8
)

func (p *Parser) tokenPrecedence(t TokenType) int {
	switch t {
	case TokenEqualEqual, TokenNotEqual:
		return equalsPrec
	case TokenLess, TokenLessEqual, TokenGreater, TokenGreaterEqual:
		return lessgreaterPrec
	case TokenPlus, TokenMinus:
		return sumPrec
	case TokenStar, TokenSlash, TokenPercent:
		return productPrec
	case TokenLParen:
		return callPrec
	case TokenLBracket:
		return indexPrec
	default:
		return lowestPrec
	}
}

func (p *Parser) parseExpression(precedence int) Expression {
	var leftExp Expression

	switch p.curToken.Type {
	case TokenIdentifier:
		leftExp = p.parseIdentifierExpression()
	case TokenString:
		leftExp = &LiteralValue{Value: p.curToken.RawValue.(string)}
	case TokenNumber:
		leftExp = &LiteralValue{Value: p.curToken.RawValue}
	case TokenBoolean:
		leftExp = &LiteralValue{Value: p.curToken.RawValue.(bool)}
	case TokenNull:
		leftExp = &LiteralValue{Value: nil}
	case TokenHeredoc:
		meta := p.curToken.RawValue.(map[string]interface{})
		leftExp = &HeredocValue{
			Delimiter:   meta["delimiter"].(string),
			Content:     p.curToken.Literal,
			StripIndent: meta["stripIndent"].(bool),
			HasInterp:   meta["hasInterp"].(bool),
		}
	case TokenMinus, TokenExclamation:
		leftExp = p.parsePrefixExpression()
	case TokenLBracket:
		leftExp = p.parseListExpression()
	case TokenLBrace:
		leftExp = p.parseMapExpression()
	case TokenLParen:
		p.nextToken()
		leftExp = p.parseExpression(lowestPrec)
		if !p.expectPeek(TokenRParen) {
			return nil
		}
	default:
		return nil
	}

	for p.peekToken.Type != TokenQuestion &&
		p.peekToken.Type != TokenRBrace &&
		p.peekToken.Type != TokenRBracket &&
		p.peekToken.Type != TokenRParen &&
		p.peekToken.Type != TokenComma &&
		p.peekToken.Type != TokenEOF &&
		precedence < p.tokenPrecedence(p.peekToken.Type) {

		if p.peekToken.Type == TokenLParen {
			if fn, ok := leftExp.(*VariableReference); ok {
				leftExp = p.parseFunctionCall(fn.Path[0])
			} else {
				return leftExp
			}
		} else if p.peekToken.Type == TokenLBracket {
			leftExp = p.parseIndexExpression(leftExp)
		} else if p.peekToken.Type == TokenDot {
			leftExp = p.parseDotExpression(leftExp)
		} else {
			leftExp = p.parseInfixExpression(leftExp)
		}
	}

	if p.peekToken.Type == TokenQuestion {
		return p.parseConditionalExpression(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifierExpression() Expression {
	path := []string{p.curToken.Literal}

	for p.peekToken.Type == TokenDot {
		p.nextToken()
		if p.peekToken.Type == TokenIdentifier {
			p.nextToken()
			path = append(path, p.curToken.Literal)
		} else {
			break
		}
	}

	if p.peekToken.Type == TokenLParen {
		return p.parseFunctionCall(path[0])
	}

	return &VariableReference{Path: path}
}

func (p *Parser) parseDotExpression(left Expression) Expression {
	if vr, ok := left.(*VariableReference); ok {
		for p.peekToken.Type == TokenDot {
			p.nextToken()
			if p.peekToken.Type == TokenIdentifier {
				p.nextToken()
				vr.Path = append(vr.Path, p.curToken.Literal)
			} else {
				break
			}
		}
		return vr
	}
	return left
}

func (p *Parser) parsePrefixExpression() Expression {
	op := p.curToken.Type.String()
	p.nextToken()
	return &UnaryOp{
		Op:      op,
		Operand: p.parseExpression(prefixPrec),
	}
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	op := p.peekToken.Type.String()
	precedence := p.tokenPrecedence(p.peekToken.Type)
	p.nextToken()
	p.nextToken()
	right := p.parseExpression(precedence)
	return &BinaryOp{
		Left:  left,
		Op:    op,
		Right: right,
	}
}

func (p *Parser) parseListExpression() Expression {
	list := &ListExpr{Elements: []Expression{}}

	if p.peekToken.Type == TokenRBracket {
		p.nextToken()
		return list
	}

	p.nextToken()
	list.Elements = append(list.Elements, p.parseExpression(lowestPrec))

	for p.peekToken.Type == TokenComma {
		p.nextToken()
		if p.peekToken.Type == TokenRBracket {
			p.nextToken()
			break
		}
		p.nextToken()
		list.Elements = append(list.Elements, p.parseExpression(lowestPrec))
	}

	if !p.expectPeek(TokenRBracket) {
		return nil
	}

	return list
}

func (p *Parser) parseMapExpression() Expression {
	m := &MapExpr{Elements: map[string]Expression{}}

	if p.peekToken.Type == TokenRBrace {
		p.nextToken()
		return m
	}

	p.nextToken()

	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		key := p.curToken.Literal
		if p.curToken.Type == TokenString {
			key = p.curToken.RawValue.(string)
		}

		if !p.expectPeek(TokenEquals) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(lowestPrec)
		if value == nil {
			return nil
		}

		m.Elements[key] = value

		if p.peekToken.Type == TokenComma {
			p.nextToken()
			if p.peekToken.Type == TokenRBrace {
				p.nextToken()
				break
			}
			p.nextToken()
		} else {
			break
		}
	}

	return m
}

func (p *Parser) parseFunctionCall(name string) Expression {
	fn := &FunctionCall{
		Name:          name,
		PositionalArgs: []Expression{},
		NamedArgs:     map[string]Expression{},
	}

	if !p.expectPeek(TokenLParen) {
		return nil
	}

	if p.peekToken.Type == TokenRParen {
		p.nextToken()
		return fn
	}

	p.nextToken()

	hasNamed := false

	for {
		arg := p.parseExpression(lowestPrec)
		if arg == nil {
			return nil
		}

		if p.peekToken.Type == TokenEquals {
			if vr, ok := arg.(*VariableReference); ok && len(vr.Path) == 1 {
				hasNamed = true
				p.nextToken()
				p.nextToken()
				value := p.parseExpression(lowestPrec)
				if value == nil {
					return nil
				}
				fn.NamedArgs[vr.Path[0]] = value
			} else {
				return nil
			}
		} else {
			if hasNamed {
				p.errors = append(p.errors, "positional argument must come before named arguments")
				return nil
			}
			fn.PositionalArgs = append(fn.PositionalArgs, arg)
		}

		if p.peekToken.Type == TokenComma {
			p.nextToken()
			if p.peekToken.Type == TokenRParen {
				p.nextToken()
				break
			}
			p.nextToken()
		} else if p.peekToken.Type == TokenRParen {
			p.nextToken()
			break
		} else {
			return nil
		}
	}

	return fn
}

func (p *Parser) parseIndexExpression(left Expression) Expression {
	p.nextToken()
	p.nextToken()
	idx := p.parseExpression(lowestPrec)
	if idx == nil {
		return nil
	}

	if !p.expectPeek(TokenRBracket) {
		return nil
	}

	if vr, ok := left.(*VariableReference); ok {
		vr.Indices = append(vr.Indices, idx)
		return vr
	}

	return &IndexExpr{
		Collection: left,
		Index:      idx,
	}
}

func (p *Parser) parseConditionalExpression(cond Expression) Expression {
	p.nextToken()
	p.nextToken()
	trueVal := p.parseExpression(lowestPrec)
	if trueVal == nil {
		return nil
	}

	if !p.expectPeek(TokenColon) {
		return nil
	}

	p.nextToken()
	falseVal := p.parseExpression(lowestPrec)
	if falseVal == nil {
		return nil
	}

	return &ConditionalExpr{
		Condition:  cond,
		TrueValue:  trueVal,
		FalseValue: falseVal,
	}
}

func ParseString(input string) (*Body, error) {
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	return parser.Parse()
}
