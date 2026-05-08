package template

import (
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenText tokenType = iota
	tokenLeftDelim
	tokenRightDelim
	tokenIdentifier
	tokenDot
	tokenIf
	tokenElse
	tokenEnd
	tokenRange
	tokenNumber
	tokenString
	tokenComma
	tokenPipe
	tokenEOF
	tokenError
)

type token struct {
	Type     tokenType
	Value    string
	Position Position
}

type lexerState int

const (
	stateText lexerState = iota
	stateAction
)

type lexer struct {
	input     string
	pos       int
	line      int
	column    int
	startPos  int
	startLine int
	startCol  int
	delims    Delims
	state     lexerState
}

func newLexer(input string, delims Delims) *lexer {
	return &lexer{
		input:     input,
		pos:       0,
		line:      1,
		column:    1,
		startPos:  0,
		startLine: 1,
		startCol:  1,
		delims:    delims,
		state:     stateText,
	}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *lexer) peekN(n int) string {
	if l.pos+n > len(l.input) {
		return l.input[l.pos:]
	}
	return l.input[l.pos : l.pos+n]
}

func (l *lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r := rune(l.input[l.pos])
	l.pos++
	if r == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r
}

func (l *lexer) position() Position {
	return Position{
		Line:   l.startLine,
		Column: l.startCol,
	}
}

func (l *lexer) currentPosition() Position {
	return Position{
		Line:   l.line,
		Column: l.column,
	}
}

func (l *lexer) emit(t tokenType, value string) token {
	return token{
		Type:     t,
		Value:    value,
		Position: l.position(),
	}
}

func (l *lexer) skipWhitespace() {
	for {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *lexer) readIdentifier() string {
	var sb strings.Builder
	for {
		r := l.peek()
		if unicode.IsLetter(r) || r == '_' || unicode.IsDigit(r) {
			sb.WriteRune(l.advance())
		} else {
			break
		}
	}
	return sb.String()
}

func (l *lexer) lex() []token {
	var tokens []token
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == tokenEOF || tok.Type == tokenError {
			break
		}
	}
	return tokens
}

func (l *lexer) nextToken() token {
	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.column

	if l.state == stateText {
		return l.lexTextState()
	}
	return l.lexActionState()
}

func (l *lexer) lexTextState() token {
	if l.pos >= len(l.input) {
		return l.emit(tokenEOF, "")
	}

	prefix := l.peekN(len(l.delims.Left))
	if prefix == l.delims.Left {
		for i := 0; i < len(l.delims.Left); i++ {
			l.advance()
		}
		l.state = stateAction
		return l.lexActionState()
	}

	var sb strings.Builder
	for {
		if l.pos >= len(l.input) {
			break
		}
		prefix := l.peekN(len(l.delims.Left))
		if prefix == l.delims.Left {
			break
		}
		r := l.advance()
		sb.WriteRune(r)
	}
	return l.emit(tokenText, sb.String())
}

func (l *lexer) lexActionState() token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return l.emit(tokenError, "unterminated action")
	}

	if l.peekN(len(l.delims.Right)) == l.delims.Right {
		for i := 0; i < len(l.delims.Right); i++ {
			l.advance()
		}
		l.state = stateText
		return l.emit(tokenRightDelim, l.delims.Right)
	}

	r := l.peek()
	switch {
	case r == '.':
		return l.lexDotOrVariable()
	case unicode.IsLetter(r) || r == '_':
		ident := l.readIdentifier()
		switch ident {
		case "if":
			return l.emit(tokenIf, "if")
		case "else":
			l.skipWhitespace()
			if l.peekN(len(l.delims.Right)) == l.delims.Right {
				for i := 0; i < len(l.delims.Right); i++ {
					l.advance()
				}
				l.state = stateText
			}
			return l.emit(tokenElse, "else")
		case "end":
			l.skipWhitespace()
			if l.peekN(len(l.delims.Right)) == l.delims.Right {
				for i := 0; i < len(l.delims.Right); i++ {
					l.advance()
				}
				l.state = stateText
			}
			return l.emit(tokenEnd, "end")
		case "range":
			return l.emit(tokenRange, "range")
		default:
			return l.emit(tokenIdentifier, ident)
		}
	case r == '"' || r == '\'':
		quote := l.advance()
		var sb strings.Builder
		for {
			r2 := l.peek()
			if r2 == 0 {
				return l.emit(tokenError, "unterminated string")
			}
			if r2 == quote {
				l.advance()
				break
			}
			if r2 == '\\' {
				l.advance()
				escaped := l.advance()
				switch escaped {
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case 'r':
					sb.WriteRune('\r')
				case '\\':
					sb.WriteRune('\\')
				case '"':
					sb.WriteRune('"')
				case '\'':
					sb.WriteRune('\'')
				default:
					sb.WriteRune('\\')
					sb.WriteRune(escaped)
				}
				continue
			}
			sb.WriteRune(l.advance())
		}
		return l.emit(tokenString, sb.String())
	case unicode.IsDigit(r):
		var sb strings.Builder
		for {
			r2 := l.peek()
			if unicode.IsDigit(r2) || r2 == '.' {
				sb.WriteRune(l.advance())
			} else {
				break
			}
		}
		return l.emit(tokenNumber, sb.String())
	case r == '|':
		l.advance()
		return l.emit(tokenPipe, "|")
	case r == ',':
		l.advance()
		return l.emit(tokenComma, ",")
	default:
		return l.emit(tokenError, "unexpected character in action")
	}
}

func (l *lexer) lexDotOrVariable() token {
	l.advance()
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return l.emit(tokenDot, ".")
	}

	r := l.peek()
	if r == 0 {
		return l.emit(tokenDot, ".")
	}

	if l.peekN(len(l.delims.Right)) == l.delims.Right {
		return l.emit(tokenDot, ".")
	}

	if !(unicode.IsLetter(r) || r == '_' || unicode.IsDigit(r)) {
		return l.emit(tokenDot, ".")
	}

	var parts []string
	parts = append(parts, "")

	for {
		ident := l.readIdentifier()
		parts[len(parts)-1] = ident

		l.skipWhitespace()
		if l.peek() != '.' {
			break
		}

		l.advance()
		l.skipWhitespace()
		parts = append(parts, "")
	}

	value := strings.Join(parts, ".")
	return l.emit(tokenIdentifier, value)
}

type parser struct {
	engine *Engine
	lexer  *lexer
	tokens []token
	pos    int
}

func newParser(engine *Engine, input string) *parser {
	l := newLexer(input, engine.config.Delims)
	return &parser{
		engine: engine,
		lexer:  l,
		tokens: l.lex(),
		pos:    0,
	}
}

func (p *parser) current() token {
	if p.pos >= len(p.tokens) {
		return token{Type: tokenEOF, Position: p.lexer.currentPosition()}
	}
	return p.tokens[p.pos]
}

func (p *parser) peek() token {
	if p.pos+1 >= len(p.tokens) {
		return token{Type: tokenEOF, Position: p.lexer.currentPosition()}
	}
	return p.tokens[p.pos+1]
}

func (p *parser) advance() token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *parser) expect(t tokenType) (token, error) {
	tok := p.current()
	if tok.Type != t {
		return tok, NewErrorf(tok.Position, "expected %v, got %v (%q)", t, tok.Type, tok.Value)
	}
	p.advance()
	return tok, nil
}

func (p *parser) parse() (*ParsedTemplate, error) {
	children, err := p.parseStatements(nil)
	if err != nil {
		return nil, err
	}

	if p.current().Type == tokenError {
		return nil, NewError(p.current().Position, p.current().Value)
	}
	if p.current().Type != tokenEOF {
		return nil, NewErrorf(p.current().Position, "unexpected token after parse complete: %v", p.current().Type)
	}

	root := &RootNode{Children: children}
	return newParsedTemplate(p.engine, root, p.lexer.input), nil
}

func (p *parser) parseStatements(stopTypes []tokenType) ([]Node, error) {
	var nodes []Node
	for {
		tok := p.current()

		shouldStop := false
		for _, st := range stopTypes {
			if tok.Type == st {
				shouldStop = true
				break
			}
		}
		if shouldStop || tok.Type == tokenEOF || tok.Type == tokenError {
			break
		}

		node, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (p *parser) parseStatement() (Node, error) {
	tok := p.current()
	switch tok.Type {
	case tokenText:
		p.advance()
		return &TextNode{
			Text:     tok.Value,
			position: tok.Position,
		}, nil
	case tokenIf:
		return p.parseIf()
	case tokenRange:
		return p.parseRange()
	case tokenLeftDelim:
		p.advance()
		return p.parseExpression()
	default:
		return p.parseExpression()
	}
}

func (p *parser) parseIf() (Node, error) {
	ifTok := p.advance()

	cond, err := p.parseExpressionNode()
	if err != nil {
		return nil, err
	}

	tok := p.current()
	if tok.Type == tokenPipe {
		p.advance()
		cond, err = p.parsePipe(cond)
		if err != nil {
			return nil, err
		}
	} else if tok.Type == tokenRightDelim {
		p.advance()
	}

	body, err := p.parseStatements([]tokenType{tokenElse, tokenEnd, tokenEOF})
	if err != nil {
		return nil, err
	}

	var elseBody []Node
	tok = p.current()
	if tok.Type == tokenElse {
		p.advance()
		elseBody, err = p.parseStatements([]tokenType{tokenEnd, tokenEOF})
		if err != nil {
			return nil, err
		}
		tok = p.current()
	}

	if tok.Type != tokenEnd {
		return nil, NewErrorf(ifTok.Position, "if statement not closed (missing {{end}})")
	}
	p.advance()

	return &IfNode{
		Condition: cond,
		Body:      body,
		Else:      elseBody,
		position:  ifTok.Position,
	}, nil
}

func (p *parser) parseRange() (Node, error) {
	rangeTok := p.advance()

	varNode, err := p.parseExpressionNode()
	if err != nil {
		return nil, err
	}

	tok := p.current()
	if tok.Type == tokenPipe {
		p.advance()
		varNode, err = p.parsePipe(varNode)
		if err != nil {
			return nil, err
		}
	} else if tok.Type == tokenRightDelim {
		p.advance()
	}

	body, err := p.parseStatements([]tokenType{tokenEnd, tokenEOF})
	if err != nil {
		return nil, err
	}

	tok = p.current()
	if tok.Type != tokenEnd {
		return nil, NewErrorf(rangeTok.Position, "range statement not closed (missing {{end}})")
	}
	p.advance()

	return &RangeNode{
		Variable: varNode,
		Body:     body,
		position: rangeTok.Position,
	}, nil
}

func (p *parser) parseExpression() (Node, error) {
	node, err := p.parseExpressionNode()
	if err != nil {
		return nil, err
	}

	tok := p.current()
	if tok.Type == tokenPipe {
		p.advance()
		return p.parsePipe(node)
	}

	if tok.Type == tokenRightDelim {
		p.advance()
	}

	return node, nil
}

func (p *parser) parseExpressionNode() (Node, error) {
	tok := p.current()

	switch tok.Type {
	case tokenDot:
		p.advance()
		return &VariableNode{
			Path:     []string{},
			position: tok.Position,
		}, nil
	case tokenIdentifier:
		p.advance()

		if p.engine.HasFunction(tok.Value) {
			nextTok := p.current()
			if nextTok.Type == tokenIdentifier || nextTok.Type == tokenString || 
			   nextTok.Type == tokenNumber || nextTok.Type == tokenDot {
				return p.parseFunctionCall(tok.Value, tok.Position)
			}
		}

		if strings.Contains(tok.Value, ".") {
			return &VariableNode{
				Path:     strings.Split(tok.Value, "."),
				position: tok.Position,
			}, nil
		}

		return &VariableNode{
			Path:     []string{tok.Value},
			position: tok.Position,
		}, nil
	case tokenString:
		p.advance()
		return &TextNode{
			Text:     tok.Value,
			position: tok.Position,
		}, nil
	case tokenNumber:
		p.advance()
		return &TextNode{
			Text:     tok.Value,
			position: tok.Position,
		}, nil
	default:
		return nil, NewErrorf(tok.Position, "unexpected token in expression: %v (%q)", tok.Type, tok.Value)
	}
}

func (p *parser) parseFunctionCall(name string, pos Position) (Node, error) {
	var args []Node

	for {
		tok := p.current()
		if tok.Type == tokenRightDelim || tok.Type == tokenPipe || tok.Type == tokenEOF || 
		   tok.Type == tokenElse || tok.Type == tokenEnd {
			break
		}

		if tok.Type == tokenComma {
			p.advance()
			continue
		}

		arg, err := p.parseExpressionNode()
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
	}

	tok := p.current()
	if tok.Type == tokenPipe {
		p.advance()
		return p.parsePipe(&FunctionNode{
			Name:     name,
			Args:     args,
			position: pos,
		})
	}

	if tok.Type == tokenRightDelim {
		p.advance()
	}

	return &FunctionNode{
		Name:     name,
		Args:     args,
		position: pos,
	}, nil
}

func (p *parser) parsePipe(left Node) (Node, error) {
	tok := p.current()
	if tok.Type != tokenIdentifier {
		return nil, NewErrorf(tok.Position, "expected function name after pipe")
	}
	p.advance()

	if !p.engine.HasFunction(tok.Value) {
		return nil, NewErrorf(tok.Position, "undefined function: %s", tok.Value)
	}

	var args []Node
	args = append(args, left)

	for {
		next := p.current()
		if next.Type == tokenRightDelim || next.Type == tokenPipe || next.Type == tokenEOF ||
		   next.Type == tokenElse || next.Type == tokenEnd {
			break
		}
		if next.Type == tokenComma {
			p.advance()
			continue
		}

		arg, err := p.parseExpressionNode()
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
	}

	tok2 := p.current()
	if tok2.Type == tokenPipe {
		p.advance()
		return p.parsePipe(&FunctionNode{
			Name:     tok.Value,
			Args:     args,
			position: tok.Position,
		})
	}

	if tok2.Type == tokenRightDelim {
		p.advance()
	}

	return &FunctionNode{
		Name:     tok.Value,
		Args:     args,
		position: tok.Position,
	}, nil
}
