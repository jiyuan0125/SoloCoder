package converter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConversionResult struct {
	Data           []byte
	ProcessedCount int
	SkippedCount   int
	TotalCount      int
	Errors          []string
}

var SupportedFormats = map[string][]string{
	"csv":      {"json"},
	"json":     {"csv", "yaml"},
	"yaml":     {"json"},
	"markdown": {"html"},
}

func IsSupported(source, target string) bool {
	targets, ok := SupportedFormats[source]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == target {
			return true
		}
	}
	return false
}

func GetSupportedFormats() map[string][]string {
	return SupportedFormats
}

func Convert(sourceFormat, targetFormat string, data []byte) (*ConversionResult, error) {
	switch sourceFormat + "->" + targetFormat {
	case "csv->json":
		return CSVToJSON(data)
	case "json->csv":
		return JSONToCSV(data)
	case "json->yaml":
		return JSONToYAML(data)
	case "yaml->json":
		return YAMLToJSON(data)
	case "markdown->html":
		return MarkdownToHTML(data)
	default:
		return nil, fmt.Errorf("不支持的转换: %s -> %s", sourceFormat, targetFormat)
	}
}

func CSVToJSON(data []byte) (*ConversionResult, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	var records []map[string]interface{}
	var headers []string
	var processed, skipped, total int
	var errors []string

	rowNum := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		total++
		rowNum++

		if err != nil {
			skipped++
			errors = append(errors, fmt.Sprintf("第%d行: 解析错误: %v", rowNum, err))
			continue
		}

		if len(row) == 0 {
			continue
		}

		if headers == nil {
			headers = row
			processed++
			continue
		}

		record, err := parseCSVRow(headers, row)
		if err != nil {
			skipped++
			errors = append(errors, fmt.Sprintf("第%d行: %v", rowNum, err))
			continue
		}
		records = append(records, record)
		processed++
	}

	result, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	return &ConversionResult{
		Data:           result,
		ProcessedCount: processed,
		SkippedCount:  skipped,
		TotalCount:    total,
		Errors:         errors,
	}, nil
}

func parseCSVRow(headers, row []string) (map[string]interface{}, error) {
	record := make(map[string]interface{})
	
	if len(row) < len(headers) {
		for i := range row {
			if i >= len(headers) {
				break
			}
			setNestedValue(record, headers[i], parseValue(row[i]))
		}
	} else {
		for i, header := range headers {
			setNestedValue(record, header, parseValue(row[i]))
		}
	}
	
	return record, nil
}

func parseValue(v string) interface{} {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	
	if v == "true" {
		return true
	}
	if v == "false" {
		return false
	}
	
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i
	}
	
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	
	return v
}

func setNestedValue(m map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		m[key] = value
		return
	}
	
	current := m
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return
		}
	}
}

func JSONToCSV(data []byte) (*ConversionResult, error) {
	var records []map[string]interface{}
	
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	var processed, skipped, total int
	var errors []string
	rowNum := 0
	
	for decoder.More() {
		var value interface{}
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}
		total++
		rowNum++
		
		if err != nil {
			skipped++
			errors = append(errors, fmt.Sprintf("第%d个JSON对象: 解析错误: %v", rowNum, err))
			continue
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			records = append(records, v)
			processed++
		case []interface{}:
			for _, item := range v {
				total++
				if record, ok := item.(map[string]interface{}); ok {
					records = append(records, record)
					processed++
				} else {
					skipped++
					errors = append(errors, fmt.Sprintf("第%d个数组项: 不是有效的对象", total))
				}
			}
		default:
			skipped++
			errors = append(errors, fmt.Sprintf("第%d个JSON项: 不是有效的对象或数组", rowNum))
		}
	}
	
	if len(records) == 0 {
		return nil, fmt.Errorf("没有可转换的数据")
	}
	
	headers := collectHeaders(records)
	
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("写入CSV头部失败: %v", err)
	}
	
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getNestedValue(record, header)
		}
		if err := writer.Write(row); err != nil {
			skipped++
			errors = append(errors, fmt.Sprintf("写入记录失败: %v", err))
			continue
		}
	}
	
	writer.Flush()
	
	return &ConversionResult{
		Data:           []byte(builder.String()),
		ProcessedCount: processed,
		SkippedCount:  skipped,
		TotalCount:    total,
		Errors:         errors,
	}, nil
}

func collectHeaders(records []map[string]interface{}) []string {
	headerSet := make(map[string]bool)
	var headers []string
	
	for _, record := range records {
		collectKeys("", record, headerSet, &headers)
	}
	
	return headers
}

func collectKeys(prefix string, value interface{}, headerSet map[string]bool, headers *[]string) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "." + key
			}
			if isNestedMap(val) {
				collectKeys(newPrefix, val, headerSet, headers)
			} else {
				if !headerSet[newPrefix] {
					headerSet[newPrefix] = true
					*headers = append(*headers, newPrefix)
				}
			}
		}
	}
}

func isNestedMap(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}

func getNestedValue(record map[string]interface{}, key string) string {
	parts := strings.Split(key, ".")
	current := record
	
	for i, part := range parts {
		if i == len(parts)-1 {
			return formatValue(current[part])
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func JSONToYAML(data []byte) (*ConversionResult, error) {
	var value interface{}
	var processed, skipped, total int
	var errors []string
	
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	rowNum := 0
	
	for {
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}
		total++
		rowNum++
		
		if err != nil {
			skipped++
			errors = append(errors, fmt.Sprintf("第%d个JSON对象: 解析错误: %v", rowNum, err))
			continue
		}
		processed++
		break
	}
	
	if value == nil {
		return nil, fmt.Errorf("没有可转换的数据")
	}
	
	result, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("YAML序列化失败: %v", err)
	}
	
	return &ConversionResult{
		Data:           result,
		ProcessedCount: processed,
		SkippedCount:  skipped,
		TotalCount:    total,
		Errors:         errors,
	}, nil
}

func YAMLToJSON(data []byte) (*ConversionResult, error) {
	var value interface{}
	var processed, skipped, total int
	var errors []string
	
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	rowNum := 0
	
	for {
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}
		total++
		rowNum++
		
		if err != nil {
			skipped++
			errors = append(errors, fmt.Sprintf("第%d个YAML文档: 解析错误: %v", rowNum, err))
			continue
		}
		processed++
		break
	}
	
	if value == nil {
		return nil, fmt.Errorf("没有可转换的数据")
	}
	
	value = convertYAMLMapToJSONMap(value)
	
	result, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}
	
	return &ConversionResult{
		Data:           result,
		ProcessedCount: processed,
		SkippedCount:  skipped,
		TotalCount:    total,
		Errors:         errors,
	}, nil
}

func convertYAMLMapToJSONMap(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, vv := range val {
			result[fmt.Sprintf("%v", k)] = convertYAMLMapToJSONMap(vv)
		}
		return result
	case map[string]interface{}:
		for k, vv := range val {
			val[k] = convertYAMLMapToJSONMap(vv)
		}
		return val
	case []interface{}:
		for i, vv := range val {
			val[i] = convertYAMLMapToJSONMap(vv)
		}
		return val
	default:
		return v
	}
}

func MarkdownToHTML(data []byte) (*ConversionResult, error) {
	content := string(data)
	var html strings.Builder
	var processed, skipped, total int
	var errors []string
	
	html.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"UTF-8\">\n</head>\n<body>\n")
	
	lines := strings.Split(content, "\n")
	total = len(lines)
	
	var inList bool
	var listType string
	var inCode bool
	
	for _, line := range lines {
		processed++
		originalLine := line
		
		if strings.HasPrefix(line, "```") {
			if inCode {
				html.WriteString("</code></pre>\n")
				inCode = false
			} else {
				html.WriteString("<pre><code>")
				inCode = true
			}
			continue
		}
		
		if inCode {
			html.WriteString(escapeHTML(line))
			html.WriteString("\n")
			continue
		}
		
		if inList && !strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "* ") && !strings.HasPrefix(line, "1. ") && !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t") {
			html.WriteString("</")
			html.WriteString(listType)
			html.WriteString(">\n")
			inList = false
		}
		
		switch {
		case strings.HasPrefix(line, "# "):
			html.WriteString("<h1>")
			html.WriteString(convertInlineMarkdown(line[2:]))
			html.WriteString("</h1>\n")
		case strings.HasPrefix(line, "## "):
			html.WriteString("<h2>")
			html.WriteString(convertInlineMarkdown(line[3:]))
			html.WriteString("</h2>\n")
		case strings.HasPrefix(line, "### "):
			html.WriteString("<h3>")
			html.WriteString(convertInlineMarkdown(line[4:]))
			html.WriteString("</h3>\n")
		case strings.HasPrefix(line, "#### "):
			html.WriteString("<h4>")
			html.WriteString(convertInlineMarkdown(line[5:]))
			html.WriteString("</h4>\n")
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			if !inList {
				html.WriteString("<ul>\n")
				inList = true
				listType = "ul"
			}
			html.WriteString("<li>")
			html.WriteString(convertInlineMarkdown(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")))
			html.WriteString("</li>\n")
		case strings.HasPrefix(line, "1. "):
			if !inList {
				html.WriteString("<ol>\n")
				inList = true
				listType = "ol"
			}
			html.WriteString("<li>")
			html.WriteString(convertInlineMarkdown(line[3:]))
			html.WriteString("</li>\n")
		case strings.HasPrefix(line, "> "):
			html.WriteString("<blockquote><p>")
			html.WriteString(convertInlineMarkdown(line[2:]))
			html.WriteString("</p></blockquote>\n")
		case strings.TrimSpace(line) == "":
			html.WriteString("<br>\n")
		default:
			html.WriteString("<p>")
			html.WriteString(convertInlineMarkdown(originalLine))
			html.WriteString("</p>\n")
		}
	}
	
	if inList {
		html.WriteString("</")
		html.WriteString(listType)
		html.WriteString(">\n")
	}
	
	html.WriteString("</body>\n</html>")
	
	return &ConversionResult{
		Data:           []byte(html.String()),
		ProcessedCount: processed,
		SkippedCount:  skipped,
		TotalCount:    total,
		Errors:         errors,
	}, nil
}

func convertInlineMarkdown(text string) string {
	text = escapeHTML(text)
	
	for strings.Contains(text, "**") {
		text = strings.Replace(text, "**", "<strong>", 1)
		text = strings.Replace(text, "**", "</strong>", 1)
	}
	
	for strings.Contains(text, "*") {
		text = strings.Replace(text, "*", "<em>", 1)
		text = strings.Replace(text, "*", "</em>", 1)
	}
	
	for strings.Contains(text, "`") {
		text = strings.Replace(text, "`", "<code>", 1)
		text = strings.Replace(text, "`", "</code>", 1)
	}
	
	return text
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
