package template

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"reportgen/internal/model"
)

type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

func ParseReportConfig(path string) (*model.ReportConfig, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("template file does not exist: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access template file: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("template path is a directory, not a file: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var config model.ReportConfig

	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	if err := decoder.Decode(&config); err != nil {
		var lineNum int
		if te, ok := err.(*yaml.TypeError); ok {
			for _, e := range te.Errors {
				parts := strings.Split(e, " ")
				if len(parts) > 0 {
					fmt.Sscanf(parts[0], "line %d", &lineNum)
				}
			}
		}
		if lineNum == 0 {
			fmt.Sscanf(err.Error(), "yaml: line %d", &lineNum)
		}
		return nil, &ParseError{Line: lineNum, Message: fmt.Sprintf("YAML parse error: %v", err)}
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(cfg *model.ReportConfig) error {
	if cfg.Name == "" {
		return &ParseError{Line: 0, Message: "report name is required"}
	}

	if cfg.DataSource.Type == "" {
		return &ParseError{Line: 0, Message: "data source type is required"}
	}

	switch cfg.DataSource.Type {
	case model.DataSourceCSV:
		if cfg.DataSource.Path == "" {
			return &ParseError{Line: 0, Message: "CSV data source requires 'path' field"}
		}
	case model.DataSourceMySQL, model.DataSourcePostgres:
		if cfg.DataSource.Database == "" {
			return &ParseError{Line: 0, Message: "database source requires 'database' field"}
		}
		if cfg.DataSource.Table == "" && cfg.DataSource.Query == "" {
			return &ParseError{Line: 0, Message: "database source requires 'table' or 'query' field"}
		}
	default:
		return &ParseError{Line: 0, Message: fmt.Sprintf("unsupported data source type: %s", cfg.DataSource.Type)}
	}

	for i, gb := range cfg.Template.GroupBy {
		if gb.Field == "" {
			return &ParseError{Line: 0, Message: fmt.Sprintf("group_by[%d]: field is required", i)}
		}
		if gb.DateGroup != "" && !model.IsValidDateGroupType(string(gb.DateGroup)) {
			return &ParseError{Line: 0, Message: fmt.Sprintf("group_by[%d]: invalid date_group '%s', valid values: year, month, week, day", i, gb.DateGroup)}
		}
	}

	for i, agg := range cfg.Template.Aggregations {
		if agg.Field == "" && agg.Type != model.AggCount {
			return &ParseError{Line: 0, Message: fmt.Sprintf("aggregations[%d]: field is required for non-count aggregations", i)}
		}
		if !model.IsValidAggregationType(string(agg.Type)) {
			return &ParseError{Line: 0, Message: fmt.Sprintf("aggregations[%d]: invalid type '%s', valid values: sum, avg, count, min, max", i, agg.Type)}
		}
	}

	if cfg.OutputFormat == "" {
		cfg.OutputFormat = model.OutputFormatText
	} else if cfg.OutputFormat != model.OutputFormatText && cfg.OutputFormat != model.OutputFormatHTML {
		return &ParseError{Line: 0, Message: fmt.Sprintf("invalid output_format '%s', valid values: text, html", cfg.OutputFormat)}
	}

	return nil
}
