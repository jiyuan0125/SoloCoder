package csvutil

import (
	"bytes"
	"encoding/csv"
	"os"
	"strings"
)

func ReadCSV(filePath string, keyColumn string) (*CSVData, []string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	
	data = normalizeLineEndings(data)
	
	reader := csv.NewReader(bytes.NewReader(data))
	reader.LazyQuotes = true
	
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	
	if len(records) == 0 {
		return &CSVData{
			Headers: []string{},
			Records: make(map[string]CSVRecord),
		}, []string{}, nil
	}
	
	headers := records[0]
	keyIndex := -1
	
	if keyColumn == "" {
		keyIndex = 0
	} else {
		for i, h := range headers {
			if strings.TrimSpace(h) == strings.TrimSpace(keyColumn) {
				keyIndex = i
				break
			}
		}
	}
	
	if keyIndex == -1 {
		return nil, nil, nil
	}
	
	csvData := &CSVData{
		Headers: headers,
		Records: make(map[string]CSVRecord),
	}
	
	var warnings []string
	seenKeys := make(map[string]bool)
	
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 {
			continue
		}
		
		record := CSVRecord{
			Values: make(map[string]string),
		}
		
		for j, h := range headers {
			if j < len(row) {
				record.Values[h] = row[j]
			} else {
				record.Values[h] = ""
			}
		}
		
		if keyIndex < len(row) {
			key := strings.TrimSpace(row[keyIndex])
			if key == "" {
				continue
			}
			
			if seenKeys[key] {
				warnings = append(warnings, "Duplicate key found: "+key+", keeping first occurrence")
				continue
			}
			
			seenKeys[key] = true
			csvData.Records[key] = record
		}
	}
	
	return csvData, warnings, nil
}

func normalizeLineEndings(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	return data
}

func GetFirstColumnHeader(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	data = normalizeLineEndings(data)
	
	reader := csv.NewReader(bytes.NewReader(data))
	reader.LazyQuotes = true
	
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}
	
	if len(records) == 0 || len(records[0]) == 0 {
		return "", nil
	}
	
	return strings.TrimSpace(records[0][0]), nil
}
