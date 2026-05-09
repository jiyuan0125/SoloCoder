package csvparser

type ParserOption struct {
	WithHeader bool
}

type ParseResult struct {
	Records [][]string
	Headers []string
	Rows    []map[string]string
}
