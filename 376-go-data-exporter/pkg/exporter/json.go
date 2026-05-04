package exporter

import (
	"encoding/json"
	"io"
)

func exportJSONArray(w io.Writer, data []map[string]string, config ExportConfig) error {
	allKeys := collectAllKeys(data, config.Fields)
	if len(allKeys) == 0 {
		if _, err := w.Write([]byte("[]")); err != nil {
			return err
		}
		return nil
	}

	filteredData := make([]map[string]string, 0, len(data))
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
		filteredData = append(filteredData, mappedRow)
	}

	encoded, err := json.Marshal(filteredData)
	if err != nil {
		return err
	}
	if _, err := w.Write(encoded); err != nil {
		return err
	}
	return nil
}

func writeJSONObjectLine(w io.Writer, obj map[string]string) error {
	encoded, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if _, err := w.Write(encoded); err != nil {
		return err
	}
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}
