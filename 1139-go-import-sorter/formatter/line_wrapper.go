package formatter

import (
	"strings"
	"unicode"
)

func processLineWidth(result *string, cfg Config) int {
	lines := strings.Split(*result, "\n")
	splitCount := 0
	newLines := make([]string, 0, len(lines))

	inString := false
	for _, line := range lines {
		lineLen := len(line)
		
		hasString := strings.Contains(line, `"`)
		if hasString {
			quoteCount := strings.Count(line, `"`)
			if quoteCount%2 != 0 {
				inString = !inString
			}
		}
		
		if lineLen <= cfg.LineWidth || inString {
			newLines = append(newLines, line)
			continue
		}

		trimmed := strings.TrimLeft(line, " \t")
		indent := line[:len(line)-len(trimmed)]

		if isLongString(line) {
			newLines = append(newLines, line+" // line too long")
			continue
		}

		wrapped, splits := wrapLine(line, indent, cfg.LineWidth)
		newLines = append(newLines, wrapped...)
		splitCount += splits
	}

	*result = strings.Join(newLines, "\n")
	return splitCount
}

func isLongString(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") {
		return false
	}
	
	quoteCount := strings.Count(trimmed, `"`)
	if quoteCount >= 2 {
		first := strings.Index(trimmed, `"`)
		last := strings.LastIndex(trimmed, `"`)
		if first < last {
			substr := trimmed[first:last]
			if strings.Count(substr, `\"`)*2 == strings.Count(substr, `"`)-2 {
				return true
			}
			if !strings.Contains(trimmed[last+1:], `"`) {
				return true
			}
		}
	}
	
	return false
}

func wrapLine(line, indent string, maxWidth int) ([]string, int) {
	trimmed := strings.TrimLeft(line, " \t")
	
	if len(trimmed) == 0 {
		return []string{line}, 0
	}

	lines := make([]string, 0)
	splits := 0

	if strings.HasPrefix(trimmed, "if ") && (strings.Contains(trimmed, "&&") || strings.Contains(trimmed, "||")) {
		lines, splits = wrapIfCondition(line, indent, maxWidth)
	} else if strings.Contains(trimmed, "func ") && strings.Contains(trimmed, "(") {
		lines, splits = wrapFuncSignature(line, indent, maxWidth)
	} else if strings.Contains(trimmed, "(") {
		lines, splits = wrapCallOrStruct(line, indent, maxWidth)
	} else {
		lines = []string{line}
	}

	return lines, splits
}

func wrapIfCondition(line, indent string, maxWidth int) ([]string, int) {
	trimmed := strings.TrimLeft(line, " \t")
	parts := make([]string, 0)
	ops := []string{" && ", " || "}
	
	current := trimmed
	for {
		found := -1
		var op string
		for _, o := range ops {
			if idx := strings.Index(current, o); idx > 0 && (found < 0 || idx < found) {
				found = idx
				op = o
			}
		}
		if found < 0 {
			parts = append(parts, current)
			break
		}
		parts = append(parts, current[:found+len(op)])
		current = current[found+len(op):]
	}

	if len(parts) <= 1 {
		return []string{line}, 0
	}

	result := make([]string, 0, len(parts))
	result = append(result, indent+parts[0])
	for i := 1; i < len(parts); i++ {
		result = append(result, indent+"\t"+parts[i])
	}

	return result, len(parts) - 1
}

func wrapFuncSignature(line, indent string, maxWidth int) ([]string, int) {
	trimmed := strings.TrimLeft(line, " \t")
	
	paramsStart := strings.Index(trimmed, "(")
	paramsEnd := findMatchingParen(trimmed, paramsStart)
	if paramsEnd < 0 {
		return []string{line}, 0
	}
	
	returns := strings.TrimSpace(trimmed[paramsEnd+1:])
	funcName := trimmed[:paramsStart]
	params := trimmed[paramsStart+1 : paramsEnd]
	
	if params == "" {
		return []string{line}, 0
	}
	
	paramsList := splitParams(params)
	if len(paramsList) <= 1 {
		return []string{line}, 0
	}
	
	result := make([]string, 0)
	result = append(result, indent+funcName+"(")
	for _, p := range paramsList {
		result = append(result, indent+"\t"+strings.TrimSpace(p)+",")
	}
	result = append(result, indent+")"+returns)
	
	return result, len(paramsList) - 1
}

func wrapCallOrStruct(line, indent string, maxWidth int) ([]string, int) {
	trimmed := strings.TrimLeft(line, " \t")
	
	parenStart := -1
	for i, ch := range trimmed {
		if ch == '(' || ch == '{' {
			parenStart = i
			break
		}
	}
	if parenStart < 0 {
		return []string{line}, 0
	}
	
	prefix := trimmed[:parenStart]
	openChar := trimmed[parenStart]
	closeChar := byte(')')
	if openChar == '{' {
		closeChar = '}'
	}
	
	parenEnd := findMatching(trimmed, parenStart, openChar, closeChar)
	if parenEnd < 0 {
		return []string{line}, 0
	}
	
	inner := trimmed[parenStart+1 : parenEnd]
	suffix := trimmed[parenEnd+1:]
	
	if strings.TrimSpace(inner) == "" {
		return []string{line}, 0
	}
	
	items := splitItems(inner, openChar)
	if len(items) <= 1 {
		return []string{line}, 0
	}
	
	result := make([]string, 0)
	result = append(result, indent+prefix+string(openChar))
	for _, item := range items {
		itemTrimmed := strings.TrimSpace(item)
		if itemTrimmed != "" {
			if openChar == '{' {
				result = append(result, indent+"\t"+itemTrimmed)
			} else {
				if !strings.HasSuffix(itemTrimmed, ",") {
					itemTrimmed = itemTrimmed + ","
				}
				result = append(result, indent+"\t"+itemTrimmed)
			}
		}
	}
	result = append(result, indent+string(closeChar)+suffix)
	
	return result, len(items) - 1
}

func splitParams(params string) []string {
	return splitItems(params, '(')
}

func splitItems(inner string, openChar byte) []string {
	var items []string
	var current strings.Builder
	depth := 0
	inString := false
	escaped := false

	for i := 0; i < len(inner); i++ {
		ch := inner[i]
		
		if inString {
			current.WriteByte(ch)
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
			current.WriteByte(ch)
		case '(', '{', '[':
			depth++
			current.WriteByte(ch)
		case ')', '}', ']':
			if depth > 0 {
				depth--
			}
			current.WriteByte(ch)
		case ',':
			if depth == 0 {
				items = append(items, current.String())
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}
	
	if current.Len() > 0 {
		items = append(items, current.String())
	}

	return items
}

func findMatchingParen(s string, start int) int {
	return findMatching(s, start, '(', ')')
}

func findMatching(s string, start int, open, close byte) int {
	depth := 1
	inString := false
	escaped := false
	
	for i := start + 1; i < len(s); i++ {
		ch := s[i]
		
		if inString {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			continue
		}

		if ch == open {
			depth++
		} else if ch == close {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	
	return -1
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if unicode.IsSpace(ch) {
			count++
		} else {
			break
		}
	}
	return count
}
