package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"data-export/pkg/common"
	"data-export/server/internal/security"
)

type Formatter struct {
	masker *security.Masker
}

func NewFormatter() *Formatter {
	return &Formatter{
		masker: security.NewMasker(),
	}
}

func (f *Formatter) ToCSV(records []map[string]interface{}, mappings []common.FieldMapping) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	headers := make([]string, len(mappings))
	for i, m := range mappings {
		headers[i] = m.TargetField
	}
	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, record := range records {
		row := make([]string, len(mappings))
		for i, m := range mappings {
			value := f.formatValue(record[m.SourceField], &m)
			if m.Sensitive {
				value = f.masker.Mask(value, string(m.FieldType))
			}
			row[i] = value
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func escapeCSVField(field string) string {
	if !strings.ContainsAny(field, ",\n\"") {
		return field
	}
	field = strings.ReplaceAll(field, "\"", "\"\"")
	return "\"" + field + "\""
}

func (f *Formatter) ToJSON(records []map[string]interface{}, mappings []common.FieldMapping) ([]byte, error) {
	var result []map[string]interface{}

	for _, record := range records {
		out := make(map[string]interface{})
		for _, m := range mappings {
			value := record[m.SourceField]
			if m.Sensitive {
				strVal := f.formatValue(value, &m)
				out[m.TargetField] = f.masker.Mask(strVal, string(m.FieldType))
			} else {
				out[m.TargetField] = f.convertValue(value, &m)
			}
		}
		result = append(result, out)
	}

	return json.MarshalIndent(result, "", "  ")
}

func (f *Formatter) ToExcel(records []map[string]interface{}, mappings []common.FieldMapping) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("\xEF\xBB\xBF")

	headers := make([]string, len(mappings))
	for i, m := range mappings {
		headers[i] = escapeCSVField(m.TargetField)
	}
	buf.WriteString(strings.Join(headers, ",") + "\n")

	for _, record := range records {
		row := make([]string, len(mappings))
		for i, m := range mappings {
			value := f.formatValue(record[m.SourceField], &m)
			if m.Sensitive {
				value = f.masker.Mask(value, string(m.FieldType))
			}
			row[i] = escapeCSVField(value)
		}
		buf.WriteString(strings.Join(row, ",") + "\n")
	}

	return buf.Bytes(), nil
}

func (f *Formatter) formatValue(value interface{}, mapping *common.FieldMapping) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		if mapping.FieldType == common.FieldTypeFloat && mapping.FormatPattern != "" {
			return fmt.Sprintf(mapping.FormatPattern, v)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		if mapping.FormatPattern != "" {
			return v.Format(mapping.FormatPattern)
		}
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (f *Formatter) convertValue(value interface{}, mapping *common.FieldMapping) interface{} {
	if value == nil {
		return nil
	}

	switch mapping.FieldType {
	case common.FieldTypeInteger:
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return v
		case float64:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	case common.FieldTypeFloat:
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	case common.FieldTypeDate:
		switch v := value.(type) {
		case time.Time:
			if mapping.FormatPattern != "" {
				return v.Format(mapping.FormatPattern)
			}
			return v.Format(time.RFC3339)
		}
	}

	return value
}

func (f *Formatter) Format(records []map[string]interface{}, mappings []common.FieldMapping, format common.ExportFormat) ([]byte, error) {
	switch format {
	case common.FormatCSV:
		return f.ToCSV(records, mappings)
	case common.FormatJSON:
		return f.ToJSON(records, mappings)
	case common.FormatExcel:
		return f.ToExcel(records, mappings)
	default:
		return f.ToCSV(records, mappings)
	}
}
