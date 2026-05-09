package parser

import (
	"fmt"
	"strconv"

	"graphql-parser/internal/graphql/ast"
	"graphql-parser/internal/graphql/lexer"
)

type Parser struct {
	l      *lexer.Lexer
	curTok lexer.Token
	peekTok lexer.Token
	errors []string
}

type ParserError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Message)
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curTok.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekTok.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekTok.Type)
	if p.peekTok.Type == lexer.TokenEOF {
		msg = "unexpected end of input"
	}
	p.errors = append(p.errors, fmt.Sprintf("line %d, column %d: %s", p.peekTok.Line, p.peekTok.Column, msg))
}

func (p *Parser) curError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d, column %d: %s", p.curTok.Line, p.curTok.Column, msg))
}

func (p *Parser) ParseDocument() (*ast.Document, error) {
	doc := &ast.Document{
		Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
	}

	if p.curTokenIs(lexer.TokenEOF) {
		return nil, &ParserError{Line: 1, Column: 1, Message: "empty query"}
	}

	for !p.curTokenIs(lexer.TokenEOF) {
		def := p.parseDefinition()
		if def != nil {
			doc.Definitions = append(doc.Definitions, def)
		}
		p.nextToken()
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}

	return doc, nil
}

func (p *Parser) parseDefinition() ast.Definition {
	switch p.curTok.Type {
	case lexer.TokenQuery, lexer.TokenMutation:
		return p.parseOperationDefinition()
	case lexer.TokenBraceL:
		return p.parseOperationDefinition()
	case lexer.TokenFragment:
		return p.parseFragmentDefinition()
	case lexer.TokenName:
		if p.peekTokenIs(lexer.TokenName) || p.peekTokenIs(lexer.TokenParenL) || p.peekTokenIs(lexer.TokenBraceL) || p.peekTokenIs(lexer.TokenColon) {
			return p.parseShortHandQuery()
		}
		p.curError(fmt.Sprintf("unsupported syntax: %s", p.curTok.Literal))
		return nil
	default:
		p.curError(fmt.Sprintf("unsupported syntax: %s", p.curTok.Literal))
		return nil
	}
}

func (p *Parser) parseShortHandQuery() ast.Definition {
	op := &ast.OperationDefinition{
		Position:      ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
		OperationType: ast.OperationTypeQuery,
	}
	op.SelectionSet = p.parseSelectionSet()
	return op
}

func (p *Parser) parseOperationDefinition() ast.Definition {
	op := &ast.OperationDefinition{
		Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
	}

	if p.curTokenIs(lexer.TokenBraceL) {
		op.OperationType = ast.OperationTypeQuery
		op.SelectionSet = p.parseSelectionSet()
		return op
	}

	switch p.curTok.Type {
	case lexer.TokenQuery:
		op.OperationType = ast.OperationTypeQuery
	case lexer.TokenMutation:
		op.OperationType = ast.OperationTypeMutation
	}

	if p.peekTokenIs(lexer.TokenName) {
		p.nextToken()
		op.Name = p.curTok.Literal
	}

	if p.peekTokenIs(lexer.TokenParenL) {
		p.nextToken()
		op.Variables = p.parseVariableDefinitions()
	}

	if p.expectPeek(lexer.TokenBraceL) {
		op.SelectionSet = p.parseSelectionSet()
	}

	return op
}

func (p *Parser) parseVariableDefinitions() []ast.VariableDefinition {
	var defs []ast.VariableDefinition

	if !p.expectPeek(lexer.TokenParenL) {
		return defs
	}

	for !p.curTokenIs(lexer.TokenParenR) && !p.curTokenIs(lexer.TokenEOF) {
		def := p.parseVariableDefinition()
		if def != nil {
			defs = append(defs, *def)
		}
		if !p.peekTokenIs(lexer.TokenParenR) {
			if !p.peekTokenIs(lexer.TokenComma) && !p.peekTokenIs(lexer.TokenParenR) {
				p.peekError(lexer.TokenComma)
				break
			}
			if p.peekTokenIs(lexer.TokenComma) {
				p.nextToken()
			}
		}
		p.nextToken()
	}

	return defs
}

func (p *Parser) parseVariableDefinition() *ast.VariableDefinition {
	def := &ast.VariableDefinition{
		Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
	}

	if !p.curTokenIs(lexer.TokenDollar) {
		p.curError("expected $ for variable")
		return nil
	}

	if !p.expectPeek(lexer.TokenName) {
		return nil
	}
	def.Variable = p.curTok.Literal

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	p.nextToken()
	def.Type = p.parseType()

	if p.peekTokenIs(lexer.TokenEquals) {
		p.nextToken()
		p.nextToken()
		def.DefaultValue = p.parseValue()
	}

	return def
}

func (p *Parser) parseType() ast.Type {
	pos := ast.Position{Line: p.curTok.Line, Column: p.curTok.Column}
	var t ast.Type

	if p.curTokenIs(lexer.TokenBracketL) {
		p.nextToken()
		inner := p.parseType()
		if inner == nil {
			return nil
		}
		if p.expectPeek(lexer.TokenBracketR) {
			t = &ast.ListType{Position: pos, OfType: inner}
		}
	} else if p.curTokenIs(lexer.TokenName) {
		t = &ast.NamedType{Position: pos, Name: p.curTok.Literal}
	} else {
		p.curError("expected type")
		return nil
	}

	if p.peekTokenIs(lexer.TokenBang) {
		p.nextToken()
		return &ast.NonNullType{Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column}, OfType: t}
	}

	return t
}

func (p *Parser) parseFragmentDefinition() ast.Definition {
	frag := &ast.FragmentDefinition{
		Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
	}

	if !p.expectPeek(lexer.TokenName) {
		return nil
	}
	frag.Name = p.curTok.Literal

	if !p.expectPeek(lexer.TokenOn) {
		return nil
	}

	if !p.expectPeek(lexer.TokenName) {
		return nil
	}
	frag.TypeCondition = p.curTok.Literal

	if p.expectPeek(lexer.TokenBraceL) {
		frag.SelectionSet = p.parseSelectionSet()
	}

	return frag
}

func (p *Parser) parseSelectionSet() ast.SelectionSet {
	var set ast.SelectionSet

	for !p.peekTokenIs(lexer.TokenBraceR) && !p.peekTokenIs(lexer.TokenEOF) {
		p.nextToken()
		sel := p.parseSelection()
		if sel != nil {
			set = append(set, sel)
		}
	}

	if !p.expectPeek(lexer.TokenBraceR) {
		return nil
	}

	return set
}

func (p *Parser) parseSelection() ast.Selection {
	if p.curTokenIs(lexer.TokenEllipsis) {
		return p.parseFragmentSpreadOrInline()
	}
	return p.parseField()
}

func (p *Parser) parseField() ast.Selection {
	field := &ast.Field{
		Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
	}

	field.Name = p.curTok.Literal

	if p.peekTokenIs(lexer.TokenColon) {
		field.Alias = field.Name
		p.nextToken()
		p.nextToken()
		field.Name = p.curTok.Literal
	}

	if p.peekTokenIs(lexer.TokenParenL) {
		p.nextToken()
		field.Arguments = p.parseArguments()
	}

	if p.peekTokenIs(lexer.TokenBraceL) {
		p.nextToken()
		field.SelectionSet = p.parseSelectionSet()
	}

	return field
}

func (p *Parser) parseFragmentSpreadOrInline() ast.Selection {
	pos := ast.Position{Line: p.curTok.Line, Column: p.curTok.Column}

	if p.peekTokenIs(lexer.TokenOn) {
		p.nextToken()
		p.nextToken()
		frag := &ast.InlineFragment{
			Position:      pos,
			TypeCondition: p.curTok.Literal,
		}
		if p.expectPeek(lexer.TokenBraceL) {
			frag.SelectionSet = p.parseSelectionSet()
		}
		return frag
	}

	if p.peekTokenIs(lexer.TokenName) {
		p.nextToken()
		return &ast.FragmentSpread{
			Position: pos,
			Name:     p.curTok.Literal,
		}
	}

	if p.peekTokenIs(lexer.TokenBraceL) {
		p.nextToken()
		frag := &ast.InlineFragment{Position: pos}
		frag.SelectionSet = p.parseSelectionSet()
		return frag
	}

	p.curError("invalid fragment spread or inline fragment")
	return nil
}

func (p *Parser) parseArguments() []ast.Argument {
	var args []ast.Argument

	if !p.expectPeek(lexer.TokenParenL) {
		return args
	}

	for !p.curTokenIs(lexer.TokenParenR) && !p.curTokenIs(lexer.TokenEOF) {
		arg := p.parseArgument()
		if arg != nil {
			args = append(args, *arg)
		}
		if !p.peekTokenIs(lexer.TokenParenR) {
			if p.peekTokenIs(lexer.TokenComma) {
				p.nextToken()
			}
		}
		p.nextToken()
	}

	return args
}

func (p *Parser) parseArgument() *ast.Argument {
	arg := &ast.Argument{
		Position: ast.Position{Line: p.curTok.Line, Column: p.curTok.Column},
		Name:     p.curTok.Literal,
	}

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	p.nextToken()
	arg.Value = p.parseValue()

	return arg
}

func (p *Parser) parseValue() ast.Value {
	switch p.curTok.Type {
	case lexer.TokenString:
		return p.curTok.Literal
	case lexer.TokenInt:
		if n, err := strconv.ParseInt(p.curTok.Literal, 10, 64); err == nil {
			return n
		}
		return p.curTok.Literal
	case lexer.TokenFloat:
		if n, err := strconv.ParseFloat(p.curTok.Literal, 64); err == nil {
			return n
		}
		return p.curTok.Literal
	case lexer.TokenTrue:
		return true
	case lexer.TokenFalse:
		return false
	case lexer.TokenNull:
		return nil
	case lexer.TokenDollar:
		if p.peekTokenIs(lexer.TokenName) {
			p.nextToken()
			return &ast.Variable{Name: p.curTok.Literal}
		}
		p.curError("expected variable name after $")
		return nil
	case lexer.TokenBracketL:
		return p.parseListValue()
	case lexer.TokenBraceL:
		return p.parseObjectValue()
	case lexer.TokenName:
		return p.curTok.Literal
	default:
		p.curError(fmt.Sprintf("unsupported value type: %s", p.curTok.Type))
		return nil
	}
}

func (p *Parser) parseListValue() []ast.Value {
	var values []ast.Value

	for !p.peekTokenIs(lexer.TokenBracketR) && !p.peekTokenIs(lexer.TokenEOF) {
		p.nextToken()
		v := p.parseValue()
		if v != nil {
			values = append(values, v)
		}
		if p.peekTokenIs(lexer.TokenComma) {
			p.nextToken()
		}
	}

	if !p.expectPeek(lexer.TokenBracketR) {
		return nil
	}

	return values
}

func (p *Parser) parseObjectValue() map[string]ast.Value {
	obj := make(map[string]ast.Value)

	for !p.peekTokenIs(lexer.TokenBraceR) && !p.peekTokenIs(lexer.TokenEOF) {
		p.nextToken()
		name := p.curTok.Literal
		if !p.expectPeek(lexer.TokenColon) {
			return nil
		}
		p.nextToken()
		obj[name] = p.parseValue()
		if p.peekTokenIs(lexer.TokenComma) {
			p.nextToken()
		}
	}

	if !p.expectPeek(lexer.TokenBraceR) {
		return nil
	}

	return obj
}

func Parse(input string) (*ast.Document, error) {
	l := lexer.New(input)
	p := New(l)
	return p.ParseDocument()
}
