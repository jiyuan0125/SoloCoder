package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

type decoder struct {
	data   []byte
	offset int
}

func Decode(data []byte) (interface{}, error) {
	d := &decoder{data: data, offset: 0}
	result, err := d.decodeValue()
	if err != nil {
		return nil, err
	}
	if d.offset != len(data) {
		return nil, fmt.Errorf("unexpected trailing data at position %d", d.offset)
	}
	return result, nil
}

func (d *decoder) decodeValue() (interface{}, error) {
	if d.offset >= len(d.data) {
		return nil, errors.New("unexpected end of data")
	}

	switch d.data[d.offset] {
	case 'i':
		return d.decodeInteger()
	case 'l':
		return d.decodeList()
	case 'd':
		return d.decodeDict()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.decodeByteString()
	default:
		return nil, fmt.Errorf("unexpected character '%c' at position %d", d.data[d.offset], d.offset)
	}
}

func (d *decoder) decodeInteger() (int64, error) {
	start := d.offset
	if d.data[d.offset] != 'i' {
		return 0, fmt.Errorf("expected 'i' at position %d", d.offset)
	}
	d.offset++

	if d.offset >= len(d.data) {
		return 0, fmt.Errorf("incomplete integer at position %d", start)
	}

	negative := false
	if d.data[d.offset] == '-' {
		negative = true
		d.offset++
		if d.offset >= len(d.data) {
			return 0, fmt.Errorf("incomplete integer at position %d", start)
		}
	}

	numStart := d.offset
	for d.offset < len(d.data) && d.data[d.offset] >= '0' && d.data[d.offset] <= '9' {
		d.offset++
	}

	if numStart == d.offset {
		return 0, fmt.Errorf("invalid integer: no digits at position %d", start)
	}

	numStr := string(d.data[numStart:d.offset])

	if len(numStr) > 1 && numStr[0] == '0' {
		return 0, fmt.Errorf("invalid integer: leading zeros at position %d", start)
	}

	if d.offset >= len(d.data) || d.data[d.offset] != 'e' {
		return 0, fmt.Errorf("expected 'e' at position %d", d.offset)
	}
	d.offset++

	if len(numStr) > 19 {
		return 0, fmt.Errorf("integer overflow at position %d", start)
	}

	if negative {
		numStr = "-" + numStr
	}

	value, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("integer overflow at position %d: %v", start, err)
	}

	return value, nil
}

func (d *decoder) decodeByteString() ([]byte, error) {
	start := d.offset

	lenStart := d.offset
	for d.offset < len(d.data) && d.data[d.offset] >= '0' && d.data[d.offset] <= '9' {
		d.offset++
	}

	if lenStart == d.offset {
		return nil, fmt.Errorf("expected length at position %d", start)
	}

	if len(d.data) > lenStart+1 && d.data[lenStart] == '0' && d.data[lenStart+1] != ':' {
		return nil, fmt.Errorf("invalid byte string: leading zeros in length at position %d", start)
	}

	if d.offset >= len(d.data) || d.data[d.offset] != ':' {
		return nil, fmt.Errorf("expected ':' at position %d", d.offset)
	}

	lenStr := string(d.data[lenStart:d.offset])
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid length at position %d: %v", start, err)
	}

	d.offset++

	if d.offset+length > len(d.data) {
		return nil, fmt.Errorf("byte string truncated at position %d, expected %d bytes", start, length)
	}

	result := make([]byte, length)
	copy(result, d.data[d.offset:d.offset+length])
	d.offset += length

	return result, nil
}

func (d *decoder) decodeList() ([]interface{}, error) {
	start := d.offset
	if d.data[d.offset] != 'l' {
		return nil, fmt.Errorf("expected 'l' at position %d", d.offset)
	}
	d.offset++

	result := make([]interface{}, 0)
	for {
		if d.offset >= len(d.data) {
			return nil, fmt.Errorf("incomplete list at position %d", start)
		}
		if d.data[d.offset] == 'e' {
			d.offset++
			return result, nil
		}
		elem, err := d.decodeValue()
		if err != nil {
			return nil, err
		}
		result = append(result, elem)
	}
}

func (d *decoder) decodeDict() (map[string]interface{}, error) {
	start := d.offset
	if d.data[d.offset] != 'd' {
		return nil, fmt.Errorf("expected 'd' at position %d", d.offset)
	}
	d.offset++

	result := make(map[string]interface{})
	var lastKey []byte

	for {
		if d.offset >= len(d.data) {
			return nil, fmt.Errorf("incomplete dictionary at position %d", start)
		}
		if d.data[d.offset] == 'e' {
			d.offset++
			return result, nil
		}

		keyBytes, err := d.decodeByteString()
		if err != nil {
			return nil, fmt.Errorf("dictionary key error: %v", err)
		}

		if lastKey != nil {
			cmp := compareByteSlices(lastKey, keyBytes)
			if cmp > 0 {
				return nil, fmt.Errorf("dictionary keys not sorted at position %d, key '%s' should come before '%s'",
					d.offset-len(keyBytes), string(keyBytes), string(lastKey))
			}
			if cmp == 0 {
				return nil, fmt.Errorf("duplicate dictionary key '%s' at position %d", string(keyBytes), d.offset-len(keyBytes))
			}
		}
		lastKey = keyBytes

		value, err := d.decodeValue()
		if err != nil {
			return nil, err
		}

		result[string(keyBytes)] = value
	}
}

func compareByteSlices(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
