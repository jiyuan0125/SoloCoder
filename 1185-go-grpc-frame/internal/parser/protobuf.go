package parser

import (
	"encoding/binary"
	"errors"
	"math"
)

type ProtobufWireType uint8

const (
	WireTypeVarint          ProtobufWireType = 0
	WireTypeFixed64         ProtobufWireType = 1
	WireTypeLengthDelimited ProtobufWireType = 2
	WireTypeStartGroup      ProtobufWireType = 3
	WireTypeEndGroup        ProtobufWireType = 4
	WireTypeFixed32         ProtobufWireType = 5
)

func (wt ProtobufWireType) String() string {
	switch wt {
	case WireTypeVarint:
		return "varint"
	case WireTypeFixed64:
		return "fixed64"
	case WireTypeLengthDelimited:
		return "length-delimited"
	case WireTypeStartGroup:
		return "start-group"
	case WireTypeEndGroup:
		return "end-group"
	case WireTypeFixed32:
		return "fixed32"
	default:
		return "unknown"
	}
}

type ProtobufField struct {
	FieldNumber uint64
	WireType    ProtobufWireType
	Value       interface{}
	RawBytes    []byte
}

func decodeVarint(data []byte, offset int) (uint64, int, error) {
	if offset >= len(data) {
		return 0, offset, errors.New("insufficient data for varint")
	}

	var result uint64
	var shift uint
	var consumed int

	for shift < 64 {
		if offset >= len(data) {
			return 0, offset, errors.New("incomplete varint encoding")
		}
		b := data[offset]
		offset++
		consumed++

		result |= uint64(b&0x7F) << shift
		if (b & 0x80) == 0 {
			return result, consumed, nil
		}
		shift += 7
	}

	return 0, consumed, errors.New("varint too long")
}

func decodeZigZag32(n uint64) int32 {
	return int32((n >> 1) ^ (-(n & 1)))
}

func decodeZigZag64(n uint64) int64 {
	return int64((n >> 1) ^ (-(n & 1)))
}

func ParseProtobuf(data []byte) ([]ProtobufField, error) {
	var fields []ProtobufField
	offset := 0

	for offset < len(data) {
		tag, consumed, err := decodeVarint(data, offset)
		if err != nil {
			return fields, err
		}
		offset += consumed

		fieldNumber := tag >> 3
		wireType := ProtobufWireType(tag & 0x7)

		field := ProtobufField{
			FieldNumber: fieldNumber,
			WireType:    wireType,
		}

		switch wireType {
		case WireTypeVarint:
			value, n, err := decodeVarint(data, offset)
			if err != nil {
				return fields, err
			}
			field.Value = map[string]interface{}{
				"uint64":     value,
				"int64":      int64(value),
				"zigzag32":   decodeZigZag32(value),
				"zigzag64":   decodeZigZag64(value),
			}
			field.RawBytes = data[offset : offset+n]
			offset += n

		case WireTypeFixed64:
			if offset+8 > len(data) {
				return fields, errors.New("insufficient data for fixed64")
			}
			value := binary.LittleEndian.Uint64(data[offset : offset+8])
			field.Value = map[string]interface{}{
				"uint64":  value,
				"int64":   int64(value),
				"float64": math.Float64frombits(value),
			}
			field.RawBytes = data[offset : offset+8]
			offset += 8

		case WireTypeFixed32:
			if offset+4 > len(data) {
				return fields, errors.New("insufficient data for fixed32")
			}
			value := binary.LittleEndian.Uint32(data[offset : offset+4])
			field.Value = map[string]interface{}{
				"uint32":  value,
				"int32":   int32(value),
				"float32": math.Float32frombits(value),
			}
			field.RawBytes = data[offset : offset+4]
			offset += 4

		case WireTypeLengthDelimited:
			length, n, err := decodeVarint(data, offset)
			if err != nil {
				return fields, err
			}
			offset += n

			if uint64(len(data)-offset) < length {
				return fields, errors.New("insufficient data for length-delimited")
			}
			valueBytes := data[offset : offset+int(length)]
			field.Value = map[string]interface{}{
				"bytes":  valueBytes,
				"string": string(valueBytes),
			}
			field.RawBytes = valueBytes
			offset += int(length)

		case WireTypeStartGroup, WireTypeEndGroup:
			return fields, errors.New("group types are not supported")

		default:
			return fields, errors.New("unknown wire type")
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func EncodeVarint(value uint64) []byte {
	var buf []byte
	for {
		b := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if value == 0 {
			break
		}
	}
	return buf
}

func EncodeZigZag32(n int32) uint64 {
	return uint64((n << 1) ^ (n >> 31))
}

func EncodeZigZag64(n int64) uint64 {
	return uint64((n << 1) ^ (n >> 63))
}

func EncodeProtobufField(field ProtobufField) ([]byte, error) {
	var buf []byte
	tag := (field.FieldNumber << 3) | uint64(field.WireType)
	buf = append(buf, EncodeVarint(tag)...)

	switch field.WireType {
	case WireTypeVarint:
		if m, ok := field.Value.(map[string]interface{}); ok {
			if v, exists := m["uint64"]; exists {
				buf = append(buf, EncodeVarint(v.(uint64))...)
			} else {
				return nil, errors.New("varint field requires uint64 value")
			}
		} else if v, ok := field.Value.(uint64); ok {
			buf = append(buf, EncodeVarint(v)...)
		} else {
			return nil, errors.New("invalid varint value type")
		}

	case WireTypeFixed64:
		var value uint64
		if m, ok := field.Value.(map[string]interface{}); ok {
			if v, exists := m["uint64"]; exists {
				value = v.(uint64)
			} else {
				return nil, errors.New("fixed64 field requires value")
			}
		} else if v, ok := field.Value.(uint64); ok {
			value = v
		} else {
			return nil, errors.New("invalid fixed64 value type")
		}
		fixedBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(fixedBytes, value)
		buf = append(buf, fixedBytes...)

	case WireTypeFixed32:
		var value uint32
		if m, ok := field.Value.(map[string]interface{}); ok {
			if v, exists := m["uint32"]; exists {
				value = v.(uint32)
			} else {
				return nil, errors.New("fixed32 field requires value")
			}
		} else if v, ok := field.Value.(uint32); ok {
			value = v
		} else {
			return nil, errors.New("invalid fixed32 value type")
		}
		fixedBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(fixedBytes, value)
		buf = append(buf, fixedBytes...)

	case WireTypeLengthDelimited:
		var bytes []byte
		if m, ok := field.Value.(map[string]interface{}); ok {
			if v, exists := m["bytes"]; exists {
				bytes = v.([]byte)
			} else if v, exists := m["string"]; exists {
				bytes = []byte(v.(string))
			} else {
				return nil, errors.New("length-delimited field requires bytes or string value")
			}
		} else if b, ok := field.Value.([]byte); ok {
			bytes = b
		} else if s, ok := field.Value.(string); ok {
			bytes = []byte(s)
		} else {
			return nil, errors.New("invalid length-delimited value type")
		}
		buf = append(buf, EncodeVarint(uint64(len(bytes)))...)
		buf = append(buf, bytes...)

	default:
		return nil, errors.New("unsupported wire type for encoding")
	}

	return buf, nil
}

func EncodeProtobuf(fields []ProtobufField) ([]byte, error) {
	var buf []byte
	for _, field := range fields {
		encoded, err := EncodeProtobufField(field)
		if err != nil {
			return nil, err
		}
		buf = append(buf, encoded...)
	}
	return buf, nil
}
