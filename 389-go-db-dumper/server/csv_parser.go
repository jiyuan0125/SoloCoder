package server

import (
	"bufio"
	"io"
	"strings"
)

type CSVParser struct {
	skipBOM bool
}

func NewCSVParser(skipBOM bool) *CSVParser {
	return &CSVParser{
		skipBOM: skipBOM,
	}
}

func (p *CSVParser) Parse(reader io.Reader) (*CSVData, error) {
	var processedReader io.Reader = reader

	if p.skipBOM {
		processedReader = &BOMSkippingReader{
			Reader: reader,
		}
	}

	scanner := bufio.NewScanner(processedReader)
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p.parseLines(lines)
}

func (p *CSVParser) parseLines(lines []string) (*CSVData, error) {
	if len(lines) == 0 {
		return &CSVData{
			Headers: []string{},
			Rows:    [][]string{},
		}, nil
	}

	headers := p.parseRow(lines[0])
	var rows [][]string

	var currentRow string
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		currentRow += line

		if p.isRowComplete(currentRow) {
			rows = append(rows, p.parseRow(currentRow))
			currentRow = ""
		} else {
			currentRow += "\n"
		}
	}

	if currentRow != "" {
		rows = append(rows, p.parseRow(strings.TrimSuffix(currentRow, "\n")))
	}

	return &CSVData{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func (p *CSVParser) isRowComplete(row string) bool {
	inQuotes := false
	for i := 0; i < len(row); i++ {
		if row[i] == '"' {
			if i+1 < len(row) && row[i+1] == '"' {
				i++
			} else {
				inQuotes = !inQuotes
			}
		}
	}
	return !inQuotes
}

func (p *CSVParser) parseRow(row string) []string {
	var fields []string
	var currentField strings.Builder
	inQuotes := false

	for i := 0; i < len(row); i++ {
		c := row[i]

		if c == '"' {
			if inQuotes && i+1 < len(row) && row[i+1] == '"' {
				currentField.WriteByte('"')
				i++
			} else {
				inQuotes = !inQuotes
			}
			continue
		}

		if c == ',' && !inQuotes {
			fields = append(fields, currentField.String())
			currentField.Reset()
			continue
		}

		currentField.WriteByte(c)
	}

	fields = append(fields, currentField.String())
	return fields
}

type CSVData struct {
	Headers []string
	Rows    [][]string
}

type BOMSkippingReader struct {
	io.Reader
	bomChecked bool
}

func (r *BOMSkippingReader) Read(p []byte) (int, error) {
	if r.bomChecked {
		return r.Reader.Read(p)
	}

	bom := make([]byte, 3)
	n, err := r.Reader.Read(bom)
	if err != nil {
		if err == io.EOF {
			r.bomChecked = true
			return 0, io.EOF
		}
		return 0, err
	}

	if n == 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		r.bomChecked = true
		return r.Reader.Read(p)
	}

	r.bomChecked = true
	copy(p, bom[:n])
	remaining, err := r.Reader.Read(p[n:])
	return n + remaining, err
}
