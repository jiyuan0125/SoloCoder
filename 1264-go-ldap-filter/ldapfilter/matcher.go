package ldapfilter

import (
	"regexp"
	"strings"
)

func Match(node Node, attributes map[string][]string) bool {
	if node == nil {
		return false
	}
	switch n := node.(type) {
	case *SimpleNode:
		return matchSimple(n, attributes)
	case *PresentNode:
		return matchPresent(n, attributes)
	case *SubstringNode:
		return matchSubstring(n, attributes)
	case *AndNode:
		return matchAnd(n, attributes)
	case *OrNode:
		return matchOr(n, attributes)
	case *NotNode:
		return !Match(n.Child, attributes)
	default:
		return false
	}
}

func getValues(attr string, attributes map[string][]string) []string {
	for k, v := range attributes {
		if strings.EqualFold(k, attr) {
			return v
		}
	}
	return nil
}

func matchSimple(n *SimpleNode, attributes map[string][]string) bool {
	values := getValues(n.Attribute, attributes)
	if values == nil {
		return false
	}
	pattern := n.Value
	for _, v := range values {
		switch n.Operator {
		case OpEqual:
			if strings.EqualFold(v, pattern) {
				return true
			}
		case OpApprox:
			if approxEqual(v, pattern) {
				return true
			}
		case OpGreaterEqual:
			if strings.Compare(strings.ToUpper(v), strings.ToUpper(pattern)) >= 0 {
				return true
			}
		case OpLessEqual:
			if strings.Compare(strings.ToUpper(v), strings.ToUpper(pattern)) <= 0 {
				return true
			}
		}
	}
	return false
}

func approxEqual(a, b string) bool {
	normA := normalizeApprox(a)
	normB := normalizeApprox(b)
	return normA == normB
}

func normalizeApprox(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			sb.WriteRune(toUpperASCII(r))
		}
	}
	return sb.String()
}

func toUpperASCII(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 'a' + 'A'
	}
	return r
}

func matchPresent(n *PresentNode, attributes map[string][]string) bool {
	values := getValues(n.Attribute, attributes)
	return values != nil && len(values) > 0
}

func matchSubstring(n *SubstringNode, attributes map[string][]string) bool {
	values := getValues(n.Attribute, attributes)
	if values == nil {
		return false
	}
	for _, v := range values {
		if matchSubstringValue(n, v) {
			return true
		}
	}
	return false
}

func matchSubstringValue(n *SubstringNode, value string) bool {
	upperVal := strings.ToUpper(value)
	if n.Initial != "" {
		upperInitial := strings.ToUpper(n.Initial)
		if !strings.HasPrefix(upperVal, upperInitial) {
			return false
		}
		upperVal = upperVal[len(upperInitial):]
	}
	for _, a := range n.Any {
		upperA := strings.ToUpper(a)
		idx := strings.Index(upperVal, upperA)
		if idx < 0 {
			return false
		}
		upperVal = upperVal[idx+len(upperA):]
	}
	if n.Final != "" {
		upperFinal := strings.ToUpper(n.Final)
		if !strings.HasSuffix(upperVal, upperFinal) {
			return false
		}
	}
	return true
}

func matchAnd(n *AndNode, attributes map[string][]string) bool {
	for _, c := range n.Children {
		if !Match(c, attributes) {
			return false
		}
	}
	return true
}

func matchOr(n *OrNode, attributes map[string][]string) bool {
	for _, c := range n.Children {
		if Match(c, attributes) {
			return true
		}
	}
	return false
}

func EscapeValue(s string) string {
	var sb strings.Builder
	for _, b := range []byte(s) {
		switch b {
		case '*':
			sb.WriteString("\\2a")
		case '(':
			sb.WriteString("\\28")
		case ')':
			sb.WriteString("\\29")
		case '\\':
			sb.WriteString("\\5c")
		case 0:
			sb.WriteString("\\00")
		default:
			sb.WriteByte(b)
		}
	}
	return sb.String()
}

func WildcardToRegex(pattern string) string {
	var sb strings.Builder
	sb.WriteString("(?i)^")
	for _, r := range pattern {
		switch r {
		case '*':
			sb.WriteString(".*")
		case '.', '^', '$', '+', '?', '{', '}', '[', ']', '\\', '|', '(', ')':
			sb.WriteByte('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteString("$")
	return sb.String()
}

func CompileWildcard(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(WildcardToRegex(pattern))
}
