package csvparser

import (
	"strings"
)

type Parser struct {
	opt ParserOption
}

func NewParser(opt ParserOption) *Parser {
	return &Parser{opt: opt}
}

func Parse(text string, opt ...ParserOption) (*ParseResult, error) {
	parserOpt := ParserOption{}
	if len(opt) > 0 {
		parserOpt = opt[0]
	}
	return NewParser(parserOpt).Parse(text)
}

func (p *Parser) Parse(text string) (*ParseResult, error) {
	if text == "" {
		return &ParseResult{Records: [][]string{}}, nil
	}

	records, err := p.parseRecords(text)
	if err != nil {
		return nil, err
	}

	result := &ParseResult{Records: records}

	if p.opt.WithHeader {
		if len(records) > 0 {
			result.Headers = records[0]
			if len(records) > 1 {
				for i := 1; i < len(records); i++ {
					row := make(map[string]string)
					for j := 0; j < len(result.Headers); j++ {
						if j < len(records[i]) {
							row[result.Headers[j]] = records[i][j]
						} else {
							row[result.Headers[j]] = ""
						}
					}
					result.Rows = append(result.Rows, row)
				}
			}
		}
	}

	return result, nil
}

func (p *Parser) parseRecords(text string) ([][]string, error) {
	var records [][]string
	var currentRecord []string
	var currentField strings.Builder

	inQuotes := false
	i := 0

	for i < len(text) {
		switch {
		case text[i] == '"':
			if inQuotes {
				if i+1 < len(text) && text[i+1] == '"' {
					currentField.WriteByte('"')
					i += 2
				} else {
					inQuotes = false
					i++
				}
			} else {
				inQuotes = true
				i++
			}

		case !inQuotes && (text[i] == '\r' || text[i] == '\n'):
			currentRecord = append(currentRecord, currentField.String())
			currentField.Reset()
			records = append(records, currentRecord)
			currentRecord = nil

			if text[i] == '\r' && i+1 < len(text) && text[i+1] == '\n' {
				i += 2
			} else {
				i++
			}

		case !inQuotes && text[i] == ',':
			currentRecord = append(currentRecord, currentField.String())
			currentField.Reset()
			i++

		default:
			currentField.WriteByte(text[i])
			i++
		}
	}

	if currentField.Len() > 0 || len(currentRecord) > 0 {
		currentRecord = append(currentRecord, currentField.String())
		records = append(records, currentRecord)
	}

	return records, nil
}
