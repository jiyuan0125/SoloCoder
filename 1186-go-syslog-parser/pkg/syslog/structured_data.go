package syslog

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type SDParam struct {
	Name  string
	Value string
}

type SDElement struct {
	ID     string
	Params []SDParam
}

func (e *SDElement) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteString(e.ID)

	for _, p := range e.Params {
		buf.WriteByte(' ')
		buf.WriteString(p.Name)
		buf.WriteByte('=')
		buf.WriteByte('"')
		buf.WriteString(escapeSDParamValue(p.Value))
		buf.WriteByte('"')
	}

	buf.WriteByte(']')
	return buf.String()
}

func escapeSDParamValue(val string) string {
	var buf bytes.Buffer
	for _, r := range val {
		switch r {
		case '\\', '"', ']':
			buf.WriteByte('\\')
			buf.WriteRune(r)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

type StructuredData []*SDElement

func (sd StructuredData) IsNil() bool {
	return len(sd) == 0
}

func (sd StructuredData) String() string {
	if len(sd) == 0 {
		return "-"
	}

	var buf bytes.Buffer
	for _, e := range sd {
		buf.WriteString(e.String())
	}
	return buf.String()
}

func ParseStructuredData(s string) (StructuredData, error) {
	s = strings.TrimSpace(s)

	if s == "-" {
		return nil, nil
	}

	if s == "" {
		return nil, nil
	}

	var result StructuredData
	remaining := s

	for len(remaining) > 0 {
		if remaining[0] != '[' {
			return nil, fmt.Errorf("expected '[' at start of SD-ELEMENT")
		}

		elem, rest, err := parseSDElement(remaining)
		if err != nil {
			return nil, err
		}

		result = append(result, elem)
		remaining = rest
	}

	return result, nil
}

func parseSDElement(s string) (*SDElement, string, error) {
	if len(s) < 3 || s[0] != '[' {
		return nil, "", fmt.Errorf("invalid SD-ELEMENT format")
	}

	idx := 1
	sdID, idx, err := parseSDID(s, idx)
	if err != nil {
		return nil, "", err
	}

	elem := &SDElement{
		ID: sdID,
	}

	for idx < len(s) {
		for idx < len(s) && unicode.IsSpace(rune(s[idx])) {
			idx++
		}

		if idx >= len(s) {
			return nil, "", fmt.Errorf("unexpected end of SD-ELEMENT")
		}

		if s[idx] == ']' {
			idx++
			return elem, s[idx:], nil
		}

		param, newIdx, err := parseSDParam(s, idx)
		if err != nil {
			return nil, "", err
		}

		elem.Params = append(elem.Params, param)
		idx = newIdx
	}

	return nil, "", fmt.Errorf("unexpected end of SD-ELEMENT, missing ']'")
}

func parseSDID(s string, startIdx int) (string, int, error) {
	idx := startIdx
	for idx < len(s) {
		c := s[idx]
		if c == ']' || unicode.IsSpace(rune(c)) {
			break
		}
		if isPrintableUSASCII(c) {
			idx++
		} else {
			return "", 0, fmt.Errorf("invalid character in SD-ID: %c", c)
		}
	}

	if idx == startIdx {
		return "", 0, fmt.Errorf("empty SD-ID")
	}

	return s[startIdx:idx], idx, nil
}

func parseSDParam(s string, startIdx int) (SDParam, int, error) {
	idx := startIdx
	nameIdx := idx

	for idx < len(s) && s[idx] != '=' {
		if !isPrintableUSASCII(s[idx]) {
			return SDParam{}, 0, fmt.Errorf("invalid character in PARAM-NAME: %c", s[idx])
		}
		idx++
	}

	if idx >= len(s) || s[idx] != '=' {
		return SDParam{}, 0, fmt.Errorf("missing '=' in SD parameter")
	}

	if idx == nameIdx {
		return SDParam{}, 0, fmt.Errorf("empty PARAM-NAME")
	}

	name := s[nameIdx:idx]
	idx++

	if idx >= len(s) || s[idx] != '"' {
		return SDParam{}, 0, fmt.Errorf("missing opening quote for PARAM-VALUE")
	}

	idx++

	var valueBuilder strings.Builder
	for idx < len(s) {
		if s[idx] == '\\' {
			idx++
			if idx >= len(s) {
				return SDParam{}, 0, fmt.Errorf("incomplete escape sequence")
			}

			escapedChar := s[idx]
			switch escapedChar {
			case '"', '\\', ']':
				valueBuilder.WriteByte(escapedChar)
			default:
				valueBuilder.WriteByte('\\')
				valueBuilder.WriteByte(escapedChar)
			}
			idx++
		} else if s[idx] == '"' {
			idx++
			return SDParam{Name: name, Value: valueBuilder.String()}, idx, nil
		} else {
			valueBuilder.WriteByte(s[idx])
			idx++
		}
	}

	return SDParam{}, 0, fmt.Errorf("unexpected end of PARAM-VALUE, missing closing quote")
}

func isPrintableUSASCII(c byte) bool {
	return c >= 33 && c <= 126
}
