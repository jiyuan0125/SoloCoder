package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// loadJSONFile loads configuration from a JSON file.
func loadJSONFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	config := New()
	if m, ok := raw.(map[string]interface{}); ok {
		flattenMap(config, m, "")
	}

	return config, nil
}

// loadYAMLFile loads configuration from a YAML file.
// Uses a simplified YAML parser implemented with standard library.
func loadYAMLFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	parsed, err := parseYAML(string(data))
	if err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}

	config := New()
	if m, ok := parsed.(map[string]interface{}); ok {
		flattenMap(config, m, "")
	}

	return config, nil
}

// flattenMap recursively flattens a nested map into dot-separated paths.
func flattenMap(config *Config, m map[string]interface{}, prefix string) {
	for key, value := range m {
		var fullPath string
		if prefix == "" {
			fullPath = key
		} else {
			fullPath = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flattenMap(config, v, fullPath)
		case []interface{}:
			config.Set(fullPath, v, SourceFile)
		default:
			config.Set(fullPath, v, SourceFile)
		}
	}
}

// YAML parser implementation using standard library.
// This is a simplified YAML parser that handles basic YAML structures.

// yamlNode represents a node in the YAML AST.
type yamlNode struct {
	kind      yamlNodeKind
	value     interface{}
	children  map[string]*yamlNode
	listItems []*yamlNode
	indent    int
}

type yamlNodeKind int

const (
	yamlNodeScalar yamlNodeKind = iota
	yamlNodeMap
	yamlNodeList
)

// parseYAML parses YAML string into an interface{}.
func parseYAML(content string) (interface{}, error) {
	lines := strings.Split(content, "\n")
	var filteredLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	if len(filteredLines) == 0 {
		return map[string]interface{}{}, nil
	}

	root, err := parseYAMLContent(filteredLines, 0, -1)
	if err != nil {
		return nil, err
	}

	return yamlNodeToInterface(root), nil
}

// parseYAMLContent parses YAML content with a given indentation level.
func parseYAMLContent(lines []string, startIdx int, parentIndent int) (*yamlNode, error) {
	if startIdx >= len(lines) {
		return nil, nil
	}

	line := lines[startIdx]
	indent := countLeadingSpaces(line)

	if indent <= parentIndent {
		return nil, nil
	}

	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "- ") {
		return parseYAMLList(lines, startIdx, parentIndent)
	}

	if idx := strings.Index(trimmed, ":"); idx != -1 {
		return parseYAMLMapEntry(lines, startIdx, parentIndent)
	}

	return &yamlNode{
		kind:  yamlNodeScalar,
		value: parseYAMLScalar(trimmed),
	}, nil
}

// parseYAMLList parses a YAML list.
func parseYAMLList(lines []string, startIdx int, parentIndent int) (*yamlNode, error) {
	items := []*yamlNode{}
	currentIdx := startIdx

	for currentIdx < len(lines) {
		line := lines[currentIdx]
		indent := countLeadingSpaces(line)

		if indent <= parentIndent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			value := strings.TrimPrefix(trimmed, "- ")
			value = strings.TrimSpace(value)

			if value == "" || strings.HasPrefix(value, "#") {
				if currentIdx+1 < len(lines) {
					nextLine := lines[currentIdx+1]
					nextIndent := countLeadingSpaces(nextLine)
					if nextIndent > indent {
						node, err := parseYAMLContent(lines, currentIdx+1, indent)
						if err != nil {
							return nil, err
						}
						if node != nil {
							items = append(items, node)
						}
						consumed := countConsumedLines(lines, currentIdx+1, indent)
						currentIdx += consumed + 1
						continue
					}
				}
			} else if idx := strings.Index(value, ":"); idx != -1 && idx != len(value)-1 {
				nestedLines := []string{strings.Repeat(" ", indent+2) + value}
				for i := currentIdx + 1; i < len(lines); i++ {
					nextLine := lines[i]
					nextIndent := countLeadingSpaces(nextLine)
					if nextIndent <= indent {
						break
					}
					nestedLines = append(nestedLines, nextLine)
				}
				node, err := parseYAML(nestedLines[0])
				if err != nil {
					return nil, err
				}
				if node != nil {
					items = append(items, &yamlNode{
						kind:  yamlNodeScalar,
						value: parseYAMLScalar(value),
					})
				}
			} else {
				node := &yamlNode{
					kind:  yamlNodeScalar,
					value: parseYAMLScalar(value),
				}
				items = append(items, node)
			}
		}

		currentIdx++
	}

	return &yamlNode{
		kind:      yamlNodeList,
		listItems: items,
	}, nil
}

// parseYAMLMapEntry parses a YAML map entry.
func parseYAMLMapEntry(lines []string, startIdx int, parentIndent int) (*yamlNode, error) {
	children := make(map[string]*yamlNode)
	currentIdx := startIdx

	for currentIdx < len(lines) {
		line := lines[currentIdx]
		indent := countLeadingSpaces(line)

		if indent <= parentIndent {
			break
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "- ") {
			if parentIndent >= 0 {
				break
			}

			listNode, err := parseYAMLList(lines, currentIdx, indent-1)
			if err != nil {
				return nil, err
			}
			children[""] = listNode
			consumed := countConsumedLines(lines, currentIdx, indent-1)
			currentIdx += consumed
			continue
		}

		idx := strings.Index(trimmed, ":")
		if idx == -1 {
			currentIdx++
			continue
		}

		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])

		if strings.HasPrefix(value, "#") || value == "" {
			if currentIdx+1 < len(lines) {
				nextLine := lines[currentIdx+1]
				nextIndent := countLeadingSpaces(nextLine)
				if nextIndent > indent {
					node, err := parseYAMLContent(lines, currentIdx+1, indent)
					if err != nil {
						return nil, err
					}
					if node != nil {
						children[key] = node
					}
					consumed := countConsumedLines(lines, currentIdx+1, indent)
					currentIdx += consumed + 1
					continue
				}
			}
			children[key] = &yamlNode{kind: yamlNodeScalar, value: nil}
		} else {
			children[key] = &yamlNode{
				kind:  yamlNodeScalar,
				value: parseYAMLScalar(value),
			}
		}

		currentIdx++
	}

	return &yamlNode{
		kind:     yamlNodeMap,
		children: children,
	}, nil
}

// countConsumedLines counts how many lines are consumed by a nested structure.
func countConsumedLines(lines []string, startIdx int, parentIndent int) int {
	count := 0
	for i := startIdx; i < len(lines); i++ {
		indent := countLeadingSpaces(lines[i])
		if indent <= parentIndent {
			break
		}
		count++
	}
	return count
}

// countLeadingSpaces counts leading spaces in a line.
func countLeadingSpaces(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 2
		} else {
			break
		}
	}
	return count
}

// parseYAMLScalar parses a YAML scalar value.
func parseYAMLScalar(value string) interface{} {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "#") {
		return nil
	}

	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\"")
	}
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return strings.Trim(value, "'")
	}

	lower := strings.ToLower(value)
	if lower == "null" || lower == "~" || value == "" {
		return nil
	}
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}

	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	return value
}

// yamlNodeToInterface converts a yamlNode to interface{}.
func yamlNodeToInterface(node *yamlNode) interface{} {
	if node == nil {
		return nil
	}

	switch node.kind {
	case yamlNodeScalar:
		return node.value

	case yamlNodeMap:
		result := make(map[string]interface{})
		for key, child := range node.children {
			result[key] = yamlNodeToInterface(child)
		}
		return result

	case yamlNodeList:
		result := make([]interface{}, len(node.listItems))
		for i, item := range node.listItems {
			result[i] = yamlNodeToInterface(item)
		}
		return result

	default:
		return nil
	}
}
