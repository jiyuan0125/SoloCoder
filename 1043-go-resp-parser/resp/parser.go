package resp

import (
	"io"
)

type Parser struct {
	r *Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{r: NewReader(r)}
}

func (p *Parser) Parse() (Value, error) {
	return p.parseValue()
}

func (p *Parser) parseValue() (Value, error) {
	b, err := p.r.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch b {
	case '+':
		return p.parseSimpleString()
	case '-':
		return p.parseError()
	case ':':
		return p.parseInteger()
	case '$':
		return p.parseBulkString()
	case '*':
		return p.parseArray()
	default:
		return Value{}, p.r.errorf("unknown type prefix: %q", b)
	}
}

func (p *Parser) parseSimpleString() (Value, error) {
	s, err := p.r.ReadLineAsString()
	if err != nil {
		return Value{}, err
	}
	return NewSimpleString(s), nil
}

func (p *Parser) parseError() (Value, error) {
	s, err := p.r.ReadLineAsString()
	if err != nil {
		return Value{}, err
	}
	return NewError(s), nil
}

func (p *Parser) parseInteger() (Value, error) {
	n, err := p.r.ReadLineAsInteger()
	if err != nil {
		return Value{}, err
	}
	return NewInteger(n), nil
}

func (p *Parser) parseBulkString() (Value, error) {
	len, err := p.r.ReadLineAsInteger()
	if err != nil {
		return Value{}, err
	}

	if len == -1 {
		return NewNullBulkString(), nil
	}

	if len < 0 {
		return Value{}, p.r.errorf("invalid bulk string length: %d", len)
	}

	if len > MaxStringSize {
		return Value{}, p.r.errorf("bulk string too large: %d (max %d)", len, MaxStringSize)
	}

	if len == 0 {
		_, err = p.r.ReadN(2)
		if err != nil {
			return Value{}, err
		}
		return NewBulkString(""), nil
	}

	data, err := p.r.ReadN(int(len))
	if err != nil {
		return Value{}, err
	}

	_, err = p.r.ReadN(2)
	if err != nil {
		return Value{}, err
	}

	return NewBulkString(string(data)), nil
}

func (p *Parser) parseArray() (Value, error) {
	len, err := p.r.ReadLineAsInteger()
	if err != nil {
		return Value{}, err
	}

	if len == -1 {
		return NewNullArray(), nil
	}

	if len < 0 {
		return Value{}, p.r.errorf("invalid array length: %d", len)
	}

	if len > MaxArraySize {
		return Value{}, p.r.errorf("array too large: %d (max %d)", len, MaxArraySize)
	}

	arr := make([]Value, 0, len)
	for i := int64(0); i < len; i++ {
		v, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}
		arr = append(arr, v)
	}

	return NewArray(arr), nil
}

func (p *Parser) Pos() int64 {
	return p.r.Pos()
}
