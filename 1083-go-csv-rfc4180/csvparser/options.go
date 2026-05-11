package csvparser

import "errors"

var ErrUnclosedQuote = errors.New("unclosed quoted field")

type ParserOption struct {
	WithHeader bool
}

type ParseResult struct {
	Records [][]string
	Headers []string
	Rows    []map[string]string
}
