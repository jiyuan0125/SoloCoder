package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenOperator
	TokenAND
	TokenOR
	TokenLParen
	TokenRParen
)

type Token struct {
	Type    TokenType
	Value   string
	Pos     int
	Line    int
	Col     int
}

type ParseError struct {
	Message string
	Pos     int
	Line    int
	Col     int
	Input   string
}

func (e *ParseError) Error() string {
	if e.Input == "" {
		return fmt.Sprintf("语法错误: %s (位置: 行%d, 列%d)", e.Message, e.Line, e.Col)
	}

	lines := strings.Split(e.Input, "\n")
	lineIdx := e.Line - 1
	if lineIdx < 0 || lineIdx >= len(lines) {
		return fmt.Sprintf("语法错误: %s (位置: %d)", e.Message, e.Pos)
	}

	line := lines[lineIdx]
	pointer := strings.Repeat(" ", e.Col-1) + "^"

	return fmt.Sprintf("语法错误: %s\n位置: 行%d, 列%d\n%s\n%s",
		e.Message, e.Line, e.Col, line, pointer)
}

type Parser struct {
	input   string
	pos     int
	line    int
	col     int
	tokens  []Token
	curIdx  int
}

func NewParser(input string) *Parser {
	return &Parser{
		input:  input,
		line:   1,
		col:    1,
		curIdx: 0,
	}
}

func (p *Parser) Parse() (string, error) {
	tokens, err := p.tokenize()
	if err != nil {
		return "", err
	}
	p.tokens = tokens
	p.curIdx = 0

	if len(tokens) == 0 || (len(tokens) == 1 && tokens[0].Type == TokenEOF) {
		return "", nil
	}

	result, err := p.parseExpression()
	if err != nil {
		return "", err
	}

	if p.curIdx < len(p.tokens) && p.tokens[p.curIdx].Type != TokenEOF {
		tok := p.tokens[p.curIdx]
		return "", &ParseError{
			Message: "意外的 token: " + tok.Value,
			Pos:     tok.Pos,
			Line:    tok.Line,
			Col:     tok.Col,
			Input:   p.input,
		}
	}

	return result, nil
}

func (p *Parser) tokenize() ([]Token, error) {
	var tokens []Token
	n := len(p.input)

	for p.pos < n {
		ch := p.input[p.pos]

		if unicode.IsSpace(rune(ch)) {
			p.consumeSpace()
			continue
		}

		if ch == '-' && p.pos+1 < n && p.input[p.pos+1] == '-' {
			return nil, &ParseError{
				Message: "不允许使用 SQL 注释",
				Pos:     p.pos,
				Line:    p.line,
				Col:     p.col,
				Input:   p.input,
			}
		}

		if ch == ';' {
			return nil, &ParseError{
				Message: "不允许使用分号",
				Pos:     p.pos,
				Line:    p.line,
				Col:     p.col,
				Input:   p.input,
			}
		}

		if ch == '(' {
			tokens = append(tokens, Token{Type: TokenLParen, Value: "(", Pos: p.pos, Line: p.line, Col: p.col})
			p.advance()
			continue
		}
		if ch == ')' {
			tokens = append(tokens, Token{Type: TokenRParen, Value: ")", Pos: p.pos, Line: p.line, Col: p.col})
			p.advance()
			continue
		}

		if op := p.readOperator(); op != "" {
			tokens = append(tokens, Token{Type: TokenOperator, Value: op, Pos: p.pos - len(op), Line: p.line, Col: p.col - len(op), Input: p.input})
			continue
		}

		if ch == '"' || ch == '\'' {
			tok, err := p.readString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		if unicode.IsDigit(rune(ch)) || (ch == '-' && p.pos+1 < n && unicode.IsDigit(rune(p.input[p.pos+1]))) {
			tok := p.readNumber()
			tokens = append(tokens, tok)
			continue
		}

		if unicode.IsLetter(rune(ch)) || ch == '_' || ch == '.' {
			tok := p.readIdentifier()
			tokens = append(tokens, tok)
			continue
		}

		return nil, &ParseError{
			Message: "无法识别的字符: " + string(ch),
			Pos:     p.pos,
			Line:    p.line,
			Col:     p.col,
			Input:   p.input,
		}
	}

	tokens = append(tokens, Token{Type: TokenEOF, Value: "", Pos: p.pos, Line: p.line, Col: p.col})
	return tokens, nil
}

func (p *Parser) readOperator() string {
	n := len(p.input)
	pos := p.pos

	if pos+2 <= n {
		two := p.input[pos : pos+2]
		if two == ">=" || two == "<=" || two == "<>" || two == "!=" || two == "==" {
			for i := 0; i < 2; i++ {
				p.advance()
			}
			if two == "!=" || two == "==" {
				return "="
			}
			return two
		}
	}

	if pos < n {
		one := p.input[pos]
		if one == '=' || one == '>' || one == '<' {
			p.advance()
			return string(one)
		}
	}

	return ""
}

func (p *Parser) readString() (Token, error) {
	startPos := p.pos
	startLine := p.line
	startCol := p.col
	quote := p.input[p.pos]
	p.advance()

	var value strings.Builder
	value.WriteByte(quote)

	n := len(p.input)
	for p.pos < n {
		ch := p.input[p.pos]
		value.WriteByte(ch)
		p.advance()

		if ch == quote {
			if p.pos < n && p.input[p.pos] == quote {
				value.WriteByte(p.input[p.pos])
				p.advance()
				continue
			}
			return Token{
				Type:  TokenString,
				Value: value.String(),
				Pos:   startPos,
				Line:  startLine,
				Col:   startCol,
			}, nil
		}
	}

	return Token{}, &ParseError{
		Message: "未闭合的字符串",
		Pos:     startPos,
		Line:    startLine,
		Col:     startCol,
		Input:   p.input,
	}
}

func (p *Parser) readNumber() Token {
	startPos := p.pos
	startLine := p.line
	startCol := p.col

	n := len(p.input)
	hasDot := false
	for p.pos < n {
		ch := p.input[p.pos]
		if unicode.IsDigit(rune(ch)) {
			p.advance()
		} else if ch == '.' && !hasDot {
			hasDot = true
			p.advance()
		} else if ch == '-' && p.pos == startPos {
			p.advance()
		} else {
			break
		}
	}

	return Token{
		Type:  TokenNumber,
		Value: p.input[startPos:p.pos],
		Pos:   startPos,
		Line:  startLine,
		Col:   startCol,
	}
}

func (p *Parser) readIdentifier() Token {
	startPos := p.pos
	startLine := p.line
	startCol := p.col

	n := len(p.input)
	for p.pos < n {
		ch := p.input[p.pos]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' || ch == '.' {
			p.advance()
		} else {
			break
		}
	}

	value := p.input[startPos:p.pos]
	upper := strings.ToUpper(value)

	var tokenType TokenType
	if upper == "AND" {
		tokenType = TokenAND
	} else if upper == "OR" {
		tokenType = TokenOR
	} else {
		tokenType = TokenIdentifier
	}

	return Token{
		Type:  tokenType,
		Value: value,
		Pos:   startPos,
		Line:  startLine,
		Col:   startCol,
	}
}

func (p *Parser) consumeSpace() {
	n := len(p.input)
	for p.pos < n {
		ch := p.input[p.pos]
		if ch == '\n' {
			p.line++
			p.col = 1
			p.pos++
		} else if unicode.IsSpace(rune(ch)) {
			p.col++
			p.pos++
		} else {
			break
		}
	}
}

func (p *Parser) advance() {
	if p.pos < len(p.input) {
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) parseExpression() (string, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (string, error) {
	left, err := p.parseAnd()
	if err != nil {
		return "", err
	}

	for p.curIdx < len(p.tokens) && p.tokens[p.curIdx].Type == TokenOR {
		orTok := p.tokens[p.curIdx]
		p.curIdx++

		if p.curIdx >= len(p.tokens) || p.tokens[p.curIdx].Type == TokenEOF {
			return "", &ParseError{
				Message: "OR 后缺少表达式",
				Pos:     orTok.Pos + len(orTok.Value),
				Line:    orTok.Line,
				Col:     orTok.Col + len(orTok.Value),
				Input:   p.input,
			}
		}

		right, err := p.parseAnd()
		if err != nil {
			return "", err
		}

		left = left + " OR " + right
	}

	return left, nil
}

func (p *Parser) parseAnd() (string, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return "", err
	}

	for p.curIdx < len(p.tokens) && p.tokens[p.curIdx].Type == TokenAND {
		andTok := p.tokens[p.curIdx]
		p.curIdx++

		if p.curIdx >= len(p.tokens) || p.tokens[p.curIdx].Type == TokenEOF {
			return "", &ParseError{
				Message: "AND 后缺少表达式",
				Pos:     andTok.Pos + len(andTok.Value),
				Line:    andTok.Line,
				Col:     andTok.Col + len(andTok.Value),
				Input:   p.input,
			}
		}

		right, err := p.parsePrimary()
		if err != nil {
			return "", err
		}

		left = left + " AND " + right
	}

	return left, nil
}

func (p *Parser) parsePrimary() (string, error) {
	if p.curIdx >= len(p.tokens) {
		return "", &ParseError{
			Message: "意外的结尾，期望表达式",
			Pos:     p.pos,
			Line:    p.line,
			Col:     p.col,
			Input:   p.input,
		}
	}

	tok := p.tokens[p.curIdx]

	if tok.Type == TokenLParen {
		p.curIdx++
		expr, err := p.parseExpression()
		if err != nil {
			return "", err
		}

		if p.curIdx >= len(p.tokens) || p.tokens[p.curIdx].Type != TokenRParen {
			return "", &ParseError{
				Message: "缺少右括号",
				Pos:     tok.Pos,
				Line:    tok.Line,
				Col:     tok.Col,
				Input:   p.input,
			}
		}
		p.curIdx++
		return "(" + expr + ")", nil
	}

	if tok.Type == TokenIdentifier {
		p.curIdx++

		if p.curIdx >= len(p.tokens) || p.tokens[p.curIdx].Type != TokenOperator {
			if p.curIdx >= len(p.tokens) {
				return "", &ParseError{
					Message: "字段后缺少比较运算符，期望 =、>、<、>=、<=、<> 等",
					Pos:     tok.Pos + len(tok.Value),
					Line:    tok.Line,
					Col:     tok.Col + len(tok.Value),
					Input:   p.input,
				}
			}
			return "", &ParseError{
				Message: "字段后缺少比较运算符，期望 =、>、<、>=、<=、<> 等",
				Pos:     tok.Pos + len(tok.Value),
				Line:    tok.Line,
				Col:     tok.Col + len(tok.Value),
				Input:   p.input,
			}
		}

		opTok := p.tokens[p.curIdx]
		p.curIdx++

		if p.curIdx >= len(p.tokens) {
			return "", &ParseError{
				Message: "运算符后缺少值",
				Pos:     opTok.Pos + len(opTok.Value),
				Line:    opTok.Line,
				Col:     opTok.Col + len(opTok.Value),
				Input:   p.input,
			}
		}

		valTok := p.tokens[p.curIdx]
		if valTok.Type != TokenNumber && valTok.Type != TokenString && valTok.Type != TokenIdentifier {
			return "", &ParseError{
				Message: "期望数字、字符串或标识符",
				Pos:     valTok.Pos,
				Line:    valTok.Line,
				Col:     valTok.Col,
				Input:   p.input,
			}
		}
		p.curIdx++

		return tok.Value + " " + opTok.Value + " " + valTok.Value, nil
	}

	if tok.Type == TokenEOF {
		return "", &ParseError{
			Message: "意外的结尾，期望表达式",
			Pos:     tok.Pos,
			Line:    tok.Line,
			Col:     tok.Col,
			Input:   p.input,
		}
	}

	return "", &ParseError{
		Message: "意外的 token: " + tok.Value,
		Pos:     tok.Pos,
		Line:    tok.Line,
		Col:     tok.Col,
		Input:   p.input,
	}
}

func ParseConditions(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}

	parser := NewParser(input)
	return parser.Parse()
}
