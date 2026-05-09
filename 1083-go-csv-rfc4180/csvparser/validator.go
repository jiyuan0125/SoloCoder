package csvparser

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Line    int
	Column  int
	Message string
}

func (e *ValidationError) String() string {
	return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Message)
}

func Validate(text string, opt ...ParserOption) []*ValidationError {
	parserOpt := ParserOption{}
	if len(opt) > 0 {
		parserOpt = opt[0]
	}
	validator := &validator{
		withHeader: parserOpt.WithHeader,
	}
	return validator.validate(text)
}

type validator struct {
	withHeader    bool
	errors        []*ValidationError
	headerColumnCount int
}

func (v *validator) validate(text string) []*ValidationError {
	if text == "" {
		return []*ValidationError{}
	}

	inQuotes := false
	i := 0
	line := 1
	column := 1
	columnCount := 0
	fieldStart := 0

	for i < len(text) {
		switch text[i] {
		case '"':
			if !inQuotes {
				if i != fieldStart {
					v.addError(line, column, "unquoted field contains quote")
				}
				inQuotes = true
			} else {
				if i+1 < len(text) && text[i+1] == '"' {
					i++
					column++
				} else {
					inQuotes = false
					if i+1 < len(text) && !strings.ContainsRune(",\r\n", rune(text[i+1])) {
						v.addError(line, column+1, "unexpected character after closing quote")
					}
				}
			}

		case ',':
			if !inQuotes {
				columnCount++
				fieldStart = i + 1
			}

		case '\r':
			if !inQuotes {
				columnCount++
				v.checkLineLength(line, columnCount)
				if i+1 < len(text) && text[i+1] == '\n' {
					i++
				}
				line++
				column = 0
				columnCount = 0
				fieldStart = i + 1
			}

		case '\n':
			if !inQuotes {
				columnCount++
				v.checkLineLength(line, columnCount)
				line++
				column = 0
				columnCount = 0
				fieldStart = i + 1
			}
		}

		i++
		column++
	}

	columnCount++
	v.checkLineLength(line, columnCount)

	if inQuotes {
		v.addError(line, column, "unclosed quoted field")
	}

	return v.errors
}

func (v *validator) checkLineLength(line int, count int) {
	if v.withHeader {
		if line == 1 {
			v.headerColumnCount = count
		} else {
			if count != v.headerColumnCount {
				v.addError(line, 1, fmt.Sprintf("column count mismatch: expected %d, got %d", v.headerColumnCount, count))
			}
		}
	}
}

func (v *validator) addError(line int, column int, message string) {
	v.errors = append(v.errors, &ValidationError{
		Line:    line,
		Column:  column,
		Message: message,
	})
}
