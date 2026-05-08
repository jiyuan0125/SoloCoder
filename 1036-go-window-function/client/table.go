package main

import (
	"fmt"
	"strings"
	"windowfunc/api"
)

func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprintf("%v", v)
}

func getMaxWidths(rows []api.Record, cols []string) []int {
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = len(col)
	}
	for _, row := range rows {
		for i, col := range cols {
			w := len(formatValue(row[col]))
			if w > widths[i] {
				widths[i] = w
			}
		}
	}
	for i := range widths {
		if widths[i] < 4 {
			widths[i] = 4
		}
	}
	return widths
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func getColumns(rows []api.Record) []string {
	if len(rows) == 0 {
		return nil
	}
	colsMap := make(map[string]bool)
	order := make([]string, 0)
	for _, row := range rows {
		for k := range row {
			if !colsMap[k] {
				colsMap[k] = true
				order = append(order, k)
			}
		}
	}
	return order
}

func printTable(rows []api.Record) {
	if len(rows) == 0 {
		fmt.Println("(no data)")
		return
	}
	cols := getColumns(rows)
	widths := getMaxWidths(rows, cols)
	sep := "+"
	for _, w := range widths {
		sep += strings.Repeat("-", w+2) + "+"
	}
	fmt.Println(sep)
	header := "|"
	for i, col := range cols {
		header += " " + padRight(col, widths[i]) + " |"
	}
	fmt.Println(header)
	fmt.Println(sep)
	for _, row := range rows {
		line := "|"
		for i, col := range cols {
			line += " " + padRight(formatValue(row[col]), widths[i]) + " |"
		}
		fmt.Println(line)
	}
	fmt.Println(sep)
}
