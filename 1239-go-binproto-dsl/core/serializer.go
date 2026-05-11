package core

import (
	"encoding/binary"
	"fmt"
	"math"
)

func ParseBinary(def *MessageDef, data []byte) (*ParseResult, error) {
	result := make(map[string]interface{})
	pos := 0
	var checksum uint8

	for i, field := range def.Fields {
		if field.IsChecksum {
			expectedChecksum := checksum
			if pos+1 > len(data) {
				return nil, &ParseError{Field: field.Name, Message: "truncated checksum field"}
			}
			actualChecksum := data[pos]
			pos++
			if actualChecksum != expectedChecksum {
				return nil, &ChecksumError{Expected: expectedChecksum, Actual: actualChecksum}
			}
			continue
		}

		value, bytesRead, err := readField(field, data, pos)
		if err != nil {
			pe, ok := err.(*ParseError)
			if ok {
				pe.Field = field.Name
				return nil, pe
			}
			return nil, &ParseError{Field: field.Name, Message: err.Error()}
		}

		for j := 0; j < bytesRead; j++ {
			checksum += data[pos+j]
		}
		pos += bytesRead

		_ = i
		result[field.Name] = value
	}

	return &ParseResult{Data: result}, nil
}

func readField(field Field, data []byte, pos int) (interface{}, int, error) {
	switch field.Type {
	case TypeUint8:
		if pos+1 > len(data) {
			return nil, 0, &ParseError{Message: "truncated uint8 field"}
		}
		return data[pos], 1, nil

	case TypeUint16:
		if pos+2 > len(data) {
			return nil, 0, &ParseError{Message: "truncated uint16 field"}
		}
		return binary.LittleEndian.Uint16(data[pos : pos+2]), 2, nil

	case TypeUint32:
		if pos+4 > len(data) {
			return nil, 0, &ParseError{Message: "truncated uint32 field"}
		}
		return binary.LittleEndian.Uint32(data[pos : pos+4]), 4, nil

	case TypeUint64:
		if pos+8 > len(data) {
			return nil, 0, &ParseError{Message: "truncated uint64 field"}
		}
		return binary.LittleEndian.Uint64(data[pos : pos+8]), 8, nil

	case TypeInt8:
		if pos+1 > len(data) {
			return nil, 0, &ParseError{Message: "truncated int8 field"}
		}
		return int8(data[pos]), 1, nil

	case TypeInt16:
		if pos+2 > len(data) {
			return nil, 0, &ParseError{Message: "truncated int16 field"}
		}
		return int16(binary.LittleEndian.Uint16(data[pos : pos+2])), 2, nil

	case TypeInt32:
		if pos+4 > len(data) {
			return nil, 0, &ParseError{Message: "truncated int32 field"}
		}
		return int32(binary.LittleEndian.Uint32(data[pos : pos+4])), 4, nil

	case TypeInt64:
		if pos+8 > len(data) {
			return nil, 0, &ParseError{Message: "truncated int64 field"}
		}
		return int64(binary.LittleEndian.Uint64(data[pos : pos+8])), 8, nil

	case TypeFixedString:
		if field.Length <= 0 {
			return nil, 0, &ParseError{Message: "invalid fixed_string length"}
		}
		actualLen := field.Length
		if pos+actualLen > len(data) {
			actualLen = len(data) - pos
		}
		strBytes := make([]byte, field.Length)
		copy(strBytes, data[pos:pos+actualLen])
		for actualLen < field.Length {
			strBytes[actualLen] = 0
			actualLen++
		}
		return string(strBytes), field.Length, nil

	case TypeVarString:
		if pos+2 > len(data) {
			return nil, 0, &ParseError{Message: "truncated var_string length field"}
		}
		strLen := binary.LittleEndian.Uint16(data[pos : pos+2])
		if pos+2+int(strLen) > len(data) {
			return nil, 0, &ParseError{
				Message: fmt.Sprintf("truncated var_string data: need %d bytes, have %d", strLen, len(data)-pos-2),
			}
		}
		return string(data[pos+2 : pos+2+int(strLen)]), 2 + int(strLen), nil

	case TypeBytes:
		if field.Length <= 0 {
			return nil, 0, &ParseError{Message: "invalid bytes length"}
		}
		actualLen := field.Length
		if pos+actualLen > len(data) {
			actualLen = len(data) - pos
		}
		bytesResult := make([]byte, field.Length)
		copy(bytesResult, data[pos:pos+actualLen])
		for actualLen < field.Length {
			bytesResult[actualLen] = 0
			actualLen++
		}
		return bytesResult, field.Length, nil

	default:
		return nil, 0, &ParseError{Message: fmt.Sprintf("unsupported field type: %s", field.Type)}
	}
}

func SerializeBinary(def *MessageDef, data map[string]interface{}) ([]byte, error) {
	result := make([]byte, 0)
	var checksum uint8

	for _, field := range def.Fields {
		if field.IsChecksum {
			result = append(result, checksum)
			continue
		}

		value, ok := data[field.Name]
		if !ok {
			return nil, &SerializeError{Field: field.Name, Message: "field not found in data"}
		}

		fieldBytes, err := writeField(field, value)
		if err != nil {
			se, ok := err.(*SerializeError)
			if ok {
				se.Field = field.Name
				return nil, se
			}
			return nil, &SerializeError{Field: field.Name, Message: err.Error()}
		}

		for _, b := range fieldBytes {
			checksum += b
		}
		result = append(result, fieldBytes...)
	}

	return result, nil
}

func writeField(field Field, value interface{}) ([]byte, error) {
	result := make([]byte, 0)

	switch field.Type {
	case TypeUint8:
		var v uint8
		switch vv := value.(type) {
		case float64:
			if vv < 0 || vv > math.MaxUint8 {
				return nil, &SerializeError{Message: "value out of range for uint8"}
			}
			v = uint8(vv)
		case int:
			if vv < 0 || vv > math.MaxUint8 {
				return nil, &SerializeError{Message: "value out of range for uint8"}
			}
			v = uint8(vv)
		case uint8:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected uint8 value"}
		}
		result = append(result, v)

	case TypeUint16:
		var v uint16
		switch vv := value.(type) {
		case float64:
			if vv < 0 || vv > math.MaxUint16 {
				return nil, &SerializeError{Message: "value out of range for uint16"}
			}
			v = uint16(vv)
		case int:
			if vv < 0 || vv > math.MaxUint16 {
				return nil, &SerializeError{Message: "value out of range for uint16"}
			}
			v = uint16(vv)
		case uint16:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected uint16 value"}
		}
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, v)
		result = append(result, buf...)

	case TypeUint32:
		var v uint32
		switch vv := value.(type) {
		case float64:
			if vv < 0 || vv > math.MaxUint32 {
				return nil, &SerializeError{Message: "value out of range for uint32"}
			}
			v = uint32(vv)
		case int:
			if vv < 0 || vv > math.MaxUint32 {
				return nil, &SerializeError{Message: "value out of range for uint32"}
			}
			v = uint32(vv)
		case uint32:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected uint32 value"}
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, v)
		result = append(result, buf...)

	case TypeUint64:
		var v uint64
		switch vv := value.(type) {
		case float64:
			if vv < 0 || vv > float64(math.MaxUint64) {
				return nil, &SerializeError{Message: "value out of range for uint64"}
			}
			v = uint64(vv)
		case int:
			if vv < 0 {
				return nil, &SerializeError{Message: "value out of range for uint64"}
			}
			v = uint64(vv)
		case uint64:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected uint64 value"}
		}
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, v)
		result = append(result, buf...)

	case TypeInt8:
		var v int8
		switch vv := value.(type) {
		case float64:
			if vv < math.MinInt8 || vv > math.MaxInt8 {
				return nil, &SerializeError{Message: "value out of range for int8"}
			}
			v = int8(vv)
		case int:
			if vv < math.MinInt8 || vv > math.MaxInt8 {
				return nil, &SerializeError{Message: "value out of range for int8"}
			}
			v = int8(vv)
		case int8:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected int8 value"}
		}
		result = append(result, byte(v))

	case TypeInt16:
		var v int16
		switch vv := value.(type) {
		case float64:
			if vv < math.MinInt16 || vv > math.MaxInt16 {
				return nil, &SerializeError{Message: "value out of range for int16"}
			}
			v = int16(vv)
		case int:
			if vv < math.MinInt16 || vv > math.MaxInt16 {
				return nil, &SerializeError{Message: "value out of range for int16"}
			}
			v = int16(vv)
		case int16:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected int16 value"}
		}
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(v))
		result = append(result, buf...)

	case TypeInt32:
		var v int32
		switch vv := value.(type) {
		case float64:
			if vv < math.MinInt32 || vv > math.MaxInt32 {
				return nil, &SerializeError{Message: "value out of range for int32"}
			}
			v = int32(vv)
		case int:
			if vv < math.MinInt32 || vv > math.MaxInt32 {
				return nil, &SerializeError{Message: "value out of range for int32"}
			}
			v = int32(vv)
		case int32:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected int32 value"}
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))
		result = append(result, buf...)

	case TypeInt64:
		var v int64
		switch vv := value.(type) {
		case float64:
			if vv < float64(math.MinInt64) || vv > float64(math.MaxInt64) {
				return nil, &SerializeError{Message: "value out of range for int64"}
			}
			v = int64(vv)
		case int:
			v = int64(vv)
		case int64:
			v = vv
		default:
			return nil, &SerializeError{Message: "expected int64 value"}
		}
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(v))
		result = append(result, buf...)

	case TypeFixedString:
		s, ok := value.(string)
		if !ok {
			return nil, &SerializeError{Message: "expected string value for fixed_string"}
		}
		if field.Length <= 0 {
			return nil, &SerializeError{Message: "invalid fixed_string length"}
		}
		strBytes := make([]byte, field.Length)
		copy(strBytes, s)
		result = append(result, strBytes...)

	case TypeVarString:
		s, ok := value.(string)
		if !ok {
			return nil, &SerializeError{Message: "expected string value for var_string"}
		}
		if len(s) > math.MaxUint16 {
			return nil, &SerializeError{Message: "var_string too long"}
		}
		lenBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(s)))
		result = append(result, lenBuf...)
		result = append(result, []byte(s)...)

	case TypeBytes:
		var b []byte
		switch vv := value.(type) {
		case []byte:
			b = vv
		case []interface{}:
			b = make([]byte, len(vv))
			for i, elem := range vv {
				switch e := elem.(type) {
				case float64:
					if e < 0 || e > 255 {
						return nil, &SerializeError{Message: "byte value out of range"}
					}
					b[i] = byte(e)
				case int:
					if e < 0 || e > 255 {
						return nil, &SerializeError{Message: "byte value out of range"}
					}
					b[i] = byte(e)
				default:
					return nil, &SerializeError{Message: "expected byte value in bytes array"}
				}
			}
		default:
			return nil, &SerializeError{Message: "expected bytes value"}
		}

		if field.Length <= 0 {
			return nil, &SerializeError{Message: "invalid bytes length"}
		}
		bytesResult := make([]byte, field.Length)
		copy(bytesResult, b)
		result = append(result, bytesResult...)

	default:
		return nil, &SerializeError{Message: fmt.Sprintf("unsupported field type: %s", field.Type)}
	}

	return result, nil
}
