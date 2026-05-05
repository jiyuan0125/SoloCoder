package xlsxreader

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	excelEpoch1900 = "1899-12-30"
	excelEpoch1904 = "1904-01-01"
)

func Open(filename string) (*XlsxReader, error) {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}

	reader := &XlsxReader{
		zipReader:     zipReader,
		sheetMap:      make(map[string]string),
		customNumFmts: make(map[int]string),
	}

	if err := reader.checkEncryption(); err != nil {
		zipReader.Close()
		return nil, err
	}

	if err := reader.parseWorkbook(); err != nil {
		zipReader.Close()
		return nil, err
	}

	if err := reader.parseSharedStrings(); err != nil {
		zipReader.Close()
		return nil, err
	}

	if err := reader.parseStyles(); err != nil {
		zipReader.Close()
		return nil, err
	}

	return reader, nil
}

func (r *XlsxReader) Close() error {
	return r.zipReader.Close()
}

func (r *XlsxReader) checkEncryption() error {
	for _, f := range r.zipReader.File {
		if f.Name == "EncryptionInfo" || strings.HasPrefix(f.Name, "EncryptedPackage") {
			return &EncryptedFileError{}
		}
	}
	return nil
}

func (r *XlsxReader) parseWorkbook() error {
	workbookFile := r.findFile("xl/workbook.xml")
	if workbookFile == nil {
		return fmt.Errorf("workbook.xml not found")
	}

	rc, err := workbookFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	var workbook Workbook
	if err := xml.Unmarshal(data, &workbook); err != nil {
		return err
	}

	r.workbook = &workbook

	for i, sheet := range workbook.Sheets.Sheet {
		r.sheetMap[sheet.Name] = fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1)
	}

	return nil
}

func (r *XlsxReader) parseSharedStrings() error {
	sharedStringsFile := r.findFile("xl/sharedStrings.xml")
	if sharedStringsFile == nil {
		r.sharedStrings = []string{}
		return nil
	}

	rc, err := sharedStringsFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	var sharedStrings SharedStrings
	if err := xml.Unmarshal(data, &sharedStrings); err != nil {
		return err
	}

	r.sharedStrings = make([]string, len(sharedStrings.SI))
	for i, si := range sharedStrings.SI {
		r.sharedStrings[i] = si.T
	}

	return nil
}

func (r *XlsxReader) parseStyles() error {
	stylesFile := r.findFile("xl/styles.xml")
	if stylesFile == nil {
		r.styles = &Styles{}
		return nil
	}

	rc, err := stylesFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	var styles Styles
	if err := xml.Unmarshal(data, &styles); err != nil {
		return err
	}

	r.styles = &styles

	for _, nf := range styles.NumFmts.NumFmt {
		id, err := strconv.Atoi(nf.NumFmtID)
		if err == nil {
			r.customNumFmts[id] = nf.FormatCode
		}
	}

	return nil
}

func (r *XlsxReader) findFile(name string) *zip.File {
	for _, f := range r.zipReader.File {
		if f.Name == name || filepath.Base(f.Name) == filepath.Base(name) {
			return f
		}
	}
	return nil
}

func (r *XlsxReader) GetSheetNames() []string {
	names := make([]string, 0, len(r.workbook.Sheets.Sheet))
	for _, sheet := range r.workbook.Sheets.Sheet {
		names = append(names, sheet.Name)
	}
	return names
}

func (r *XlsxReader) ReadSheet(sheetName string, options ReadOptions) ([][]string, error) {
	sheetPath, ok := r.sheetMap[sheetName]
	if !ok {
		return nil, &SheetNotFoundError{
			RequestedSheet:   sheetName,
			AvailableSheets: r.GetSheetNames(),
		}
	}

	sheetFile := r.findFile(sheetPath)
	if sheetFile == nil {
		return nil, fmt.Errorf("worksheet file not found: %s", sheetPath)
	}

	rc, err := sheetFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	var worksheet Worksheet
	if err := xml.Unmarshal(data, &worksheet); err != nil {
		return nil, err
	}

	mergedCells := r.parseMergeCells(worksheet.MergeCells)

	return r.processRows(&worksheet, mergedCells, options), nil
}

func (r *XlsxReader) parseMergeCells(mergeCells MergeCells) map[string]string {
	merged := make(map[string]string)

	for _, mc := range mergeCells.MergeCell {
		parts := strings.Split(mc.Ref, ":")
		if len(parts) == 2 {
			startCell := parts[0]
			merged[startCell] = startCell

			startCol, startRow := r.cellRefToIndices(startCell)
			endCol, endRow := r.cellRefToIndices(parts[1])

			for row := startRow; row <= endRow; row++ {
				for col := startCol; col <= endCol; col++ {
					cellRef := r.indicesToCellRef(col, row)
					if cellRef != startCell {
						merged[cellRef] = startCell
					}
				}
			}
		}
	}

	return merged
}

func (r *XlsxReader) cellRefToIndices(ref string) (int, int) {
	colStr := ""
	rowStr := ""

	for _, c := range ref {
		if c >= 'A' && c <= 'Z' {
			colStr += string(c)
		} else {
			rowStr += string(c)
		}
	}

	col := 0
	for _, c := range colStr {
		col = col*26 + int(c-'A'+1)
	}

	row, _ := strconv.Atoi(rowStr)

	return col, row
}

func (r *XlsxReader) indicesToCellRef(col, row int) string {
	colStr := ""
	for col > 0 {
		col--
		colStr = string('A'+col%26) + colStr
		col /= 26
	}
	return colStr + strconv.Itoa(row)
}

func (r *XlsxReader) processRows(worksheet *Worksheet, mergedCells map[string]string, options ReadOptions) [][]string {
	result := [][]string{}

	maxCol := 0
	for _, row := range worksheet.SheetData.Rows {
		for _, cell := range row.Cells {
			col, _ := r.cellRefToIndices(cell.R)
			if col > maxCol {
				maxCol = col
			}
		}
	}

	for _, row := range worksheet.SheetData.Rows {
		rowNum := row.R

		if options.StartRow > 0 && rowNum < options.StartRow {
			continue
		}
		if options.EndRow > 0 && rowNum > options.EndRow {
			continue
		}

		rowData := make([]string, maxCol)
		isEmpty := true

		for _, cell := range row.Cells {
			col, _ := r.cellRefToIndices(cell.R)
			if col-1 >= 0 && col-1 < len(rowData) {
				value := r.getCellValue(&cell)

				if masterCell, isMerged := mergedCells[cell.R]; isMerged {
					if masterCell != cell.R {
						value = ""
					}
				}

				rowData[col-1] = value
				if value != "" {
					isEmpty = false
				}
			}
		}

		if options.SkipEmptyRows && isEmpty {
			continue
		}

		result = append(result, rowData)
	}

	return result
}

func (r *XlsxReader) getCellValue(cell *Cell) string {
	switch cell.T {
	case "s":
		if cell.V == "" {
			return ""
		}
		idx, err := strconv.Atoi(cell.V)
		if err == nil && idx >= 0 && idx < len(r.sharedStrings) {
			return r.sharedStrings[idx]
		}
		return cell.V

	case "str":
		return cell.V

	case "inlineStr":
		return cell.Is.T

	case "b":
		if cell.V == "" {
			return ""
		}
		if cell.V == "1" {
			return "TRUE"
		}
		return "FALSE"

	case "n":
		fallthrough

	default:
		if cell.V == "" {
			return ""
		}
		numFmtID := r.getNumFmtID(cell)
		if r.isDateFormat(numFmtID) {
			return r.formatDate(cell.V)
		}
		return r.formatNumber(cell.V)
	}
}

func (r *XlsxReader) getNumFmtID(cell *Cell) int {
	if cell.S == "" {
		return 0
	}

	styleIdx, err := strconv.Atoi(cell.S)
	if err != nil {
		return 0
	}

	if r.styles == nil || styleIdx >= len(r.styles.CellXfs.Xf) {
		return 0
	}

	numFmtIDStr := r.styles.CellXfs.Xf[styleIdx].NumFmtID
	if numFmtIDStr == "" {
		return 0
	}

	numFmtID, err := strconv.Atoi(numFmtIDStr)
	if err != nil {
		return 0
	}

	return numFmtID
}

func (r *XlsxReader) isDateFormat(numFmtID int) bool {
	if numFmtID >= 14 && numFmtID <= 22 {
		return true
	}

	if formatCode, ok := r.customNumFmts[numFmtID]; ok {
		lower := strings.ToLower(formatCode)
		if strings.Contains(lower, "yy") || strings.Contains(lower, "mm") || 
		   strings.Contains(lower, "dd") || strings.Contains(lower, "h:") ||
		   strings.Contains(lower, ":mm") || strings.Contains(lower, "ss") {
			return true
		}
	}

	return false
}

func (r *XlsxReader) formatDate(value string) string {
	excelDate, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}

	days := int(excelDate)
	frac := excelDate - float64(days)

	epoch, _ := time.Parse("2006-01-02", excelEpoch1900)
	date := epoch.AddDate(0, 0, days)

	hours := int(frac * 24)
	minutes := int((frac*24 - float64(hours)) * 60)
	seconds := int(((frac*24-float64(hours))*60 - float64(minutes)) * 60)

	date = date.Add(time.Hour*time.Duration(hours) + 
		time.Minute*time.Duration(minutes) + 
		time.Second*time.Duration(seconds))

	if hours == 0 && minutes == 0 && seconds == 0 {
		return date.Format("2006-01-02")
	}

	return date.Format("2006-01-02 15:04:05")
}

func (r *XlsxReader) formatNumber(value string) string {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}

	if num == float64(int64(num)) {
		return strconv.FormatInt(int64(num), 10)
	}

	return strconv.FormatFloat(num, 'f', -1, 64)
}
