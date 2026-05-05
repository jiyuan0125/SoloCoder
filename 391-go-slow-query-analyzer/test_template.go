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
		"SELECT * FROM users WHERE id IN (1,2,3)",
		"SELECT * FROM users WHERE id IN (4,5,6)",
		"SELECT * FROM users2025 WHERE id = 'abc'",
		"SELECT * FROM table_1 WHERE col_2 = 123.45",
	}

	fmt.Println("=== SQL Template Test ===")
	for _, sql := range testCases {
		tmpl := t.TemplateSQL(sql)
		fmt.Printf("Original: %s\n", sql)
		fmt.Printf("Template: %s\n", tmpl)
		fmt.Println()
	}

	fmt.Println("\n=== Grouping Test ===")
	groupTest := []string{
		"SELECT * FROM table2 WHERE col1=1",
		"SELECT * FROM table2 WHERE col1=2",
		"SELECT * FROM table2 WHERE col1=100",
	}

	groups := make(map[string][]string)
	for _, sql := range groupTest {
		tmpl := t.TemplateSQL(sql)
		groups[tmpl] = append(groups[tmpl], sql)
	}

	for tmpl, sqls := range groups {
		fmt.Printf("Template: %s\n", tmpl)
		for _, sql := range sqls {
			fmt.Printf("  - %s\n", sql)
		}
		fmt.Println()
	}
}
