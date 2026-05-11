package parser

import (
	"bufio"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go-escape-analyze/pkg/api"
)

var locationPattern = regexp.MustCompile(`^(.+?):(\d+)(?::\d+)?:\s*(.+)$`)

var escapeReasonPatterns = []escapePattern{
	{
		regex:  regexp.MustCompile(`^([^\s:]+)\s+escapes to heap\b`),
		reason: "escapes to heap",
	},
	{
		regex:  regexp.MustCompile(`^moved to heap:\s*(.+)$`),
		reason: "moved to heap",
	},
	{
		regex:  regexp.MustCompile(`^parameter\s+([^\s]+)\s+leaks to result$`),
		reason: "parameter leaks to result",
	},
	{
		regex:  regexp.MustCompile(`^([^\s]+)\s+too complex for escape analysis$`),
		reason: "too complex for escape analysis",
	},
	{
		regex:  regexp.MustCompile(`^([^\s]+)\s+ escapes to heap`),
		reason: "escapes to heap",
	},
	{
		regex:  regexp.MustCompile(`^([^\s]+)\s+ escapes$`),
		reason: "escapes",
	},
}

var funcPattern = regexp.MustCompile(`^func\s+(.+?)\s+does not escape`)

type escapePattern struct {
	regex  *regexp.Regexp
	reason string
}

func ParseOutput(output string, filter *api.FilterOptions) *api.EscapeReport {
	scanner := bufio.NewScanner(strings.NewReader(output))
	seen := make(map[string]bool)
	var records []api.EscapeRecord

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if isFuncAnnotation(line) {
			continue
		}

		record := parseEscapeLine(line)
		if record == nil {
			continue
		}

		key := buildRecordKey(record)
		if seen[key] {
			continue
		}
		seen[key] = true

		if matchesFilter(record, filter) {
			records = append(records, *record)
		}
	}

	return buildReport(records, filter)
}

func parseEscapeLine(line string) *api.EscapeRecord {
	locMatches := locationPattern.FindStringSubmatch(line)
	if locMatches == nil {
		return nil
	}

	fileName := locMatches[1]
	lineNum, _ := strconv.Atoi(locMatches[2])
	message := locMatches[3]

	for _, pattern := range escapeReasonPatterns {
		if msgMatches := pattern.regex.FindStringSubmatch(message); msgMatches != nil {
			varName := ""
			if len(msgMatches) > 1 {
				varName = strings.TrimSpace(msgMatches[1])
			}

			return &api.EscapeRecord{
				FileName:        fileName,
				LineNumber:      lineNum,
				VariableName:    varName,
				EscapeReason:    pattern.reason,
				IsGenerated:     isGeneratedVar(varName),
				IsReturnValue:   isReturnValue(varName),
				IsClosureVar:    isClosureVar(varName),
			}
		}
	}

	return nil
}

func isFuncAnnotation(line string) bool {
	if funcPattern.MatchString(line) {
		return true
	}
	locMatches := locationPattern.FindStringSubmatch(line)
	if locMatches != nil {
		return strings.Contains(locMatches[3], "does not escape")
	}
	return strings.Contains(line, "does not escape")
}

func isGeneratedVar(name string) bool {
	return strings.HasPrefix(name, "~") || strings.HasPrefix(name, ".")
}

func isReturnValue(name string) bool {
	return strings.HasPrefix(name, "~r")
}

func isClosureVar(name string) bool {
	return strings.Contains(name, ".") && !strings.HasPrefix(name, "~")
}

func buildRecordKey(r *api.EscapeRecord) string {
	return r.FileName + ":" + strconv.Itoa(r.LineNumber) + ":" + r.VariableName + ":" + r.EscapeReason
}

func matchesFilter(record *api.EscapeRecord, filter *api.FilterOptions) bool {
	if filter == nil {
		return true
	}

	if filter.FileName != "" && !strings.Contains(record.FileName, filter.FileName) {
		return false
	}

	if filter.FunctionName != "" && !strings.Contains(record.FunctionSignature, filter.FunctionName) {
		return false
	}

	if filter.EscapeReason != "" && !strings.Contains(record.EscapeReason, filter.EscapeReason) {
		return false
	}

	return true
}

func buildReport(records []api.EscapeRecord, filter *api.FilterOptions) *api.EscapeReport {
	report := &api.EscapeReport{
		TotalEscapes: len(records),
		FilteredBy:   filter,
	}

	fileMap := make(map[string][]api.EscapeRecord)

	for _, r := range records {
		fileMap[r.FileName] = append(fileMap[r.FileName], r)
	}

	var files []api.FileReport

	for filePath, fileRecords := range fileMap {
		sort.Slice(fileRecords, func(i, j int) bool {
			if fileRecords[i].LineNumber != fileRecords[j].LineNumber {
				return fileRecords[i].LineNumber < fileRecords[j].LineNumber
			}
			return fileRecords[i].VariableName < fileRecords[j].VariableName
		})

		files = append(files, api.FileReport{
			FilePath:      filePath,
			EscapeCount:   len(fileRecords),
			EscapeRecords: fileRecords,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].FilePath < files[j].FilePath
	})

	report.Files = files

	return report
}
