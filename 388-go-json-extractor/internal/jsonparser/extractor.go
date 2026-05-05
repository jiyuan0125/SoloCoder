package jsonparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"json-extractor/pkg/common"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) ExtractFromFile(filename string, fields []string, outputJSON, compact bool) ([]string, error) {
	var reader io.Reader

	if filename == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	return e.ExtractFromReader(reader, fields, outputJSON, compact)
}

func (e *Extractor) ExtractFromReader(reader io.Reader, fields []string, outputJSON, compact bool) ([]string, error) {
	var results []string
	var allResults map[string][]interface{}

	if outputJSON {
		allResults = make(map[string][]interface{})
		for _, field := range fields {
			allResults[field] = []interface{}{}
		}
	}

	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		for _, field := range fields {
			values := e.extractField(data, field)
			if outputJSON {
				allResults[field] = append(allResults[field], values...)
			} else {
				for _, val := range values {
					if compact {
						results = append(results, common.FormatValue(val, true))
					} else {
						results = append(results, fmt.Sprintf("%s=%s", field, common.FormatValue(val, false)))
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if outputJSON {
		var jsonData []byte
		var err error
		if compact {
			jsonData, err = json.Marshal(allResults)
		} else {
			jsonData, err = json.MarshalIndent(allResults, "", "  ")
		}
		if err != nil {
			return nil, err
		}
		return []string{string(jsonData)}, nil
	}

	return results, nil
}

func (e *Extractor) extractField(data interface{}, fieldPath string) []interface{} {
	parts := splitPath(fieldPath)
	return e.traverse(data, parts)
}

func splitPath(path string) []string {
	var parts []string
	var current strings.Builder

	for _, char := range path {
		if char == '.' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (e *Extractor) traverse(data interface{}, parts []string) []interface{} {
	if len(parts) == 0 {
		return []interface{}{data}
	}

	part := parts[0]
	remainingParts := parts[1:]

	if data == nil {
		return []interface{}{common.NotFound}
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if part == "*" {
			var results []interface{}
			for _, value := range v {
				results = append(results, e.traverse(value, remainingParts)...)
			}
			return results
		}

		if value, exists := v[part]; exists {
			return e.traverse(value, remainingParts)
		}
		return []interface{}{common.NotFound}

	case []interface{}:
		if part == "*" {
			var results []interface{}
			for _, item := range v {
				results = append(results, e.traverse(item, remainingParts)...)
			}
			return results
		}

		index, err := strconv.Atoi(part)
		if err != nil {
			return []interface{}{common.NotFound}
		}

		if index < 0 || index >= len(v) {
			return []interface{}{common.NotFound}
		}

		return e.traverse(v[index], remainingParts)

	default:
		return []interface{}{common.NotFound}
	}
}
