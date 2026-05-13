package report

import (
	"fmt"

	"reportgen/internal/datasource"
	"reportgen/internal/model"
	"reportgen/internal/output"
	"reportgen/internal/processor"
	"reportgen/internal/template"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) GenerateFromConfig(configPath string) (*model.ReportResult, error) {
	cfg, err := template.ParseReportConfig(configPath)
	if err != nil {
		return nil, err
	}

	return e.Generate(cfg)
}

func (e *Engine) Generate(cfg *model.ReportConfig) (*model.ReportResult, error) {
	rows, err := e.readData(cfg.DataSource)
	if err != nil {
		return nil, err
	}

	filteredRows := processor.ApplyFilters(rows, cfg.Template.Filters)

	result := &model.ReportResult{
		Name:    cfg.Name,
		Title:   cfg.Template.Title,
		Headers: cfg.Template.Columns,
	}

	if len(cfg.Template.GroupBy) > 0 || len(cfg.Template.Aggregations) > 0 {
		groups, groupKeys, err := processor.ApplyGroupBy(filteredRows, cfg.Template.GroupBy)
		if err != nil {
			return nil, err
		}

		if groups != nil && len(cfg.Template.Aggregations) > 0 {
			groupData, err := processor.CalculateAggregations(groups, cfg.Template.Aggregations, groupKeys)
			if err != nil {
				return nil, err
			}
			result.Groups = groupData
		} else {
			processed := processor.DetectAnomalies(filteredRows, cfg.Template.AnomalyRules)
			result.Rows = processed
			result.HasAnomaly = hasAnomaly(processed)
		}
	} else {
		processed := processor.DetectAnomalies(filteredRows, cfg.Template.AnomalyRules)
		result.Rows = processed
		result.HasAnomaly = hasAnomaly(processed)
	}

	if len(result.Headers) == 0 && len(result.Rows) > 0 {
		for k := range result.Rows[0].Values {
			result.Headers = append(result.Headers, k)
		}
	}

	return result, nil
}

func (e *Engine) GenerateAndOutput(cfg *model.ReportConfig) error {
	result, err := e.Generate(cfg)
	if err != nil {
		return err
	}

	switch cfg.OutputFormat {
	case model.OutputFormatHTML:
		return output.GenerateHTMLReport(result, cfg.OutputPath)
	case model.OutputFormatText:
		fallthrough
	default:
		return output.GenerateTextReport(result, cfg.OutputPath)
	}
}

func (e *Engine) readData(ds model.DataSource) ([]model.DataRow, error) {
	switch ds.Type {
	case model.DataSourceCSV:
		reader := datasource.NewCSVReader(ds.Path, ds.Delimiter, ds.Columns)
		return reader.Read()
	case model.DataSourceMySQL, model.DataSourcePostgres:
		reader := datasource.NewDatabaseReader(ds)
		return reader.Read()
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", ds.Type)
	}
}

func hasAnomaly(rows []model.ProcessedRow) bool {
	for _, row := range rows {
		for _, pv := range row.Values {
			if pv.IsAnomaly {
				return true
			}
		}
	}
	return false
}
