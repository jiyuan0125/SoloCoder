package templater

import (
	"regexp"
	"strings"
)

type SQLTemplater struct{}

func NewSQLTemplater() *SQLTemplater {
	return &SQLTemplater{}
}

func (t *SQLTemplater) TemplateSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	sql = t.normalizeWhitespace(sql)
	sql = t.collapseINClause(sql)
	sql = t.replaceStringLiterals(sql)
	sql = t.replaceNumericValues(sql)
	return strings.TrimSpace(sql)
}

func (t *SQLTemplater) normalizeWhitespace(sql string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(sql, " ")
}

func (t *SQLTemplater) collapseINClause(sql string) string {
	re := regexp.MustCompile(`(?i)\bIN\s*\(\s*([^)]+)\s*\)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) > 1 {
			content := submatch[1]
			parts := strings.Split(content, ",")
			allLiterals := true
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if !t.isLiteralValue(part) {
					allLiterals = false
					break
				}
			}
			if allLiterals {
				return "IN (?)"
			}
		}
		return match
	})
}

func (t *SQLTemplater) isLiteralValue(s string) bool {
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		return true
	}
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return true
	}
	re := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	return re.MatchString(s)
}

func (t *SQLTemplater) replaceStringLiterals(sql string) string {
	var result []byte
	i := 0
	for i < len(sql) {
		if sql[i] == '\'' || sql[i] == '"' {
			quote := sql[i]
			result = append(result, '?')
			i++
			for i < len(sql) {
				if sql[i] == quote {
					if i+1 < len(sql) && sql[i+1] == quote {
						i += 2
						continue
					}
					i++
					break
				}
				if sql[i] == '\\' && i+1 < len(sql) {
					i += 2
					continue
				}
				i++
			}
		} else {
			result = append(result, sql[i])
			i++
		}
	}
	return string(result)
}

func (t *SQLTemplater) replaceNumericValues(sql string) string {
	re := regexp.MustCompile(`(\b[+-]?\d+(\.\d+)?(e[+-]?\d+)?\b)`)

	matches := re.FindAllStringIndex(sql, -1)
	if len(matches) == 0 {
		return sql
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		result.WriteString(sql[lastEnd:start])

		if t.isNumericInContextOfIdentifierAt(sql, start, end) {
			result.WriteString(sql[start:end])
		} else {
			result.WriteString("?")
		}

		lastEnd = end
	}

	result.WriteString(sql[lastEnd:])
	return result.String()
}

func (t *SQLTemplater) isNumericInContextOfIdentifierAt(sql string, start, end int) bool {
	if start > 0 {
		prevChar := sql[start-1]
		if (prevChar >= 'a' && prevChar <= 'z') ||
			(prevChar >= 'A' && prevChar <= 'Z') ||
			prevChar == '_' || prevChar == '$' {
			return true
		}
	}

	if end < len(sql) {
		nextChar := sql[end]
		if (nextChar >= 'a' && nextChar <= 'z') ||
			(nextChar >= 'A' && nextChar <= 'Z') ||
			nextChar == '_' || nextChar == '$' {
			return true
		}
	}

	return false
}
