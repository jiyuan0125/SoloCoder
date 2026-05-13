package security

import (
	"regexp"
	"strings"
)

var dangerousKeywords = []string{
	"DROP",
	"DELETE",
	"UPDATE",
	"INSERT",
	"CREATE",
	"ALTER",
	"TRUNCATE",
	"GRANT",
	"REVOKE",
	"EXEC",
	"EXECUTE",
	"UNION",
}

var commentPattern = regexp.MustCompile(`--|\/\*|\*\/`)
var semicolonPattern = regexp.MustCompile(`;`)

func CheckDangerousKeywords(input string) (bool, []string) {
	upper := strings.ToUpper(input)
	var found []string

	for _, kw := range dangerousKeywords {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(kw) + `\b`)
		if pattern.MatchString(upper) {
			found = append(found, kw)
		}
	}

	return len(found) > 0, found
}

func CheckSQLiPatterns(input string) (bool, string) {
	if commentPattern.MatchString(input) {
		return true, "检测到 SQL 注释符"
	}
	return false, ""
}

func ValidateWhereClause(where string) error {
	if where == "" {
		return nil
	}

	if dangerous, keywords := CheckDangerousKeywords(where); dangerous {
		return &DangerousQueryError{Keywords: keywords}
	}

	if hasPattern, reason := CheckSQLiPatterns(where); hasPattern {
		return &DangerousQueryError{Reason: reason}
	}

	return nil
}

type DangerousQueryError struct {
	Keywords []string
	Reason   string
}

func (e *DangerousQueryError) Error() string {
	if len(e.Keywords) > 0 {
		return "检测到危险的 SQL 关键字: " + strings.Join(e.Keywords, ", ") + "，拒绝执行写操作"
	}
	if e.Reason != "" {
		return "检测到危险模式: " + e.Reason
	}
	return "危险的 SQL 查询"
}
