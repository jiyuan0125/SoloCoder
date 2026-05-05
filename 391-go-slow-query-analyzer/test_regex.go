//go:build ignore

package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`(\b[+-]?\d+(\.\d+)?(e[+-]?\d+)?\b)`)

	testCases := []string{
		"SELECT * FROM table2 WHERE col1=1",
		"SELECT * FROM table2 WHERE col1=100",
		"SELECT * FROM table2 WHERE col1 = 1",
		"SELECT * FROM table2 WHERE col1= 1",
		"SELECT * FROM table2 WHERE col1 =1",
	}

	for _, sql := range testCases {
		fmt.Printf("SQL: %s\n", sql)
		matches := re.FindAllStringIndex(sql, -1)
		for i, match := range matches {
			start, end := match[0], match[1]
			matchStr := sql[start:end]
			fmt.Printf("  Match %d: [%d:%d] '%s'\n", i, start, end, matchStr)
			if start > 0 {
				fmt.Printf("    prev char: '%c' (0x%02X)\n", sql[start-1], sql[start-1])
			}
			if end < len(sql) {
				fmt.Printf("    next char: '%c' (0x%02X)\n", sql[end], sql[end])
			}
		}
		fmt.Println()
	}
}
