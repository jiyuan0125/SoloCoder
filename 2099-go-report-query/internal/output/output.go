package output

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type QueryResult struct {
	Columns []string
	Rows    [][]string
}

func ReadRows(rows *sql.Rows) (*QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    [][]string{},
	}

	numCols := len(columns)
	values := make([]interface{}, numCols)
	valuePtrs := make([]interface{}, numCols)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make([]string, numCols)
		for i, v := range values {
			row[i] = formatValue(v)
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func PrintTable(result *QueryResult, writer io.Writer) {
	if len(result.Rows) == 0 {
		fmt.Fprintln(writer, "无匹配数据")
		return
	}

	table := tablewriter.NewWriter(writer)
	table.SetHeader(result.Columns)
	table.SetAutoWrapText(true)
	table.SetAutoFormatHeaders(true)
	table.SetBorder(true)
	table.AppendBulk(result.Rows)
	table.Render()

	fmt.Fprintf(writer, "\n共 %d 条记录\n", len(result.Rows))
}

func WriteCSV(result *QueryResult, filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("无法创建目录 %s: %w", dir, err)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("无法创建文件 %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(result.Columns); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	for _, row := range result.Rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入 CSV 行失败: %w", err)
		}
	}

	return nil
}

func Output(result *QueryResult, format string, csvPath string) error {
	switch strings.ToLower(format) {
	case "csv":
		if csvPath == "" {
			return fmt.Errorf("CSV 输出需要指定 --output 参数")
		}
		if len(result.Rows) == 0 {
			fmt.Println("无匹配数据")
			return nil
		}
		if err := WriteCSV(result, csvPath); err != nil {
			return err
		}
		fmt.Printf("数据已导出到: %s (共 %d 条记录)\n", csvPath, len(result.Rows))
	case "table", "":
		PrintTable(result, os.Stdout)
	default:
		return fmt.Errorf("不支持的输出格式: %s (支持: table, csv)", format)
	}
	return nil
}
