//go:build ignore

package main

import (
	"fmt"

	"slowquery/internal/templater"
)

func main() {
	t := templater.NewSQLTemplater()

	testCases := []string{
		"SELECT * FROM table2 WHERE col1=1",
		"SELECT * FROM table2 WHERE col1=2",
		"SELECT * FROM table2 WHERE col1=100",
	}

	for _, sql := range testCases {
		result := t.TemplateSQL(sql)
		fmt.Printf("Input:  %s\n", sql)
		fmt.Printf("Output: %s\n", result)
		fmt.Println()
	}
}
