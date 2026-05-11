package tableparser

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

func (t *Table) ToCSV() (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if len(t.Headers) > 0 {
		if err := w.Write(t.Headers); err != nil {
			return "", err
		}
	}

	for _, row := range t.Rows {
		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ParseCSV(csvText string) (*Table, error) {
	r := csv.NewReader(strings.NewReader(csvText))
	r.LazyQuotes = true

	var headers []string
	var rows [][]string
	isFirstRow := true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if isFirstRow {
			headers = record
			isFirstRow = false
		} else {
			rows = append(rows, record)
		}
	}

	return NewTable(headers, rows, nil), nil
}

func ParseCSVWithOptions(csvText string, hasHeader bool) (*Table, error) {
	r := csv.NewReader(strings.NewReader(csvText))
	r.LazyQuotes = true

	var headers []string
	var rows [][]string
	isFirstRow := true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if isFirstRow {
			if hasHeader {
				headers = record
			} else {
				headers = make([]string, len(record))
				for i := range headers {
					headers[i] = "Col" + string(rune('A'+i))
				}
				rows = append(rows, record)
			}
			isFirstRow = false
		} else {
			rows = append(rows, record)
		}
	}

	return NewTable(headers, rows, nil), nil
}
