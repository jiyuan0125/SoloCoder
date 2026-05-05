package template

import (
	"fmt"
	"strings"
)

type TokenType int

const (
	TokenText TokenType = iota
	TokenVariable
	TokenDefaultValue
	TokenIfStart
	TokenElse
	TokenEnd
	TokenEachStart
)

type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Column   int
	Raw      string
}

type Lexer struct {
	input    string
	pos      int
	line     int
	column   int
	startPos int
	startLine int
	startColumn int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) Lex() ([]Token, error) {
	var tokens []Token

	for l.pos < len(l.input) {
		if l.peek() == '{' && l.peekN(1) == '{' {
			if l.pos > l.startPos {
				tokens = append(tokens, l.createTextToken())
			}
			l.startPos = l.pos
			l.startLine = l.line
			l.startColumn = l.column
			
			token, err := l.lexTag()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			l.startPos = l.pos
		} else {
			l.advance()
		}
	}

	if l.pos > l.startPos {
		tokens = append(tokens, l.createTextToken())
	}

	return tokens, nil
}

func (l *Lexer) createTextToken() Token {
	return Token{
		Type:   TokenText,
		Value:  l.input[l.startPos:l.pos],
		Line:   l.startLine,
		Column: l.startColumn,
	}
}

func (l *Lexer) lexTag() (Token, error) {
	l.advance() // consume '{'
	l.advance() // consume '{'
	l.skipWhitespace()

	if l.match("if") {
		return l.lexIfStart()
	} else if l.match("else") {
		return l.lexElse()
	} else if l.match("end") {
		return l.lexEnd()
	} else if l.match("each") {
		return l.lexEachStart()
	} else {
		return l.lexVariable()
	}
}

func (l *Lexer) lexIfStart() (Token, error) {
	l.advanceN(2) // consume 'if'
	l.skipWhitespace()
	
	if l.peek() == '}' && l.peekN(1) == '}' {
		return Token{}, fmt.Errorf("语法错误: if标签缺少条件表达式 (第%d行)", l.line)
	}
	
	var value strings.Builder
	for l.pos < len(l.input) && !(l.peek() == '}' && l.peekN(1) == '}') {
		value.WriteByte(l.peek())
		l.advance()
	}
	
	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("语法错误: 未闭合的if标签 (第%d行)", l.startLine)
	}
	
	l.skipWhitespace()
	l.advance() // consume '}'
	l.advance() // consume '}'
	
	return Token{
		Type:   TokenIfStart,
		Value:  strings.TrimSpace(value.String()),
		Line:   l.startLine,
		Column: l.startColumn,
		Raw:    l.input[l.startPos:l.pos],
	}, nil
}

func (l *Lexer) lexElse() (Token, error) {
	l.advanceN(4) // consume 'else'
	l.skipWhitespace()
	
	if !(l.peek() == '}' && l.peekN(1) == '}') {
		return Token{}, fmt.Errorf("语法错误: else标签格式错误，应为{{else}} (第%d行)", l.startLine)
	}
	
	l.advance() // consume '}'
	l.advance() // consume '}'
	
	return Token{
		Type:   TokenElse,
		Value:  "",
		Line:   l.startLine,
		Column: l.startColumn,
		Raw:    l.input[l.startPos:l.pos],
	}, nil
}

func (l *Lexer) lexEnd() (Token, error) {
	l.advanceN(3) // consume 'end'
	l.skipWhitespace()
	
	if !(l.peek() == '}' && l.peekN(1) == '}') {
		return Token{}, fmt.Errorf("语法错误: end标签格式错误，应为{{end}} (第%d行)", l.startLine)
	}
	
	l.advance() // consume '}'
	l.advance() // consume '}'
	
	return Token{
		Type:   TokenEnd,
		Value:  "",
		Line:   l.startLine,
		Column: l.startColumn,
		Raw:    l.input[l.startPos:l.pos],
	}, nil
}

func (l *Lexer) lexEachStart() (Token, error) {
	l.advanceN(4) // consume 'each'
	l.skipWhitespace()
	
	if l.peek() == '}' && l.peekN(1) == '}' {
		return Token{}, fmt.Errorf("语法错误: each标签缺少数组表达式 (第%d行)", l.line)
	}
	
	var value strings.Builder
	for l.pos < len(l.input) && !(l.peek() == '}' && l.peekN(1) == '}') {
		value.WriteByte(l.peek())
		l.advance()
	}
	
	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("语法错误: 未闭合的each标签 (第%d行)", l.startLine)
	}
	
	l.skipWhitespace()
	l.advance() // consume '}'
	l.advance() // consume '}'
	
	return Token{
		Type:   TokenEachStart,
		Value:  strings.TrimSpace(value.String()),
		Line:   l.startLine,
		Column: l.startColumn,
		Raw:    l.input[l.startPos:l.pos],
	}, nil
}

func (l *Lexer) lexVariable() (Token, error) {
	var value strings.Builder
	var defaultValue string
	hasDefault := false
	
	for l.pos < len(l.input) {
		if l.peek() == '|' && !hasDefault {
			hasDefault = true
			l.advance()
			l.skipWhitespace()
			
			var defaultBuilder strings.Builder
			for l.pos < len(l.input) && !(l.peek() == '}' && l.peekN(1) == '}') {
				defaultBuilder.WriteByte(l.peek())
				l.advance()
			}
			defaultValue = strings.TrimSpace(defaultBuilder.String())
			break
		}
		
		if l.peek() == '}' && l.peekN(1) == '}' {
			break
		}
		
		value.WriteByte(l.peek())
		l.advance()
	}
	
	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("语法错误: 未闭合的变量标签 (第%d行)", l.startLine)
	}
	
	l.skipWhitespace()
	l.advance() // consume '}'
	l.advance() // consume '}'
	
	varName := strings.TrimSpace(value.String())
	
	if hasDefault {
		return Token{
			Type:   TokenDefaultValue,
			Value:  varName + "|" + defaultValue,
			Line:   l.startLine,
			Column: l.startColumn,
			Raw:    l.input[l.startPos:l.pos],
		}, nil
	}
	
	return Token{
		Type:   TokenVariable,
		Value:  varName,
		Line:   l.startLine,
		Column: l.startColumn,
		Raw:    l.input[l.startPos:l.pos],
	}, nil
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekN(n int) byte {
	if l.pos+n >= len(l.input) {
		return 0
	}
	return l.input[l.pos+n]
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.pos++
	}
}

func (l *Lexer) advanceN(n int) {
	for i := 0; i < n; i++ {
		l.advance()
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\n' || l.peek() == '\r') {
		l.advance()
	}
}

func (l *Lexer) match(s string) bool {
	if l.pos+len(s) > len(l.input) {
		return false
	}
	return l.input[l.pos:l.pos+len(s)] == s
}
