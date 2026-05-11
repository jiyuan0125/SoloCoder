package regexgen

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Config struct {
	ShortestMatch bool
	AllowOptional bool
}

type Result struct {
	Regex       string
	Explanation string
	Examples    []string
}

func Generate(examples []string, config Config) (*Result, error) {
	if len(examples) == 0 {
		return nil, fmt.Errorf("no examples provided")
	}

	filtered := make([]string, 0, len(examples))
	seen := make(map[string]bool)
	for _, ex := range examples {
		if !seen[ex] {
			seen[ex] = true
			filtered = append(filtered, ex)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("all examples are empty")
	}

	if len(filtered) == 1 {
		return generateFromSingle(filtered[0])
	}

	return generateSmart(filtered, config)
}

func generateFromSingle(s string) (*Result, error) {
	pattern, explanation := analyzeSingle(s)
	return &Result{
		Regex:       pattern,
		Explanation: explanation,
		Examples:    []string{s},
	}, nil
}

func analyzeSingle(s string) (string, string) {
	var pattern strings.Builder
	var explanation strings.Builder
	
	i := 0
	for i < len(s) {
		c := rune(s[i])
		switch {
		case unicode.IsDigit(c):
			end := i
			for end < len(s) && unicode.IsDigit(rune(s[end])) {
				end++
			}
			count := end - i
			if count == 1 {
				pattern.WriteString("\\d")
			} else {
				pattern.WriteString(fmt.Sprintf("\\d{%d}", count))
			}
			if explanation.Len() > 0 {
				explanation.WriteString("，")
			}
			explanation.WriteString(fmt.Sprintf("%d位数字", count))
			i = end
		case unicode.IsLetter(c):
			end := i
			for end < len(s) && unicode.IsLetter(rune(s[end])) {
				end++
			}
			count := end - i
			if count == 1 {
				pattern.WriteString("[a-zA-Z]")
			} else {
				pattern.WriteString(fmt.Sprintf("[a-zA-Z]{%d}", count))
			}
			if explanation.Len() > 0 {
				explanation.WriteString("，")
			}
			explanation.WriteString(fmt.Sprintf("%d位字母", count))
			i = end
		case c == '_' || unicode.IsDigit(c) || unicode.IsLetter(c):
			end := i
			for end < len(s) && (rune(s[end]) == '_' || unicode.IsDigit(rune(s[end])) || unicode.IsLetter(rune(s[end]))) {
				end++
			}
			count := end - i
			if count == 1 {
				pattern.WriteString("\\w")
			} else {
				pattern.WriteString(fmt.Sprintf("\\w{%d}", count))
			}
			if explanation.Len() > 0 {
				explanation.WriteString("，")
			}
			explanation.WriteString(fmt.Sprintf("%d位单词字符", count))
			i = end
		default:
			pattern.WriteString(escapeRune(c))
			if explanation.Len() > 0 {
				explanation.WriteString("，")
			}
			explanation.WriteString(fmt.Sprintf("字面量 '%c'", c))
			i++
		}
	}
	
	return pattern.String(), explanation.String()
}

func escapeRune(c rune) string {
	special := map[rune]bool{
		'.': true, '+': true, '*': true, '?': true,
		'(': true, ')': true, '[': true, ']': true,
		'{': true, '}': true, '^': true, '$': true,
		'\\': true, '|': true,
	}
	if special[c] {
		return "\\" + string(c)
	}
	return string(c)
}

func generateSmart(examples []string, config Config) (*Result, error) {
	prefix := findLongestCommonPrefix(examples)
	suffix := findLongestCommonSuffix(examples)
	
	prefixLen := len(prefix)
	suffixLen := len(suffix)
	
	var middleParts []string
	for _, ex := range examples {
		if len(ex) > prefixLen+suffixLen {
			middle := ex[prefixLen : len(ex)-suffixLen]
			middleParts = append(middleParts, middle)
		} else {
			middleParts = append(middleParts, "")
		}
	}
	
	var regex strings.Builder
	var explanation strings.Builder
	
	if prefixLen > 0 {
		p, e := analyzeSingle(prefix)
		regex.WriteString(p)
		explanation.WriteString(e)
	}
	
	if len(middleParts) > 0 {
		middleRegex, middleExp := analyzeMiddle(middleParts)
		if middleRegex != "" {
			if regex.Len() > 0 {
				explanation.WriteString("，")
			}
			regex.WriteString(middleRegex)
			explanation.WriteString(middleExp)
		}
	}
	
	if suffixLen > 0 {
		p, e := analyzeSingle(suffix)
		if regex.Len() > 0 {
			explanation.WriteString("，")
		}
		regex.WriteString(p)
		explanation.WriteString(e)
	}
	
	if regex.Len() == 0 {
		regex.WriteString(".*")
		explanation.WriteString("任意字符（无固定模式）")
	}
	
	return &Result{
		Regex:       regex.String(),
		Explanation: explanation.String(),
		Examples:    examples,
	}, nil
}

func analyzeMiddle(parts []string) (string, string) {
	if len(parts) == 0 {
		return "", ""
	}
	
	allEmpty := true
	for _, p := range parts {
		if p != "" {
			allEmpty = false
			break
		}
	}
	if allEmpty {
		return "", ""
	}
	
	allDigits := true
	allNonEmptyDigits := true
	for _, p := range parts {
		if p == "" {
			allNonEmptyDigits = false
			continue
		}
		for _, c := range p {
			if !unicode.IsDigit(c) {
				allDigits = false
				allNonEmptyDigits = false
				break
			}
		}
		if !allDigits {
			break
		}
	}
	
	if allDigits {
		minLen := -1
		for _, p := range parts {
			if minLen == -1 || len(p) < minLen {
				minLen = len(p)
			}
		}
		
		if minLen == 0 {
			return "\\d*", "任意数量的数字"
		}
		if minLen == 1 {
			return "\\d+", "至少1位数字"
		}
		return fmt.Sprintf("\\d{%d,}", minLen), fmt.Sprintf("至少%d位数字", minLen)
	}
	
	if allNonEmptyDigits {
		minLen := -1
		for _, p := range parts {
			if len(p) > 0 && (minLen == -1 || len(p) < minLen) {
				minLen = len(p)
			}
		}
		
		if minLen == 1 {
			return "\\d*", "任意数量的数字（可空）"
		}
		return fmt.Sprintf("\\d{%d,}", minLen), fmt.Sprintf("至少%d位数字（可空）", minLen)
	}
	
	allWordChars := true
	for _, p := range parts {
		for _, c := range p {
			if c != '_' && !unicode.IsDigit(c) && !unicode.IsLetter(c) {
				allWordChars = false
				break
			}
		}
		if !allWordChars {
			break
		}
	}
	
	if allWordChars {
		minLen := -1
		for _, p := range parts {
			if minLen == -1 || len(p) < minLen {
				minLen = len(p)
			}
		}
		
		if minLen == 0 {
			return "\\w*", "任意数量的单词字符"
		}
		if minLen == 1 {
			return "\\w+", "至少1位单词字符"
		}
		return fmt.Sprintf("\\w{%d,}", minLen), fmt.Sprintf("至少%d位单词字符", minLen)
	}
	
	minLen := -1
	for _, p := range parts {
		if minLen == -1 || len(p) < minLen {
			minLen = len(p)
		}
	}
	
	if minLen == 0 {
		return ".*?", "任意数量的任意字符（非贪婪匹配）"
	}
	if minLen == 1 {
		return ".+?", "至少1位任意字符（非贪婪匹配）"
	}
	return fmt.Sprintf(".{%d,}?", minLen), fmt.Sprintf("至少%d位任意字符（非贪婪匹配）", minLen)
}

func findLongestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) && len(prefix) > 0 {
			prefix = prefix[:len(prefix)-1]
		}
		if len(prefix) == 0 {
			break
		}
	}
	return prefix
}

func findLongestCommonSuffix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	suffix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasSuffix(s, suffix) && len(suffix) > 0 {
			suffix = suffix[1:]
		}
		if len(suffix) == 0 {
			break
		}
	}
	return suffix
}

func Validate(regex string, testStr string) (bool, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return false, err
	}
	return re.MatchString(testStr), nil
}
