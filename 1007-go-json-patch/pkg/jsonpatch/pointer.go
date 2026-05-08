package jsonpatch

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrPointerInvalidEscape = errors.New("invalid escape sequence in JSON Pointer")
	ErrPointerInvalidIndex  = errors.New("invalid array index in JSON Pointer")
	ErrPointerNotFound      = errors.New("path not found")
)

type Pointer []string

func ParsePointer(s string) (Pointer, error) {
	if s == "" {
		return Pointer{}, nil
	}
	if s[0] != '/' {
		return nil, errors.New("JSON Pointer must start with '/' or be empty")
	}
	
	parts := strings.Split(s[1:], "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.Contains(part, "~") {
			if strings.Contains(part, "~") {
				part = strings.ReplaceAll(part, "~1", "/")
				part = strings.ReplaceAll(part, "~0", "~")
			}
			for _, r := range part {
				if r == '~' {
					return nil, ErrPointerInvalidEscape
				}
			}
		}
		result = append(result, part)
	}
	return result, nil
}

func (p Pointer) Encode() string {
	if len(p) == 0 {
		return ""
	}
	
	var sb strings.Builder
	sb.WriteByte('/')
	for i, part := range p {
		if i > 0 {
			sb.WriteByte('/')
		}
		part = strings.ReplaceAll(part, "~", "~0")
		part = strings.ReplaceAll(part, "/", "~1")
		sb.WriteString(part)
	}
	return sb.String()
}

func (p Pointer) parent() Pointer {
	if len(p) == 0 {
		return nil
	}
	return p[:len(p)-1]
}

func (p Pointer) lastKey() string {
	if len(p) == 0 {
		return ""
	}
	return p[len(p)-1]
}

func (p Pointer) IsParentOf(other Pointer) bool {
	if len(p) >= len(other) {
		return false
	}
	for i := range p {
		if p[i] != other[i] {
			return false
		}
	}
	return true
}

func getValue(doc interface{}, ptr Pointer) (interface{}, error) {
	current := doc
	for _, key := range ptr {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[key]
			if !ok {
				return nil, ErrPointerNotFound
			}
			current = val
		case []interface{}:
			if key == "-" {
				return nil, ErrPointerNotFound
			}
			idx, err := parseArrayIndex(key, len(v))
			if err != nil {
				return nil, err
			}
			if idx >= len(v) {
				return nil, ErrPointerNotFound
			}
			current = v[idx]
		default:
			return nil, ErrPointerNotFound
		}
	}
	return current, nil
}

func parseArrayIndex(key string, maxLen int) (int, error) {
	if key == "" {
		return 0, nil
	}
	if key[0] == '0' && len(key) > 1 {
		return 0, ErrPointerInvalidIndex
	}
	if key == "-" {
		return maxLen, nil
	}
	idx, err := strconv.Atoi(key)
	if err != nil {
		return 0, ErrPointerInvalidIndex
	}
	if idx < 0 {
		return 0, ErrPointerInvalidIndex
	}
	return idx, nil
}
