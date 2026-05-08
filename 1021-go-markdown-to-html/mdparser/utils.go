package mdparser

import (
	"regexp"
	"strings"
)

var htmlEscapeReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"'", "&#39;",
)

func escapeHTML(s string) string {
	return htmlEscapeReplacer.Replace(s)
}

func trimSpaces(s string) string {
	return strings.TrimSpace(s)
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

var escapeChars = map[rune]bool{
	'\\': true, '`': true, '*': true, '_': true, '{': true, '}': true,
	'[': true, ']': true, '(': true, ')': true, '#': true, '+': true,
	'-': true, '.': true, '!': true, '|': true, '>': true, '~': true,
}

func processEscapeChars(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) && escapeChars[runes[i+1]] {
			result.WriteRune(runes[i+1])
			i++
		} else {
			result.WriteRune(runes[i])
		}
	}
	return result.String()
}

var headingRegex = regexp.MustCompile("^(#{1,6})\\s+(.*)$")
var ulRegex = regexp.MustCompile("^(\\s*)[-*+]\\s+(.*)$")
var olRegex = regexp.MustCompile("^(\\s*)(\\d+)\\.\\s+(.*)$")
var checkboxRegex = regexp.MustCompile("^\\[([ xX])\\]\\s+(.*)$")
var fenceRegex = regexp.MustCompile("^```\\s*([\\w-]*)\\s*$")
var quoteRegex = regexp.MustCompile("^>\\s?(.*)$")
var tableRowRegex = regexp.MustCompile("^([^\\n]*\\|[^\\n]*)$")
var tableSepRegex = regexp.MustCompile("^[\\s\\-:|]+$")

func isHeading(line string) (int, string) {
	matches := headingRegex.FindStringSubmatch(line)
	if matches != nil {
		level := len(matches[1])
		content := trimSpaces(matches[2])
		return level, content
	}
	return 0, ""
}

func isUnorderedList(line string) (int, string) {
	matches := ulRegex.FindStringSubmatch(line)
	if matches != nil {
		indent := countLeadingSpaces(matches[1])
		content := matches[2]
		return indent, content
	}
	return -1, ""
}

func isOrderedList(line string) (int, int, string) {
	matches := olRegex.FindStringSubmatch(line)
	if matches != nil {
		indent := countLeadingSpaces(matches[1])
		num := 0
		for _, ch := range matches[2] {
			num = num*10 + int(ch-'0')
		}
		content := matches[3]
		return indent, num, content
	}
	return -1, 0, ""
}

func isCheckbox(content string) (bool, bool, string) {
	matches := checkboxRegex.FindStringSubmatch(content)
	if matches != nil {
		checked := matches[1] == "x" || matches[1] == "X"
		return true, checked, matches[2]
	}
	return false, false, content
}

func isFence(line string) (bool, string) {
	matches := fenceRegex.FindStringSubmatch(line)
	if matches != nil {
		return true, matches[1]
	}
	return false, ""
}

func isQuote(line string) (bool, string) {
	matches := quoteRegex.FindStringSubmatch(line)
	if matches != nil {
		return true, matches[1]
	}
	return false, ""
}

func isBlank(line string) bool {
	return trimSpaces(line) == ""
}
