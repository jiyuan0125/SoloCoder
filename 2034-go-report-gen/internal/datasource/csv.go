package datasource

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"reportgen/internal/model"
)

type CSVReader struct {
	path      string
	delimiter rune
	columns   []model.Column
}

func NewCSVReader(path string, delimiter string, columns []model.Column) *CSVReader {
	r := ','
	if len(delimiter) > 0 {
		r = rune(delimiter[0])
	}
	return &CSVReader{
		path:      path,
		delimiter: r,
		columns:   columns,
	}
}

func (r *CSVReader) Read() ([]model.DataRow, error) {
	if r.path == "" {
		return nil, fmt.Errorf("data source path is empty")
	}

	fileInfo, err := os.Stat(r.path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("data source file does not exist: %s", r.path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access data source file: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("data source path is a directory, not a file: %s", r.path)
	}

	file, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = r.delimiter
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("data source file is empty: %s", r.path)
		}
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if len(r.columns) == 0 {
		for _, h := range headers {
			r.columns = append(r.columns, model.Column{
				Name: h,
				Type: model.FieldTypeString,
			})
		}
	}

	columnIndex := make(map[string]int)
	for i, h := range headers {
		columnIndex[h] = i
	}

	var rows []model.DataRow
	lineNum := 2

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row at line %d: %w", lineNum, err)
		}

		row := model.DataRow{
			Values: make(map[string]interface{}),
		}

		for _, col := range r.columns {
			idx, exists := columnIndex[col.Name]
			if !exists {
				return nil, fmt.Errorf("line %d: column '%s' not found in CSV headers", lineNum, col.Name)
			}
			if idx >= len(record) {
				return nil, fmt.Errorf("line %d: missing value for column '%s'", lineNum, col.Name)
			}

			value, err := r.convertValue(record[idx], col.Type)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			row.Values[col.Name] = value
		}

		rows = append(rows, row)
		lineNum++
	}

	return rows, nil
}

func (r *CSVReader) convertValue(str string, fieldType model.FieldType) (interface{}, error) {
	if str == "" {
		return nil, nil
	}

	switch fieldType {
	case model.FieldTypeString:
		return str, nil
	case model.FieldTypeNumber:
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number value '%s'", str)
		}
		return num, nil
	case model.FieldTypeDate:
		layouts := []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006/01/02",
			"01/02/2006",
			"2006-01-02T15:04:05Z",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, str); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("invalid date value '%s', expected format like YYYY-MM-DD", str)
	case model.FieldTypeBoolean:
		switch str {
		case "true", "True", "TRUE", "1", "yes", "Yes", "YES":
			return true, nil
		case "false", "False", "FALSE", "0", "no", "No", "NO":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid boolean value '%s'", str)
		}
	default:
		return str, nil
	}
}
