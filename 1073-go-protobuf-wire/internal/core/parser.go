package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unicode/utf8"
)

var (
	ErrNestingDepthExceeded = errors.New("nesting depth exceeded")
	ErrLengthValueTooLarge  = errors.New("length value too large")
	ErrDataTruncated        = errors.New("data truncated")
	ErrUnknownWireType      = errors.New("unknown wire type")
)

type Parser struct {
	opts *ParseOptions
}

func NewParser(opts *ParseOptions) *Parser {
	if opts == nil {
		opts = DefaultParseOptions()
	}
	return &Parser{opts: opts}
}

func (p *Parser) Parse(data []byte, depth int) ([]*ParsedField, error) {
	if depth > p.opts.MaxNestingDepth {
		return nil, ErrNestingDepthExceeded
	}

	var fields []*ParsedField
	offset := 0

	for offset < len(data) {
		field, bytesRead, err := p.parseField(data[offset:], depth)
		if err != nil {
			return nil, fmt.Errorf("at offset %d: %w", offset, err)
		}
		fields = append(fields, field)
		offset += bytesRead
	}

	mergedFields := mergeRepeatedFields(fields)
	return mergedFields, nil
}

func mergeRepeatedFields(fields []*ParsedField) []*ParsedField {
	if len(fields) == 0 {
		return fields
	}

	fieldMap := make(map[int]*ParsedField)
	var result []*ParsedField

	for _, f := range fields {
		existing, found := fieldMap[f.FieldNumber]
		if !found {
			fieldMap[f.FieldNumber] = f
			result = append(result, f)
			continue
		}

		if !existing.Repeated {
			existing.Repeated = true
			existing.Values = []interface{}{existing.Value}
		}

		existing.Values = append(existing.Values, f.Value)
		existing.RawBytes = append(existing.RawBytes, f.RawBytes...)
		existing.RawByteSize += f.RawByteSize
	}

	return result
}

func (p *Parser) parseField(data []byte, depth int) (*ParsedField, int, error) {
	tag, tagBytes, err := ReadVarint(data)
	if err != nil {
		return nil, 0, err
	}

	wireType := WireType(tag & 0x7)
	fieldNumber := int(tag >> 3)

	valueBytes := 0
	var value interface{}
	var valueType string

	switch wireType {
	case WireVarint:
		v, vb, err := ReadVarint(data[tagBytes:])
		if err != nil {
			return nil, 0, err
		}
		valueBytes = vb
		value = v
		valueType = "uint64"

	case Wire64BitFixed:
		if len(data) < tagBytes+8 {
			return nil, 0, ErrDataTruncated
		}
		raw64 := binary.LittleEndian.Uint64(data[tagBytes : tagBytes+8])
		valueBytes = 8
		value = raw64
		valueType = "uint64/fixed64/double"

	case Wire32BitFixed:
		if len(data) < tagBytes+4 {
			return nil, 0, ErrDataTruncated
		}
		raw32 := binary.LittleEndian.Uint32(data[tagBytes : tagBytes+4])
		valueBytes = 4
		value = raw32
		valueType = "uint32/fixed32/float"

	case WireLengthDelimited:
		length, lb, err := ReadVarint(data[tagBytes:])
		if err != nil {
			return nil, 0, err
		}
		if length > uint64(p.opts.MaxLengthValue) {
			return nil, 0, ErrLengthValueTooLarge
		}
		lengthVal := int(length)
		totalLen := tagBytes + lb + lengthVal
		if totalLen > len(data) {
			return nil, 0, ErrDataTruncated
		}
		valueBytes = lb + lengthVal
		ldData := data[tagBytes+lb : totalLen]

		if isProbablyUTF8(ldData) {
			value = string(ldData)
			valueType = "string"
		} else {
			nestedFields, err := p.Parse(ldData, depth+1)
			if err == nil && len(nestedFields) > 0 {
				value = nestedFields
				valueType = "nested_message"
			} else {
				value = ldData
				valueType = "bytes"
			}
		}

	case WireStartGroup, WireEndGroup:
		return nil, 0, fmt.Errorf("group wire types (%d) are deprecated and not supported", wireType)

	default:
		return nil, 0, fmt.Errorf("%w: %d", ErrUnknownWireType, wireType)
	}

	totalBytes := tagBytes + valueBytes
	rawBytes := make([]byte, totalBytes)
	copy(rawBytes, data[:totalBytes])

	return &ParsedField{
		FieldNumber: fieldNumber,
		WireType:    wireType,
		RawBytes:    rawBytes,
		RawByteSize: totalBytes,
		Value:       value,
		ValueType:   valueType,
	}, totalBytes, nil
}

func isProbablyUTF8(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if !utf8.Valid(data) {
		return false
	}
	for _, b := range data {
		if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
			return false
		}
	}
	return true
}

func FormatFloat64(bits uint64) float64 {
	return math.Float64frombits(bits)
}

func FormatFloat32(bits uint32) float32 {
	return math.Float32frombits(bits)
}
