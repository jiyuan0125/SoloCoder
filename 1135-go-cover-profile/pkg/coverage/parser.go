package coverage

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/solocoder/coverage-analyzer/pkg/api"
)

type CoverBlock struct {
	File      string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	Count     int
	Stmts     int
}

type CoverProfile struct {
	Mode   api.CoverageMode
	Blocks []CoverBlock
}

type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return e.Message
}

func Parse(content string) (*CoverProfile, error) {
	reader := strings.NewReader(content)
	return ParseReader(reader)
}

func ParseReader(reader io.Reader) (*CoverProfile, error) {
	profile := &CoverProfile{}
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "mode:") {
			mode := strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
			switch mode {
			case "set":
				profile.Mode = api.ModeSet
			case "count":
				profile.Mode = api.ModeCount
			case "atomic":
				profile.Mode = api.ModeAtomic
			default:
				return nil, &ParseError{Line: lineNum, Message: fmt.Sprintf("unknown coverage mode: %s", mode)}
			}
			continue
		}

		block, err := parseLine(line, lineNum)
		if err != nil {
			return nil, err
		}
		profile.Blocks = append(profile.Blocks, *block)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading coverprofile: %w", err)
	}

	if profile.Mode == "" {
		return nil, &ParseError{Message: "no mode declaration found"}
	}

	return profile, nil
}

func parseLine(line string, lineNum int) (*CoverBlock, error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, &ParseError{Line: lineNum, Message: "invalid format, expected 3 fields"}
	}

	rangePart := parts[0]
	countStr := parts[1]
	stmtsStr := parts[2]

	lastColon := strings.LastIndex(rangePart, ":")
	if lastColon == -1 {
		return nil, &ParseError{Line: lineNum, Message: "invalid range format, missing colon"}
	}

	file := rangePart[:lastColon]
	rangeStr := rangePart[lastColon+1:]

	rangeParts := strings.Split(rangeStr, ",")
	if len(rangeParts) != 2 {
		return nil, &ParseError{Line: lineNum, Message: "invalid range format, expected start,end"}
	}

	startLine, startCol, err := parsePosition(rangeParts[0], lineNum, "start")
	if err != nil {
		return nil, err
	}

	endLine, endCol, err := parsePosition(rangeParts[1], lineNum, "end")
	if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return nil, &ParseError{Line: lineNum, Message: fmt.Sprintf("invalid count: %s", countStr)}
	}

	stmts, err := strconv.Atoi(stmtsStr)
	if err != nil {
		return nil, &ParseError{Line: lineNum, Message: fmt.Sprintf("invalid stmts: %s", stmtsStr)}
	}

	return &CoverBlock{
		File:      file,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		Count:     count,
		Stmts:     stmts,
	}, nil
}

func parsePosition(pos string, lineNum int, name string) (int, int, error) {
	parts := strings.Split(pos, ".")
	if len(parts) != 2 {
		return 0, 0, &ParseError{Line: lineNum, Message: fmt.Sprintf("invalid %s position format", name)}
	}

	line, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, &ParseError{Line: lineNum, Message: fmt.Sprintf("invalid %s line number: %s", name, parts[0])}
	}

	col, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, &ParseError{Line: lineNum, Message: fmt.Sprintf("invalid %s column: %s", name, parts[1])}
	}

	return line, col, nil
}
