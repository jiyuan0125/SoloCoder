package tableparser

import (
	"strings"
	"unicode"
)

func (t *Table) ToMarkdown() string {
	var builder strings.Builder
	columnCount := len(t.Headers)
	if columnCount == 0 {
		return ""
	}

	columnWidths := t.calculateColumnWidths()

	builder.WriteString("| ")
	for i, header := range t.Headers {
		padded := padString(header, columnWidths[i], t.Alignments[i])
		builder.WriteString(padded)
		if i < columnCount-1 {
			builder.WriteString(" | ")
		}
	}
	builder.WriteString(" |\n")

	builder.WriteString("| ")
	for i := 0; i < columnCount; i++ {
		builder.WriteString(t.createSeparatorCell(columnWidths[i], t.Alignments[i]))
		if i < columnCount-1 {
			builder.WriteString(" | ")
		}
	}
	builder.WriteString(" |\n")

	for _, row := range t.Rows {
		builder.WriteString("| ")
		for i := 0; i < columnCount; i++ {
			cell := ""
			if i < len(row) {
				cell = escapeCellContent(row[i])
			}
			padded := padString(cell, columnWidths[i], t.Alignments[i])
			builder.WriteString(padded)
			if i < columnCount-1 {
				builder.WriteString(" | ")
			}
		}
		builder.WriteString(" |\n")
	}

	return builder.String()
}

func (t *Table) calculateColumnWidths() []int {
	columnCount := len(t.Headers)
	widths := make([]int, columnCount)

	for i, header := range t.Headers {
		widths[i] = len(header)
	}

	for _, row := range t.Rows {
		for i := 0; i < columnCount && i < len(row); i++ {
			cellLen := len(escapeCellContent(row[i]))
			if cellLen > widths[i] {
				widths[i] = cellLen
			}
		}
	}

	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
	}

	return widths
}

func (t *Table) createSeparatorCell(width int, alignment Alignment) string {
	var builder strings.Builder

	if width < 3 {
		width = 3
	}

	switch alignment {
	case AlignmentLeft:
		builder.WriteString(":")
		for i := 0; i < width-1; i++ {
			builder.WriteString("-")
		}
	case AlignmentRight:
		for i := 0; i < width-1; i++ {
			builder.WriteString("-")
		}
		builder.WriteString(":")
	case AlignmentCenter:
		builder.WriteString(":")
		for i := 0; i < width-2; i++ {
			builder.WriteString("-")
		}
		builder.WriteString(":")
	default:
		for i := 0; i < width; i++ {
			builder.WriteString("-")
		}
	}

	return builder.String()
}

func escapeCellContent(content string) string {
	return strings.ReplaceAll(content, "|", "\\|")
}

func padString(s string, width int, alignment Alignment) string {
	if len(s) >= width {
		return s
	}

	padding := width - len(s)

	switch alignment {
	case AlignmentLeft:
		return s + strings.Repeat(" ", padding)
	case AlignmentRight:
		return strings.Repeat(" ", padding) + s
	case AlignmentCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
	default:
		return s + strings.Repeat(" ", padding)
	}
}

func (t *Table) InferAlignments() {
	if len(t.Headers) == 0 {
		return
	}

	columnCount := len(t.Headers)
	t.Alignments = make([]Alignment, columnCount)

	for col := 0; col < columnCount; col++ {
		if isNumericColumn(t, col) {
			t.Alignments[col] = AlignmentRight
		} else {
			t.Alignments[col] = AlignmentLeft
		}
	}
}

func isNumericColumn(t *Table, col int) bool {
	numericCount := 0
	totalCount := 0

	for _, row := range t.Rows {
		if col < len(row) {
			totalCount++
			if isNumeric(row[col]) {
				numericCount++
			}
		}
	}

	if totalCount == 0 {
		return false
	}

	return float64(numericCount)/float64(totalCount) >= 0.7
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	hasDigits := false
	hasDot := false

	for i, r := range s {
		if i == 0 && (r == '+' || r == '-') {
			continue
		}
		if r == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if unicode.IsDigit(r) {
			hasDigits = true
			continue
		}
		return false
	}

	return hasDigits
}

func NewTableFromCSV(csvText string) (*Table, error) {
	table, err := ParseCSV(csvText)
	if err != nil {
		return nil, err
	}
	table.InferAlignments()
	return table, nil
}
