//go:build ignore

package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	testCases := []string{
		"SELECT * FROM table2 WHERE col1=1",
		"SELECT * FROM table2 WHERE col1=100",
	}

	for _, sql := range testCases {
		fmt.Printf("=== Testing: %s ===\n", sql)
		
		re := regexp.MustCompile(`(\b[+-]?\d+(\.\d+)?(e[+-]?\d+)?\b)`)
		matches := re.FindAllStringIndex(sql, -1)
		
		for i, match := range matches {
			start, end := match[0], match[1]
			matchStr := sql[start:end]
			
			fmt.Printf("Match %d: [%d:%d] '%s'\n", i, start, end, matchStr)
			
			if start > 0 {
				prevChar := sql[start-1]
				isIdentifierContext := (prevChar >= 'a' && prevChar <= 'z') || 
					(prevChar >= 'A' && prevChar <= 'Z') || 
					prevChar == '_' || prevChar == '$'
				fmt.Printf("  prev char: '%c' (0x%02X), isIdentifierContext: %v\n", prevChar, prevChar, isIdentifierContext)
			}
			
			if end < len(sql) {
				nextChar := sql[end]
				isIdentifierContext := (nextChar >= 'a' && nextChar <= 'z') || 
					(nextChar >= 'A' && nextChar <= 'Z') || 
					nextChar == '_' || nextChar == '$'
				fmt.Printf("  next char: '%c' (0x%02X), isIdentifierContext: %v\n", nextChar, nextChar, isIdentifierContext)
			}
			
			fmt.Printf("  Should replace: %v\n", !isNumericInContextOfIdentifierAt(sql, start, end))
		}
		
		result := replaceNumericValues(sql)
		fmt.Printf("Result: %s\n\n", result)
	}
}

func isNumericInContextOfIdentifierAt(sql string, start, end int) bool {
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

func replaceNumericValues(sql string) string {
	re := regexp.MustCompile(`(\b[+-]?\d+(\.\d+)?(e[+-]?\d+)?\b)`)
	
	matches := re.FindAllStringIndex(sql, -1)
	if len(matches) == 0 {
		return sql
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]
		matchStr := sql[start:end]

		result.WriteString(sql[lastEnd:start])

		if isNumericInContextOfIdentifierAt(sql, start, end) {
			result.WriteString(matchStr)
		} else {
			result.WriteString("?")
		}

		lastEnd = end
	}

	result.WriteString(sql[lastEnd:])
	return result.String()
}
