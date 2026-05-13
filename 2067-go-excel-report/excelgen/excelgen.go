package excelgen

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"excel-report-system/template"

	"github.com/xuri/excelize/v2"
)

const BatchSize = 100000

type DataSource struct {
	StaticData map[string]interface{}            `json:"static_data"`
	TableData  map[string][]map[string]interface{} `json:"table_data"`
}

type ChartError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

func GenerateExcel(tmpl *template.TemplateDef, dataSource *DataSource, outputPath string) ([]*ChartError, error) {
	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := "Sheet1"
	hasDefault := false
	for _, sheet := range tmpl.Sheets {
		if sheet.SheetName == defaultSheet {
			hasDefault = true
			break
		}
	}
	if !hasDefault && len(tmpl.Sheets) > 0 {
		f.SetSheetName(defaultSheet, tmpl.Sheets[0].SheetName)
	}

	var chartErrors []*ChartError

	for i, sheetConfig := range tmpl.Sheets {
		sheetName := sheetConfig.SheetName
		if i > 0 || (i == 0 && hasDefault) {
			index, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, err
			}
			f.SetActiveSheet(index)
		}

		if err := applyMerges(f, sheetName, sheetConfig.Merges); err != nil {
			return nil, err
		}

		if err := applyStaticCells(f, sheetName, sheetConfig.StaticCells, dataSource.StaticData); err != nil {
			return nil, err
		}

		if err := applyTables(f, sheetName, sheetConfig.Tables, dataSource.TableData); err != nil {
			return nil, err
		}

		if err := applyConditionals(f, sheetName, sheetConfig.Conditionals); err != nil {
			return nil, err
		}

		for ci, chart := range sheetConfig.Charts {
			if err := applyChart(f, sheetName, &chart); err != nil {
				chartErrors = append(chartErrors, &ChartError{
					Index:   ci,
					Message: err.Error(),
				})
			}
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return nil, err
	}

	return chartErrors, nil
}

func ParseDataSource(jsonData string) (*DataSource, error) {
	if strings.TrimSpace(jsonData) == "" {
		return &DataSource{
			StaticData: make(map[string]interface{}),
			TableData:  make(map[string][]map[string]interface{}),
		}, nil
	}

	var ds DataSource
	if err := json.Unmarshal([]byte(jsonData), &ds); err != nil {
		return nil, err
	}

	if ds.StaticData == nil {
		ds.StaticData = make(map[string]interface{})
	}
	if ds.TableData == nil {
		ds.TableData = make(map[string][]map[string]interface{})
	}

	return &ds, nil
}

func applyMerges(f *excelize.File, sheetName string, merges []template.MergeRange) error {
	for _, m := range merges {
		if err := f.MergeCell(sheetName, m.StartCell, m.EndCell); err != nil {
			return fmt.Errorf("merge %s:%s: %w", m.StartCell, m.EndCell, err)
		}
	}
	return nil
}

func applyStaticCells(f *excelize.File, sheetName string, cells []template.CellMapping, staticData map[string]interface{}) error {
	for _, cell := range cells {
		var styleID int
		var err error
		if cell.Style != nil {
			styleID, err = createStyle(f, cell.Style)
			if err != nil {
				return err
			}
		}

		if cell.Formula != "" {
			if err := f.SetCellFormula(sheetName, cell.Cell, cell.Formula); err != nil {
				return err
			}
		} else {
			var val interface{}
			if cell.Field != "" {
				val = staticData[cell.Field]
			} else {
				val = cell.Value
			}

			if err := f.SetCellValue(sheetName, cell.Cell, val); err != nil {
				return err
			}
		}

		if cell.Style != nil {
			if err := f.SetCellStyle(sheetName, cell.Cell, cell.Cell, styleID); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyTables(f *excelize.File, sheetName string, tables []template.TableMapping, tableData map[string][]map[string]interface{}) error {
	for ti, table := range tables {
		startCol, startRow, err := excelize.CellNameToCoordinates(table.StartCell)
		if err != nil {
			return fmt.Errorf("table[%d] invalid start cell: %w", ti, err)
		}

		dataKey := fmt.Sprintf("table_%d", ti)
		rows := tableData[dataKey]

		var headerStyleID, cellStyleID int
		if table.HeaderStyle != nil {
			headerStyleID, err = createStyle(f, table.HeaderStyle)
			if err != nil {
				return err
			}
		}
		if table.CellStyle != nil {
			cellStyleID, err = createStyle(f, table.CellStyle)
			if err != nil {
				return err
			}
		}

		for ci, header := range table.Headers {
			cell, _ := excelize.CoordinatesToCellName(startCol+ci, startRow)
			if err := f.SetCellValue(sheetName, cell, header); err != nil {
				return err
			}
			if table.HeaderStyle != nil {
				if err := f.SetCellStyle(sheetName, cell, cell, headerStyleID); err != nil {
					return err
				}
			}
		}

		totalRows := len(rows)
		for batchStart := 0; batchStart < totalRows; batchStart += BatchSize {
			batchEnd := batchStart + BatchSize
			if batchEnd > totalRows {
				batchEnd = totalRows
			}

			for ri := batchStart; ri < batchEnd; ri++ {
				rowData := rows[ri]
				excelRow := startRow + 1 + ri

				for ci, field := range table.Fields {
					cell, _ := excelize.CoordinatesToCellName(startCol+ci, excelRow)

					if table.Formulas != nil && ci < len(table.Formulas) && table.Formulas[ci] != "" {
						formula := replaceRowPlaceholder(table.Formulas[ci], excelRow)
						if err := f.SetCellFormula(sheetName, cell, formula); err != nil {
							return err
						}
					} else {
						val := rowData[field]
						if err := f.SetCellValue(sheetName, cell, val); err != nil {
							return err
						}
					}

					if table.CellStyle != nil {
						if err := f.SetCellStyle(sheetName, cell, cell, cellStyleID); err != nil {
							return err
						}
					}
				}
			}
		}

		if table.AutoFit {
			endCol := startCol + len(table.Headers) - 1
			endRow := startRow + len(rows)
			for c := startCol; c <= endCol; c++ {
				colName, _ := excelize.ColumnNumberToName(c)
				if err := f.SetColWidth(sheetName, colName, colName, 15); err != nil {
					return err
				}
			}
			if err := autoFitColumns(f, sheetName, startCol, endCol, startRow, endRow); err != nil {
				return err
			}
		}
	}
	return nil
}

func replaceRowPlaceholder(formula string, rowNum int) string {
	result := formula
	for strings.Contains(result, "{row}") {
		result = strings.Replace(result, "{row}", strconv.Itoa(rowNum), 1)
	}
	return result
}

func applyConditionals(f *excelize.File, sheetName string, conditionals []template.ConditionalFormat) error {
	for _, cond := range conditionals {
		style := &excelize.Style{}

		if cond.FontColor != "" {
			style.Font = &excelize.Font{Color: cond.FontColor}
		}
		if cond.FillColor != "" {
			style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{cond.FillColor}}
		}

		styleID, err := f.NewStyle(style)
		if err != nil {
			return err
		}

		switch cond.Condition {
		case "greater_than", "less_than", "equal", "greater_equal", "less_equal":
		default:
			continue
		}

		if err := f.SetConditionalFormat(sheetName, cond.Range, []excelize.ConditionalFormatOptions{
			{
				Type:     "cell",
				Criteria: ">",
				Value:    cond.Threshold,
				Format:   styleID,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func extractFirstCell(rangeStr string) string {
	parts := strings.Split(rangeStr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return rangeStr
}

func applyChart(f *excelize.File, sheetName string, chart *template.ChartConfig) error {
	if strings.TrimSpace(chart.DataRange) == "" {
		return fmt.Errorf("empty data range")
	}
	if !isValidRange(chart.DataRange) {
		return fmt.Errorf("invalid data range: %s", chart.DataRange)
	}
	if chart.CategoryRange != "" && !isValidRange(chart.CategoryRange) {
		return fmt.Errorf("invalid category range: %s", chart.CategoryRange)
	}

	var chartType excelize.ChartType
	switch strings.ToLower(chart.Type) {
	case "bar":
		chartType = excelize.Col
	case "line":
		chartType = excelize.Line
	case "pie":
		chartType = excelize.Pie
	default:
		return fmt.Errorf("unsupported chart type: %s", chart.Type)
	}

	series := excelize.ChartSeries{
		Name:   chart.Title,
		Values: fmt.Sprintf("'%s'!%s", sheetName, chart.DataRange),
	}
	if chart.CategoryRange != "" {
		series.Categories = fmt.Sprintf("'%s'!%s", sheetName, chart.CategoryRange)
	}

	chartOpts := &excelize.Chart{
		Type: chartType,
		Title: []excelize.RichTextRun{
			{Text: chart.Title},
		},
		Series: []excelize.ChartSeries{series},
	}

	if err := f.AddChart(sheetName, chart.Position, chartOpts); err != nil {
		return fmt.Errorf("failed to add chart: %w", err)
	}

	return nil
}

func isValidRange(rangeStr string) bool {
	parts := strings.Split(rangeStr, ":")
	if len(parts) != 2 {
		return false
	}

	_, startRow, err := excelize.CellNameToCoordinates(parts[0])
	if err != nil {
		return false
	}
	_, endRow, err := excelize.CellNameToCoordinates(parts[1])
	if err != nil {
		return false
	}

	return startRow <= endRow
}

func createStyle(f *excelize.File, cs *template.CellStyle) (int, error) {
	style := &excelize.Style{}

	if cs.FontSize > 0 || cs.Bold || cs.Italic || cs.FontColor != "" {
		style.Font = &excelize.Font{
			Bold:   cs.Bold,
			Italic: cs.Italic,
			Size:   cs.FontSize,
			Color:  cs.FontColor,
		}
	}

	if cs.FillColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{cs.FillColor},
		}
	}

	if cs.Alignment != "" {
		align := &excelize.Alignment{}
		switch cs.Alignment {
		case "center":
			align.Horizontal = "center"
			align.Vertical = "center"
		case "left":
			align.Horizontal = "left"
		case "right":
			align.Horizontal = "right"
		}
		style.Alignment = align
	}

	if cs.Border != "" {
		style.Border = []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		}
	}

	return f.NewStyle(style)
}

func autoFitColumns(f *excelize.File, sheetName string, startCol, endCol, startRow, endRow int) error {
	for c := startCol; c <= endCol; c++ {
		colName, _ := excelize.ColumnNumberToName(c)
		maxLen := 0.0

		for r := startRow; r <= endRow; r++ {
			cell, _ := excelize.CoordinatesToCellName(c, r)
			val, err := f.GetCellValue(sheetName, cell)
			if err != nil {
				continue
			}
			l := float64(len([]rune(val)))
			if l > maxLen {
				maxLen = l
			}
		}

		if maxLen > 0 {
			width := maxLen*1.2 + 2
			if width > 50 {
				width = 50
			}
			if err := f.SetColWidth(sheetName, colName, colName, width); err != nil {
				return err
			}
		}
	}
	return nil
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
