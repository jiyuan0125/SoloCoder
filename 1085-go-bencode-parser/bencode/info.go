package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

type Info struct {
	OuterType    string
	DictKeyCount int
	ListElemCount int
	MaxDepth     int
}

func ParseInfo(data []byte) (*Info, error) {
	d := &decoder{data: data, offset: 0}
	info := &Info{}

	outerType, depth, dictCount, listCount, err := d.analyzeValue()
	if err != nil {
		return nil, err
	}
	if d.offset != len(data) {
		return nil, fmt.Errorf("unexpected trailing data at position %d", d.offset)
	}

	info.OuterType = outerType
	info.MaxDepth = depth
	switch outerType {
	case "dictionary":
		info.DictKeyCount = dictCount
	case "list":
		info.ListElemCount = listCount
	}

	return info, nil
}

func (d *decoder) analyzeValue() (string, int, int, int, error) {
	if d.offset >= len(d.data) {
		return "", 0, 0, 0, errors.New("unexpected end of data")
	}

	switch d.data[d.offset] {
	case 'i':
		return d.analyzeInteger()
	case 'l':
		return d.analyzeList()
	case 'd':
		return d.analyzeDict()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.analyzeByteString()
	default:
		return "", 0, 0, 0, fmt.Errorf("unexpected character '%c' at position %d", d.data[d.offset], d.offset)
	}
}

func (d *decoder) analyzeInteger() (string, int, int, int, error) {
	start := d.offset
	if d.data[d.offset] != 'i' {
		return "", 0, 0, 0, fmt.Errorf("expected 'i' at position %d", d.offset)
	}
	d.offset++

	if d.offset >= len(d.data) {
		return "", 0, 0, 0, fmt.Errorf("incomplete integer at position %d", start)
	}

	if d.data[d.offset] == '-' {
		d.offset++
		if d.offset >= len(d.data) {
			return "", 0, 0, 0, fmt.Errorf("incomplete integer at position %d", start)
		}
	}

	numStart := d.offset
	for d.offset < len(d.data) && d.data[d.offset] >= '0' && d.data[d.offset] <= '9' {
		d.offset++
	}

	if numStart == d.offset {
		return "", 0, 0, 0, fmt.Errorf("invalid integer: no digits at position %d", start)
	}

	if d.offset >= len(d.data) || d.data[d.offset] != 'e' {
		return "", 0, 0, 0, fmt.Errorf("expected 'e' at position %d", d.offset)
	}
	d.offset++

	return "integer", 1, 0, 0, nil
}

func (d *decoder) analyzeByteString() (string, int, int, int, error) {
	start := d.offset

	lenStart := d.offset
	for d.offset < len(d.data) && d.data[d.offset] >= '0' && d.data[d.offset] <= '9' {
		d.offset++
	}

	if lenStart == d.offset {
		return "", 0, 0, 0, fmt.Errorf("expected length at position %d", start)
	}

	if d.offset >= len(d.data) || d.data[d.offset] != ':' {
		return "", 0, 0, 0, fmt.Errorf("expected ':' at position %d", d.offset)
	}

	lenStr := string(d.data[lenStart:d.offset])
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("invalid length at position %d: %v", start, err)
	}

	d.offset++

	if d.offset+length > len(d.data) {
		return "", 0, 0, 0, fmt.Errorf("byte string truncated at position %d, expected %d bytes", start, length)
	}

	d.offset += length

	return "string", 1, 0, 0, nil
}

func (d *decoder) analyzeList() (string, int, int, int, error) {
	start := d.offset
	if d.data[d.offset] != 'l' {
		return "", 0, 0, 0, fmt.Errorf("expected 'l' at position %d", d.offset)
	}
	d.offset++

	maxDepth := 1
	elemCount := 0

	for {
		if d.offset >= len(d.data) {
			return "", 0, 0, 0, fmt.Errorf("incomplete list at position %d", start)
		}
		if d.data[d.offset] == 'e' {
			d.offset++
			return "list", maxDepth, 0, elemCount, nil
		}
		_, childDepth, _, _, err := d.analyzeValue()
		if err != nil {
			return "", 0, 0, 0, err
		}
		if childDepth+1 > maxDepth {
			maxDepth = childDepth + 1
		}
		elemCount++
	}
}

func (d *decoder) analyzeDict() (string, int, int, int, error) {
	start := d.offset
	if d.data[d.offset] != 'd' {
		return "", 0, 0, 0, fmt.Errorf("expected 'd' at position %d", d.offset)
	}
	d.offset++

	maxDepth := 1
	keyCount := 0
	var lastKey []byte

	for {
		if d.offset >= len(d.data) {
			return "", 0, 0, 0, fmt.Errorf("incomplete dictionary at position %d", start)
		}
		if d.data[d.offset] == 'e' {
			d.offset++
			return "dictionary", maxDepth, keyCount, 0, nil
		}

		keyBytes, err := d.decodeByteString()
		if err != nil {
			return "", 0, 0, 0, fmt.Errorf("dictionary key error: %v", err)
		}

		if lastKey != nil {
			cmp := compareByteSlices(lastKey, keyBytes)
			if cmp > 0 {
				return "", 0, 0, 0, fmt.Errorf("dictionary keys not sorted at position %d, key '%s' should come before '%s'",
					d.offset-len(keyBytes), string(keyBytes), string(lastKey))
			}
			if cmp == 0 {
				return "", 0, 0, 0, fmt.Errorf("duplicate dictionary key '%s' at position %d", string(keyBytes), d.offset-len(keyBytes))
			}
		}
		lastKey = keyBytes

		_, childDepth, _, _, err := d.analyzeValue()
		if err != nil {
			return "", 0, 0, 0, err
		}
		if childDepth+1 > maxDepth {
			maxDepth = childDepth + 1
		}
		keyCount++
	}
}
