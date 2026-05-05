package csvproc

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strings"
)

type CSVFile struct {
	Header []string
	Rows   [][]string
	Path   string
}

func ReadCSVFile(filePath string) (*CSVFile, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	encoding := DetectEncoding(data)
	utf8Data, err := ConvertToUTF8(data, encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to convert encoding for %s: %w", filePath, err)
	}

	utf8Data = bytes.ReplaceAll(utf8Data, []byte("\r\n"), []byte("\n"))

	reader := csv.NewReader(bytes.NewReader(utf8Data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV %s: %w", filePath, err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("file %s is empty", filePath)
	}

	header := records[0]
	var rows [][]string
	if len(records) > 1 {
		rows = records[1:]
	}

	return &CSVFile{
		Header: header,
		Rows:   rows,
		Path:   filePath,
	}, nil
}

func ValidateHeaders(files []*CSVFile) error {
	if len(files) == 0 {
		return nil
	}

	referenceHeader := files[0].Header
	referenceMap := make(map[string]int)
	for i, col := range referenceHeader {
		referenceMap[col] = i
	}

	for _, file := range files[1:] {
		if len(file.Header) != len(referenceHeader) {
			return fmt.Errorf("file %s has different column count: expected %d, got %d",
				file.Path, len(referenceHeader), len(file.Header))
		}

		var diffCols []string
		for i, col := range file.Header {
			if referenceMap[col] != i {
				diffCols = append(diffCols, col)
			}
		}

		if len(diffCols) > 0 {
			return fmt.Errorf("file %s has different columns: %s",
				file.Path, strings.Join(diffCols, ", "))
		}
	}

	return nil
}
