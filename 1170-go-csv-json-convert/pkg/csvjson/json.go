package csvjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

func flattenValue(prefix string, value interface{}, result map[string]string) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "." + key
			}
			flattenValue(newPrefix, val, result)
		}
	case []interface{}:
		for i, val := range v {
			newPrefix := prefix + "[" + strconv.Itoa(i) + "]"
			flattenValue(newPrefix, val, result)
		}
	default:
		if prefix == "" {
			return
		}
		result[prefix] = valueToString(value)
	}
}

func valueToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case float64:
		if float64(int64(val)) == val {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		if float32(int32(val)) == val {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case json.Number:
		return val.String()
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func escapeCSVField(value string) string {
	if strings.ContainsAny(value, "\"\n\r,") || strings.Contains(value, "\t") {
		escaped := strings.ReplaceAll(value, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return value
}

func writeCSVLine(writer io.Writer, fields []string, delimiter Delimiter) error {
	w := bufio.NewWriter(writer)
	for i, field := range fields {
		if i > 0 {
			if _, err := w.WriteRune(rune(delimiter)); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(escapeCSVField(field)); err != nil {
			return err
		}
	}
	if _, err := w.WriteString("\n"); err != nil {
		return err
	}
	return w.Flush()
}

func JSONToCSV(input io.Reader, output io.Writer, opts *ConvertOptions) error {
	if opts == nil {
		opts = DefaultOptions()
	}
	decoder := json.NewDecoder(input)
	var records []map[string]interface{}
	for {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			var arr []map[string]interface{}
			if seeker, ok := input.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
				decoder = json.NewDecoder(input)
				if err := decoder.Decode(&arr); err != nil {
					return fmt.Errorf("failed to parse JSON: %w", err)
				}
				records = arr
				break
			}
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		records = append(records, obj)
	}
	if len(records) == 0 {
		return nil
	}
	allHeaders := make(map[string]bool)
	flattenedRecords := make([]map[string]string, 0, len(records))
	for _, record := range records {
		flattened := make(map[string]string)
		flattenValue("", record, flattened)
		for k := range flattened {
			allHeaders[k] = true
		}
		flattenedRecords = append(flattenedRecords, flattened)
	}
	headers := make([]string, 0, len(allHeaders))
	for k := range allHeaders {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	if err := writeCSVLine(output, headers, opts.Delimiter); err != nil {
		return err
	}
	for _, record := range flattenedRecords {
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = record[h]
		}
		if err := writeCSVLine(output, row, opts.Delimiter); err != nil {
			return err
		}
	}
	return nil
}
