package core

import (
	"fmt"
	"strconv"
	"unicode"
)

type Parser struct {
	input  string
	pos    int
	line   int
	column int
}

func NewParser(input string) *Parser {
	return &Parser{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (p *Parser) Parse() (*MessageDef, error) {
	p.skipWhitespaceAndComments()
	return p.parseMessage()
}

func (p *Parser) parseMessage() (*MessageDef, error) {
	if err := p.expectKeyword("message"); err != nil {
		return nil, err
	}
	p.skipWhitespaceAndComments()

	name, err := p.parseMessageName()
	if err != nil {
		return nil, err
	}
	p.skipWhitespaceAndComments()

	if err := p.expectChar('{'); err != nil {
		return nil, err
	}
	p.skipWhitespaceAndComments()

	fields := make([]Field, 0)
	fieldNames := make(map[string]bool)

	for !p.atEnd() && p.currentChar() != '}' {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}

		if fieldNames[field.Name] {
			return nil, &SyntaxError{
				Line:    p.line,
				Column:  p.column,
				Message: fmt.Sprintf("duplicate field name: %s", field.Name),
			}
		}
		fieldNames[field.Name] = true

		fields = append(fields, *field)
		p.skipWhitespaceAndComments()
	}

	if err := p.expectChar('}'); err != nil {
		return nil, err
	}

	return &MessageDef{Name: name, Fields: fields}, nil
}

func (p *Parser) parseMessageName() (string, error) {
	start := p.pos
	for !p.atEnd() && (unicode.IsLetter(rune(p.currentChar())) || unicode.IsDigit(rune(p.currentChar()))) {
		p.advance()
	}
	if start == p.pos {
		return "", &SyntaxError{Line: p.line, Column: p.column, Message: "expected message name"}
	}

	name := p.input[start:p.pos]
	if !unicode.IsLetter(rune(name[0])) {
		return "", &SyntaxError{Line: p.line, Column: p.column, Message: "message name must start with a letter"}
	}
	return name, nil
}

func (p *Parser) parseField() (*Field, error) {
	fieldType, length, isChecksum, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()

	if p.atEnd() {
		return nil, &SyntaxError{Line: p.line, Column: p.column, Message: "expected field name"}
	}

	fieldName, err := p.parseFieldName()
	if err != nil {
		return nil, err
	}

	return &Field{
		Name:       fieldName,
		Type:       fieldType,
		Length:     length,
		IsChecksum: isChecksum,
	}, nil
}

func (p *Parser) parseFieldType() (FieldType, int, bool, error) {
	keywords := map[string]FieldType{
		"uint8":  TypeUint8,
		"uint16": TypeUint16,
		"uint32": TypeUint32,
		"uint64": TypeUint64,
		"int8":   TypeInt8,
		"int16":  TypeInt16,
		"int32":  TypeInt32,
		"int64":  TypeInt64,
		"var_string": TypeVarString,
	}

	start := p.pos
	for !p.atEnd() && (unicode.IsLetter(rune(p.currentChar())) || p.currentChar() == '_' || unicode.IsDigit(rune(p.currentChar()))) {
		p.advance()
	}

	keyword := p.input[start:p.pos]
	if keyword == "" {
		return 0, 0, false, &SyntaxError{Line: p.line, Column: p.column, Message: "expected field type"}
	}

	if keyword == "checksum" {
		return TypeChecksum, 0, true, nil
	}

	if fieldType, ok := keywords[keyword]; ok {
		return fieldType, 0, false, nil
	}

	if keyword == "fixed_string" || keyword == "bytes" {
		return p.parseFixedType(keyword)
	}

	return 0, 0, false, &SyntaxError{
		Line:    p.line,
		Column:  p.column,
		Message: fmt.Sprintf("unknown field type: %s", keyword),
	}
}

func (p *Parser) parseFixedType(keyword string) (FieldType, int, bool, error) {
	if err := p.expectChar('('); err != nil {
		return 0, 0, false, err
	}

	start := p.pos
	for !p.atEnd() && unicode.IsDigit(rune(p.currentChar())) {
		p.advance()
	}

	if start == p.pos {
		return 0, 0, false, &SyntaxError{
			Line:    p.line,
			Column:  p.column,
			Message: fmt.Sprintf("expected positive integer length for %s", keyword),
		}
	}

	length, err := strconv.Atoi(p.input[start:p.pos])
	if err != nil {
		return 0, 0, false, &SyntaxError{
			Line:    p.line,
			Column:  p.column,
			Message: fmt.Sprintf("invalid length: %v", err),
		}
	}

	if length <= 0 {
		return 0, 0, false, &SyntaxError{
			Line:    p.line,
			Column:  p.column,
			Message: fmt.Sprintf("length must be positive for %s", keyword),
		}
	}

	if err := p.expectChar(')'); err != nil {
		return 0, 0, false, err
	}

	var fieldType FieldType
	if keyword == "fixed_string" {
		fieldType = TypeFixedString
	} else {
		fieldType = TypeBytes
	}

	return fieldType, length, false, nil
}

func (p *Parser) parseFieldName() (string, error) {
	start := p.pos
	if !p.atEnd() && unicode.IsLetter(rune(p.currentChar())) {
		p.advance()
	} else {
		return "", &SyntaxError{
			Line:    p.line,
			Column:  p.column,
			Message: "field name must start with a letter",
		}
	}

	for !p.atEnd() && (unicode.IsLetter(rune(p.currentChar())) || unicode.IsDigit(rune(p.currentChar())) || p.currentChar() == '_') {
		p.advance()
	}

	return p.input[start:p.pos], nil
}

func (p *Parser) expectKeyword(keyword string) error {
	start := p.pos
	for !p.atEnd() && unicode.IsLetter(rune(p.currentChar())) {
		p.advance()
	}

	if p.input[start:p.pos] != keyword {
		return &SyntaxError{
			Line:    p.line,
			Column:  p.column,
			Message: fmt.Sprintf("expected '%s'", keyword),
		}
	}
	return nil
}

func (p *Parser) expectChar(c byte) error {
	if p.atEnd() || p.currentChar() != c {
		return &SyntaxError{
			Line:    p.line,
			Column:  p.column,
			Message: fmt.Sprintf("expected '%c'", c),
		}
	}
	p.advance()
	return nil
}

func (p *Parser) skipWhitespaceAndComments() {
	for {
		p.skipWhitespace()
		if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '/' {
			for !p.atEnd() && p.currentChar() != '\n' {
				p.advance()
			}
		} else {
			break
		}
	}
}

func (p *Parser) skipWhitespace() {
	for !p.atEnd() && (p.currentChar() == ' ' || p.currentChar() == '\t' || p.currentChar() == '\n' || p.currentChar() == '\r') {
		p.advance()
	}
}

func (p *Parser) atEnd() bool {
	return p.pos >= len(p.input)
}

func (p *Parser) currentChar() byte {
	if p.atEnd() {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) advance() {
	if p.atEnd() {
		return
	}
	if p.input[p.pos] == '\n' {
		p.line++
		p.column = 1
	} else {
		p.column++
	}
	p.pos++
}

func ValidateDSL(dsl string) error {
	parser := NewParser(dsl)
	def, err := parser.Parse()
	if err != nil {
		return err
	}

	for i, field := range def.Fields {
		if field.IsChecksum && i != len(def.Fields)-1 {
			return &SyntaxError{Message: "checksum field must be the last field"}
		}
	}

	return nil
}

func ParseDSL(dsl string) (*MessageDef, error) {
	parser := NewParser(dsl)
	def, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	for i, field := range def.Fields {
		if field.IsChecksum && i != len(def.Fields)-1 {
			return nil, &SyntaxError{Message: "checksum field must be the last field"}
		}
	}

	return def, nil
}
