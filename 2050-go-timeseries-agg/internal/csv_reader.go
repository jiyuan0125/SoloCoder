package internal

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type CSVReader struct {
	reader      *csv.Reader
	file        *os.File
	scanner     *bufio.Scanner
	timezone    *time.Location
	header      []string
	headerLine  int
	currentLine int
}

func NewCSVReader(filename string) (*CSVReader, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("输入文件不存在: %s", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("无法访问输入文件: %s, 错误: %v", filename, err)
	}
	if info.Size() == 0 {
		return nil, fmt.Errorf("输入文件为空: %s", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %s, 错误: %v", filename, err)
	}

	scanner := bufio.NewScanner(file)

	cr := &CSVReader{
		file:        file,
		scanner:     scanner,
		currentLine: 0,
	}

	return cr, nil
}

func (cr *CSVReader) ReadHeader() ([]string, error) {
	var lines []string
	for i := 0; i < 3; i++ {
		if !cr.scanner.Scan() {
			break
		}
		line := cr.scanner.Text()
		cr.currentLine++
		lines = append(lines, line)
	}

	if err := cr.scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件头失败: %v", err)
	}

	var tzName string
	var headerStart int

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "timezone=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				tzName = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, ",") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			headerStart = i
			break
		}
	}

	if tzName == "" {
		tzName = "UTC"
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("解析时区失败，行号: %d, 原因: %s", cr.currentLine-len(lines)+1, err.Error())
	}
	cr.timezone = loc

	if headerStart >= len(lines) {
		return nil, fmt.Errorf("未找到 CSV 表头")
	}

	header := strings.Split(lines[headerStart], ",")
	for i, h := range header {
		header[i] = strings.TrimSpace(h)
	}
	cr.header = header
	cr.headerLine = cr.currentLine - len(lines) + headerStart + 1

	remaining := strings.Join(lines[headerStart+1:], "\n")

	file, err := os.Open(cr.file.Name())
	if err != nil {
		return nil, fmt.Errorf("重新打开文件失败: %v", err)
	}
	cr.file.Close()
	cr.file = file

	reader := bufio.NewReader(file)
	for i := 0; i < cr.headerLine; i++ {
		_, _, err := reader.ReadLine()
		if err != nil {
			break
		}
	}

	if remaining != "" {
		cr.reader = csv.NewReader(io.MultiReader(strings.NewReader(remaining+"\n"), reader))
	} else {
		cr.reader = csv.NewReader(reader)
	}
	cr.reader.FieldsPerRecord = -1
	cr.currentLine = cr.headerLine

	return header, nil
}

func (cr *CSVReader) GetTimezone() *time.Location {
	return cr.timezone
}

func (cr *CSVReader) GetHeader() []string {
	return cr.header
}

type DataRow struct {
	Time    time.Time
	Values  []float64
	LineNum int
}

type RowCallback func(row DataRow, valid bool, warning string)

func (cr *CSVReader) ReadAll(timeColIdx int, valueColIdxs []int, callback RowCallback) error {
	if cr.reader == nil {
		return fmt.Errorf("请先调用 ReadHeader")
	}

	expectedCols := len(cr.header)

	for {
		record, err := cr.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			cr.currentLine++
			callback(DataRow{}, false, fmt.Sprintf("行号 %d: 读取错误: %v", cr.currentLine, err))
			continue
		}
		cr.currentLine++

		if len(record) != expectedCols {
			callback(DataRow{}, false, fmt.Sprintf("行号 %d: 列数不一致，期望 %d，实际 %d", cr.currentLine, expectedCols, len(record)))
			continue
		}

		if timeColIdx < 0 || timeColIdx >= len(record) {
			callback(DataRow{}, false, fmt.Sprintf("行号 %d: 时间列索引 %d 超出范围", cr.currentLine, timeColIdx))
			continue
		}

		timeStr := strings.TrimSpace(record[timeColIdx])
		t, err := cr.parseTime(timeStr)
		if err != nil {
			callback(DataRow{}, false, fmt.Sprintf("行号 %d: 时间解析失败: %v", cr.currentLine, err))
			continue
		}

		values := make([]float64, len(valueColIdxs))
		valid := true
		var warnMsg string

		for i, colIdx := range valueColIdxs {
			if colIdx < 0 || colIdx >= len(record) {
				values[i] = 0
				warnMsg = fmt.Sprintf("行号 %d: 列索引 %d 超出范围", cr.currentLine, colIdx)
				valid = false
				continue
			}

			valStr := strings.TrimSpace(record[colIdx])
			if valStr == "" || strings.EqualFold(valStr, "null") || strings.EqualFold(valStr, "na") {
				values[i] = 0
				continue
			}

			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				warnMsg = fmt.Sprintf("行号 %d: 数值解析失败: %v", cr.currentLine, err)
				valid = false
				continue
			}
			values[i] = val
		}

		callback(DataRow{Time: t, Values: values, LineNum: cr.currentLine}, valid, warnMsg)
	}

	return nil
}

func (cr *CSVReader) parseTime(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"15:04:05",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, timeStr, cr.timezone); err == nil {
			return t, nil
		}
	}

	if ts, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(ts, 0).In(cr.timezone), nil
	}

	return time.Time{}, fmt.Errorf("无法解析时间格式: %s", timeStr)
}

func (cr *CSVReader) Close() {
	if cr.file != nil {
		cr.file.Close()
	}
}
