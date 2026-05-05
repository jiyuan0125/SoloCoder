package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

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

func colIndexToName(index int) string {
	name := ""
	for index >= 0 {
		name = string(rune('A'+(index%26))) + name
		index = index/26 - 1
	}
	return name
}

func (f *Formatter) ToExcel(records []map[string]interface{}, mappings []common.FieldMapping) ([]byte, error) {
	fx := excelize.NewFile()
	defer fx.Close()

	sheetName := "Sheet1"

	for i, m := range mappings {
		cell := colIndexToName(i) + "1"
		fx.SetCellValue(sheetName, cell, m.TargetField)
	}

	for rowIdx, record := range records {
		excelRow := rowIdx + 2
		for colIdx, m := range mappings {
			cell := colIndexToName(colIdx) + strconv.Itoa(excelRow)
			value := record[m.SourceField]

			if m.Sensitive {
				strVal := f.formatValue(value, &m)
				maskedVal := f.masker.Mask(strVal, string(m.FieldType))
				fx.SetCellValue(sheetName, cell, maskedVal)
			} else {
				switch v := value.(type) {
				case int:
					fx.SetCellValue(sheetName, cell, v)
				case int64:
					fx.SetCellValue(sheetName, cell, v)
				case float64:
					fx.SetCellValue(sheetName, cell, v)
				case bool:
					fx.SetCellValue(sheetName, cell, v)
				case time.Time:
					if m.FormatPattern != "" {
						fx.SetCellValue(sheetName, cell, v.Format(m.FormatPattern))
					} else {
						fx.SetCellValue(sheetName, cell, v)
					}
				default:
					fx.SetCellValue(sheetName, cell, fmt.Sprintf("%v", v))
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := fx.Write(&buf); err != nil {
		return nil, err
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
