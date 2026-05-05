package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"slowquery/protocol"
)

type Parser struct {
	format    *protocol.LogFormat
	delimiter string
	columns   []protocol.LogColumn
}

func NewParser(format *protocol.LogFormat) *Parser {
	delimiter := "\t"
	columns := []protocol.LogColumn{
		protocol.ColumnSQL,
		protocol.ColumnExecTime,
		protocol.ColumnScanRows,
		protocol.ColumnLockWait,
	}

	if format != nil {
		if format.Delimiter != "" {
			delimiter = format.Delimiter
		}
		if len(format.Columns) > 0 {
			columns = format.Columns
		}
	}

	return &Parser{
		format:    format,
		delimiter: delimiter,
		columns:   columns,
	}
}

type ParseResult struct {
	Entry   *protocol.LogEntry
	Error   error
	LineNum int
}

func (p *Parser) ParseFile(filePath string) (<-chan ParseResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	results := make(chan ParseResult, 100)

	go func() {
		defer file.Close()
		defer close(results)

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			entry, err := p.ParseLine(line)
			results <- ParseResult{
				Entry:   entry,
				Error:   err,
				LineNum: lineNum,
			}
		}

		if err := scanner.Err(); err != nil {
			results <- ParseResult{
				Error:   fmt.Errorf("error reading file: %v", err),
				LineNum: lineNum,
			}
		}
	}()

	return results, nil
}

func (p *Parser) ParseLine(line string) (*protocol.LogEntry, error) {
	parts := strings.Split(line, p.delimiter)

	entry := &protocol.LogEntry{}
	parsedFields := 0

	for i, col := range p.columns {
		if i >= len(parts) {
			continue
		}

		value := strings.TrimSpace(parts[i])
		if value == "" {
			continue
		}

		switch col {
		case protocol.ColumnSQL:
			entry.SQL = p.trimSQLQuotes(value)
			parsedFields++
		case protocol.ColumnExecTime:
			val, err := p.parseFloat(value)
			if err != nil {
				return nil, fmt.Errorf("invalid exec_time: %v", err)
			}
			entry.ExecTimeMs = val
			parsedFields++
		case protocol.ColumnScanRows:
			val, err := p.parseInt(value)
			if err != nil {
				return nil, fmt.Errorf("invalid scan_rows: %v", err)
			}
			entry.ScanRows = val
			parsedFields++
		case protocol.ColumnLockWait:
			val, err := p.parseFloat(value)
			if err != nil {
				return nil, fmt.Errorf("invalid lock_wait: %v", err)
			}
			entry.LockWaitMs = val
			parsedFields++
		}
	}

	if parsedFields == 0 {
		return nil, fmt.Errorf("no valid fields parsed")
	}

	return entry, nil
}

func (p *Parser) trimSQLQuotes(sql string) string {
	sql = strings.TrimSpace(sql)
	if len(sql) >= 2 {
		if (strings.HasPrefix(sql, "'") && strings.HasSuffix(sql, "'")) ||
		   (strings.HasPrefix(sql, "\"") && strings.HasSuffix(sql, "\"")) {
			sql = sql[1 : len(sql)-1]
		}
	}
	return sql
}

func (p *Parser) parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "ms")
	s = strings.TrimSuffix(s, "s")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, nil
	}

	return strconv.ParseFloat(s, 64)
}

func (p *Parser) parseInt(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func ParseFormatString(formatStr string) (*protocol.LogFormat, error) {
	format := &protocol.LogFormat{
		Delimiter: "\t",
		Columns: []protocol.LogColumn{
			protocol.ColumnSQL,
			protocol.ColumnExecTime,
			protocol.ColumnScanRows,
			protocol.ColumnLockWait,
		},
	}

	if formatStr == "" {
		return format, nil
	}

	parts := strings.SplitN(formatStr, ":", 2)
	if len(parts) == 2 {
		format.Delimiter = parts[0]
		if format.Delimiter == "" {
			format.Delimiter = "\t"
		}
		formatStr = parts[1]
	} else {
		formatStr = parts[0]
	}

	if formatStr != "" {
		colStrs := strings.Split(formatStr, ",")
		var columns []protocol.LogColumn
		for _, colStr := range colStrs {
			colStr = strings.TrimSpace(strings.ToLower(colStr))
			switch colStr {
			case "sql":
				columns = append(columns, protocol.ColumnSQL)
			case "exec_time", "exectime", "time":
				columns = append(columns, protocol.ColumnExecTime)
			case "scan_rows", "scanrows", "rows":
				columns = append(columns, protocol.ColumnScanRows)
			case "lock_wait", "lockwait", "lock":
				columns = append(columns, protocol.ColumnLockWait)
			default:
				return nil, fmt.Errorf("unknown column type: %s", colStr)
			}
		}
		if len(columns) > 0 {
			format.Columns = columns
		}
	}

	return format, nil
}
