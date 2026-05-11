package core

import (
	"encoding/json"
	"fmt"
)

type FieldType int

const (
	TypeUint8 FieldType = iota
	TypeUint16
	TypeUint32
	TypeUint64
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeFixedString
	TypeVarString
	TypeBytes
	TypeChecksum
)

func (ft FieldType) String() string {
	switch ft {
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"
	case TypeFixedString:
		return "fixed_string"
	case TypeVarString:
		return "var_string"
	case TypeBytes:
		return "bytes"
	case TypeChecksum:
		return "checksum"
	default:
		return "unknown"
	}
}

type Field struct {
	Name    string
	Type    FieldType
	Length  int
	IsChecksum bool
}

type MessageDef struct {
	Name   string
	Fields []Field
}

type ParseResult struct {
	Data map[string]interface{}
}

func (pr *ParseResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(pr.Data)
}

type SyntaxError struct {
	Line    int
	Column  int
	Message string
}

func (e *SyntaxError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("syntax error at line %d: %s", e.Line, e.Message)
	}
	return "syntax error: " + e.Message
}

type ParseError struct {
	Field   string
	Message string
}

func (e *ParseError) Error() string {
	if e.Field != "" {
		return "parse error at field '" + e.Field + "': " + e.Message
	}
	return "parse error: " + e.Message
}

type SerializeError struct {
	Field   string
	Message string
}

func (e *SerializeError) Error() string {
	if e.Field != "" {
		return "serialize error at field '" + e.Field + "': " + e.Message
	}
	return "serialize error: " + e.Message
}

type ChecksumError struct {
	Expected uint8
	Actual   uint8
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch: expected %d, got %d", e.Expected, e.Actual)
}
