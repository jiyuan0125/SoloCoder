package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CellStyle struct {
	Bold       bool    `json:"bold,omitempty"`
	Italic     bool    `json:"italic,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	FontColor  string  `json:"font_color,omitempty"`
	FillColor  string  `json:"fill_color,omitempty"`
	Alignment  string  `json:"alignment,omitempty"`
	Border     string  `json:"border,omitempty"`
}

type CellMapping struct {
	Cell        string      `json:"cell"`
	Field       string      `json:"field,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	Formula     string      `json:"formula,omitempty"`
	Style       *CellStyle  `json:"style,omitempty"`
}

type TableMapping struct {
	StartCell    string      `json:"start_cell"`
	Headers      []string    `json:"headers"`
	Fields       []string    `json:"fields"`
	Formulas     []string    `json:"formulas,omitempty"`
	HeaderStyle  *CellStyle  `json:"header_style,omitempty"`
	CellStyle    *CellStyle  `json:"cell_style,omitempty"`
	AutoFit      bool        `json:"auto_fit,omitempty"`
}

type MergeRange struct {
	StartCell string `json:"start_cell"`
	EndCell   string `json:"end_cell"`
}

type ConditionalFormat struct {
	Range       string `json:"range"`
	Condition   string `json:"condition"`
	Threshold   string `json:"threshold,omitempty"`
	FontColor   string `json:"font_color,omitempty"`
	FillColor   string `json:"fill_color,omitempty"`
}

type ChartConfig struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	DataRange   string `json:"data_range"`
	CategoryRange string `json:"category_range,omitempty"`
	Position    string `json:"position"`
}

type TemplateConfig struct {
	SheetName    string              `json:"sheet_name"`
	StaticCells  []CellMapping       `json:"static_cells,omitempty"`
	Tables       []TableMapping      `json:"tables,omitempty"`
	Merges       []MergeRange        `json:"merges,omitempty"`
	Conditionals []ConditionalFormat `json:"conditionals,omitempty"`
	Charts       []ChartConfig       `json:"charts,omitempty"`
}

type TemplateDef struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Sheets      []TemplateConfig  `json:"sheets"`
}

type ParseError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func LoadTemplate(tmplDir, name string) (*TemplateDef, error) {
	path := filepath.Join(tmplDir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template not found: %s", name)
		}
		return nil, err
	}

	var tmpl TemplateDef
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, &ParseError{Field: "json", Message: err.Error()}
	}

	if err := ValidateTemplate(&tmpl); err != nil {
		return nil, err
	}

	return &tmpl, nil
}

func SaveTemplate(tmplDir string, tmpl *TemplateDef) error {
	if err := ValidateTemplate(tmpl); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(tmplDir, tmpl.Name+".json")
	return os.WriteFile(path, data, 0644)
}

func ListTemplateNames(tmplDir string) ([]string, error) {
	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

func DeleteTemplate(tmplDir, name string) error {
	path := filepath.Join(tmplDir, name+".json")
	return os.Remove(path)
}

func ValidateTemplate(tmpl *TemplateDef) error {
	if strings.TrimSpace(tmpl.Name) == "" {
		return &ParseError{Field: "name", Message: "template name is required"}
	}
	if len(tmpl.Sheets) == 0 {
		return &ParseError{Field: "sheets", Message: "at least one sheet is required"}
	}

	for i, sheet := range tmpl.Sheets {
		if strings.TrimSpace(sheet.SheetName) == "" {
			return &ParseError{
				Field:   fmt.Sprintf("sheets[%d].sheet_name", i),
				Message: "sheet name is required",
			}
		}

		for j, table := range sheet.Tables {
			if len(table.Headers) != len(table.Fields) {
				return &ParseError{
					Field:   fmt.Sprintf("sheets[%d].tables[%d]", i, j),
					Message: "headers and fields must have the same length",
				}
			}
		}

		for j, chart := range sheet.Charts {
			if !isValidChartType(chart.Type) {
				return &ParseError{
					Field:   fmt.Sprintf("sheets[%d].charts[%d].type", i, j),
					Message: fmt.Sprintf("invalid chart type: %s (must be bar, line, or pie)", chart.Type),
				}
			}
			if strings.TrimSpace(chart.DataRange) == "" {
				return &ParseError{
					Field:   fmt.Sprintf("sheets[%d].charts[%d].data_range", i, j),
					Message: "data_range is required",
				}
			}
		}
	}

	return nil
}

func isValidChartType(t string) bool {
	types := []string{"bar", "line", "pie"}
	for _, valid := range types {
		if strings.EqualFold(t, valid) {
			return true
		}
	}
	return false
}
