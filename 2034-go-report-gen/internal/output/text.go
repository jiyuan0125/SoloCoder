package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"

	"reportgen/internal/model"
)

func GenerateTextReport(result *model.ReportResult, outputPath string) error {
	var buf bytes.Buffer
	bufWriter := io.Writer(&buf)

	if result.Title != "" {
		buf.WriteString(fmt.Sprintf("=== %s ===\n\n", result.Title))
	}

	if len(result.Groups) > 0 {
		renderGroupedTable(bufWriter, result)
	} else {
		renderSimpleTable(bufWriter, result)
	}

	if result.HasAnomaly {
		buf.WriteString("\n注: 标有 * 的值为异常值\n")
	}

	if outputPath == "" {
		fmt.Println(buf.String())
		return nil
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func renderSimpleTable(w io.Writer, result *model.ReportResult) {
	if len(result.Headers) == 0 {
		return
	}

	table := tablewriter.NewWriter(w)
	headerRow := make([]interface{}, len(result.Headers))
	for i, h := range result.Headers {
		headerRow[i] = h
	}
	table.Header(headerRow...)

	for _, row := range result.Rows {
		data := make([]interface{}, len(result.Headers))
		for i, h := range result.Headers {
			pv, exists := row.Values[h]
			if !exists {
				data[i] = ""
				continue
			}
			data[i] = formatValue(pv, true)
		}
		table.Append(data)
	}

	table.Render()
}

func renderGroupedTable(w io.Writer, result *model.ReportResult) {
	if len(result.Groups) == 0 {
		return
	}

	groupKeys := make([]string, 0)
	if len(result.Groups) > 0 {
		for k := range result.Groups[0].Key.Fields {
			groupKeys = append(groupKeys, k)
		}
	}

	aggKeys := make([]string, 0)
	if len(result.Groups) > 0 {
		for k := range result.Groups[0].Values {
			aggKeys = append(aggKeys, k)
		}
	}

	headers := make([]string, 0, len(groupKeys)+len(aggKeys))
	headers = append(headers, groupKeys...)
	headers = append(headers, aggKeys...)

	table := tablewriter.NewWriter(w)
	headerRow := make([]interface{}, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	table.Header(headerRow...)

	for _, group := range result.Groups {
		row := make([]interface{}, len(headers))
		idx := 0

		for _, k := range groupKeys {
			row[idx] = fmt.Sprintf("%v", group.Key.Fields[k])
			idx++
		}

		for _, k := range aggKeys {
			row[idx] = formatValueSimple(group.Values[k])
			idx++
		}

		table.Append(row)
	}

	table.Render()
}

func formatValue(pv model.ProcessedValue, showAnomaly bool) string {
	if pv.Value == nil {
		return ""
	}

	str := formatValueSimple(pv.Value)

	if showAnomaly && pv.IsAnomaly {
		return "*" + str + "*"
	}
	return str
}

func formatValueSimple(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case float32:
		return fmt.Sprintf("%.2f", val)
	case int, int64, int32, int16, int8, uint, uint64, uint32:
		return fmt.Sprintf("%d", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func GenerateText(rows []model.ProcessedRow, columns []string, hasAnomaly bool) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)

	if len(columns) == 0 {
		if len(rows) > 0 {
			for k := range rows[0].Values {
				columns = append(columns, k)
			}
		}
	}

	table := tablewriter.NewWriter(w)
	headerRow := make([]interface{}, len(columns))
	for i, h := range columns {
		headerRow[i] = h
	}
	table.Header(headerRow...)

	for _, row := range rows {
		data := make([]interface{}, len(columns))
		for i, c := range columns {
			if pv, ok := row.Values[c]; ok {
				data[i] = formatValue(pv, true)
			} else {
				data[i] = ""
			}
		}
		table.Append(data)
	}

	table.Render()

	if hasAnomaly {
		buf.WriteString("\n注: 标有 * 的值为异常值\n")
	}

	return buf.String()
}
