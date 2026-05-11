package csvjson

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var pathPattern = regexp.MustCompile(`^([^\.\[\]]+)(\[\d+\])?$`)

func parsePathPart(part string) (string, *int) {
	m := pathPattern.FindStringSubmatch(part)
	if len(m) == 0 {
		return part, nil
	}
	name := m[1]
	if m[2] != "" {
		idxStr := m[2][1 : len(m[2])-1]
		if idx, err := strconv.Atoi(idxStr); err == nil {
			return name, &idx
		}
	}
	return name, nil
}

func splitPath(header string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	for i := 0; i < len(header); i++ {
		c := header[i]
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == '.' && !inQuote {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func setValue(obj map[string]interface{}, path string, value interface{}) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return
	}
	current := obj
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		name, idx := parsePathPart(part)
		if i == len(parts)-1 {
			if idx != nil {
				var arr []interface{}
				if existing, ok := current[name]; ok {
					if a, ok := existing.([]interface{}); ok {
						arr = a
					} else {
						arr = []interface{}{existing}
					}
				}
				for len(arr) <= *idx {
					arr = append(arr, nil)
				}
				arr[*idx] = value
				current[name] = arr
			} else {
				current[name] = value
			}
		} else {
			var next interface{}
			if idx != nil {
				var arr []interface{}
				if existing, ok := current[name]; ok {
					if a, ok := existing.([]interface{}); ok {
						arr = a
					} else {
						arr = []interface{}{existing}
					}
				}
				for len(arr) <= *idx {
					arr = append(arr, nil)
				}
				if arr[*idx] == nil {
					arr[*idx] = make(map[string]interface{})
				}
				next = arr[*idx]
				current[name] = arr
			} else {
				if existing, ok := current[name]; ok {
					if m, ok := existing.(map[string]interface{}); ok {
						next = m
					} else {
						newMap := make(map[string]interface{})
						newMap["_value"] = existing
						next = newMap
						current[name] = next
					}
				} else {
					newMap := make(map[string]interface{})
					current[name] = newMap
					next = newMap
				}
			}
			if m, ok := next.(map[string]interface{}); ok {
				current = m
			} else {
				return
			}
		}
	}
}

func compactArray(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, v := range arr {
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}

func compactArrays(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		for k, val := range v {
			v[k] = compactArrays(val)
		}
		return v
	case []interface{}:
		compacted := compactArray(v)
		for i, val := range compacted {
			compacted[i] = compactArrays(val)
		}
		return compacted
	default:
		return obj
	}
}

func readCSVLine(reader *bufio.Reader, delimiter Delimiter) ([]string, error) {
	var fields []string
	var current strings.Builder
	inQuote := false
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if current.Len() > 0 || len(fields) > 0 {
					fields = append(fields, current.String())
					return fields, nil
				}
				return nil, io.EOF
			}
			return nil, err
		}
		if inQuote {
			if r == '"' {
				next, _, err := reader.ReadRune()
				if err == nil {
					if next == '"' {
						current.WriteRune('"')
						continue
					}
					if next == '\r' {
						reader.ReadRune()
						fields = append(fields, current.String())
						return fields, nil
					}
					if next == '\n' {
						fields = append(fields, current.String())
						return fields, nil
					}
					reader.UnreadRune()
					inQuote = false
					continue
				}
				inQuote = false
				continue
			}
			current.WriteRune(r)
			continue
		}
		switch r {
		case '"':
			inQuote = true
		case rune(delimiter):
			fields = append(fields, current.String())
			current.Reset()
		case '\r':
			reader.ReadRune()
			fields = append(fields, current.String())
			return fields, nil
		case '\n':
			fields = append(fields, current.String())
			return fields, nil
		default:
			current.WriteRune(r)
		}
	}
}

func CSVToJSON(input io.Reader, opts *ConvertOptions) ([]map[string]interface{}, error) {
	if opts == nil {
		opts = DefaultOptions()
	}
	reader := bufio.NewReader(input)
	headers, err := readCSVLine(reader, opts.Delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}
	var records []map[string]interface{}
	for {
		row, err := readCSVLine(reader, opts.Delimiter)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read row: %w", err)
		}
		if len(row) == 1 && strings.TrimSpace(row[0]) == "" {
			continue
		}
		obj := make(map[string]interface{})
		for i, header := range headers {
			if i >= len(row) {
				setValue(obj, header, nil)
				continue
			}
			value := InferType(row[i], opts.MaxDigits)
			setValue(obj, header, value)
		}
		records = append(records, compactArrays(obj).(map[string]interface{}))
	}
	return records, nil
}
