package tableparser

import (
	"strings"
	"unicode"
)

func ExtractTables(markdown string) []*Table {
	var tables []*Table
	lines := strings.Split(markdown, "\n")

	for i := 0; i < len(lines); i++ {
		if isTableRowLine(lines[i]) {
			table, nextIdx := parseTable(lines, i)
			if table != nil {
				tables = append(tables, table)
			}
			i = nextIdx
		}
	}

	return tables
}

func parseTable(lines []string, startIdx int) (*Table, int) {
	var tableLines []string

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		if isTableRowLine(line) || isSeparatorLine(line) {
			tableLines = append(tableLines, line)
		} else if len(tableLines) > 0 {
			break
		}
	}

	if len(tableLines) < 2 {
		return nil, startIdx
	}

	separatorIdx := -1
	for i, line := range tableLines {
		if isSeparatorLine(line) {
			separatorIdx = i
			break
		}
	}

	if separatorIdx == -1 || separatorIdx == 0 {
		return nil, startIdx
	}

	headerLine := tableLines[separatorIdx-1]
	separatorLine := tableLines[separatorIdx]
	dataLines := tableLines[separatorIdx+1:]

	headers := parseTableRow(headerLine)
	if len(headers) == 0 {
		return nil, startIdx
	}

	alignments := parseSeparatorLine(separatorLine)
	if len(alignments) == 0 {
		return nil, startIdx
	}

	if len(alignments) < len(headers) {
		for len(alignments) < len(headers) {
			alignments = append(alignments, AlignmentLeft)
		}
	}

	var rows [][]string
	for _, dataLine := range dataLines {
		row := parseTableRow(dataLine)
		if len(row) == 0 {
			continue
		}
		for len(row) < len(headers) {
			row = append(row, "")
		}
		rows = append(rows, row[:len(headers)])
	}

	return NewTable(headers, rows, alignments), startIdx + len(tableLines)
}

func isTableRowLine(line string) bool {
	if strings.TrimSpace(line) == "" {
		return false
	}
	return strings.Contains(line, "|")
}

func isSeparatorLine(line string) bool {
	if !isTableRowLine(line) {
		return false
	}

	cells := splitLineByPipe(line)
	if len(cells) == 0 {
		return false
	}

	for _, cell := range cells {
		trimmed := strings.TrimSpace(cell)
		if trimmed == "" {
			continue
		}
		if !isSeparatorCell(trimmed) {
			return false
		}
	}

	return true
}

func isSeparatorCell(cell string) bool {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return false
	}

	hasDashes := false

	for i, r := range cell {
		if i == 0 && r == ':' {
			continue
		}
		if i == len(cell)-1 && r == ':' {
			continue
		}
		if r == '-' {
			hasDashes = true
			continue
		}
		return false
	}

	return hasDashes
}

func splitLineByPipe(line string) []string {
	line = strings.TrimSpace(line)

	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	var cells []string
	var currentCell strings.Builder
	escaped := false

	for _, r := range line {
		if escaped {
			currentCell.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '|' {
			cells = append(cells, currentCell.String())
			currentCell.Reset()
			continue
		}
		currentCell.WriteRune(r)
	}

	if currentCell.Len() > 0 {
		cells = append(cells, currentCell.String())
	}

	return cells
}

func parseTableRow(line string) []string {
	rawCells := splitLineByPipe(line)
	var cells []string
	for _, rawCell := range rawCells {
		cell := strings.TrimFunc(rawCell, func(r rune) bool {
			return unicode.IsSpace(r)
		})
		cells = append(cells, cell)
	}
	return cells
}

func parseSeparatorLine(line string) []Alignment {
	rawCells := splitLineByPipe(line)
	var alignments []Alignment

	for _, rawCell := range rawCells {
		cell := strings.TrimSpace(rawCell)
		if cell == "" {
			continue
		}

		hasLeftColon := len(cell) > 0 && cell[0] == ':'
		hasRightColon := len(cell) > 0 && cell[len(cell)-1] == ':'

		var alignment Alignment
		if hasLeftColon && hasRightColon {
			alignment = AlignmentCenter
		} else if hasRightColon {
			alignment = AlignmentRight
		} else {
			alignment = AlignmentLeft
		}
		alignments = append(alignments, alignment)
	}

	return alignments
}
