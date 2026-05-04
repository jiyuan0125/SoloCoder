package exporter

import (
	"io"
)

type ExportConfig struct {
	Fields       []string
	FieldMapping map[string]string
}

func ExportCSV(w io.Writer, data []map[string]string, config ExportConfig) error {
	allKeys := collectAllKeys(data, config.Fields)
	if len(allKeys) == 0 {
		return nil
	}

	headerMapped := make([]string, len(allKeys))
	for i, key := range allKeys {
		if mapped, ok := config.FieldMapping[key]; ok {
			headerMapped[i] = mapped
		} else {
			headerMapped[i] = key
		}
	}

	if err := writeCSVRow(w, headerMapped); err != nil {
		return err
	}

	for _, row := range data {
		rowValues := make([]string, len(allKeys))
		for i, key := range allKeys {
			rowValues[i] = row[key]
		}
		if err := writeCSVRow(w, rowValues); err != nil {
			return err
		}
	}

	return nil
}

func ExportJSON(w io.Writer, data []map[string]string, config ExportConfig) error {
	return exportJSONArray(w, data, config)
}

func ExportJSONLines(w io.Writer, data []map[string]string, config ExportConfig) error {
	allKeys := collectAllKeys(data, config.Fields)
	if len(allKeys) == 0 {
		return nil
	}

	for _, row := range data {
		filteredRow := make(map[string]string)
		for _, key := range allKeys {
			if val, ok := row[key]; ok {
				filteredRow[key] = val
			} else {
				filteredRow[key] = ""
			}
		}

		mappedRow := make(map[string]string)
		for key, val := range filteredRow {
			if mappedKey, ok := config.FieldMapping[key]; ok {
				mappedRow[mappedKey] = val
			} else {
				mappedRow[key] = val
			}
		}

		if err := writeJSONObjectLine(w, mappedRow); err != nil {
			return err
		}
	}

	return nil
}

func collectAllKeys(data []map[string]string, specifiedFields []string) []string {
	if len(specifiedFields) > 0 {
		return specifiedFields
	}

	keyMap := make(map[string]struct{})
	for _, row := range data {
		for key := range row {
			keyMap[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(keyMap))
	for key := range keyMap {
		keys = append(keys, key)
	}
	return keys
}
