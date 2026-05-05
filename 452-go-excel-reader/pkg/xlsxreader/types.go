package xlsxreader

import (
	"archive/zip"
	"encoding/xml"
)

type Workbook struct {
	XMLName xml.Name `xml:"workbook"`
	Sheets  Sheets   `xml:"sheets"`
}

type Sheets struct {
	Sheet []Sheet `xml:"sheet"`
}

type Sheet struct {
	Name    string `xml:"name,attr"`
	SheetID string `xml:"sheetId,attr"`
	ID      string `xml:"id,attr"`
}

type Worksheet struct {
	XMLName     xml.Name    `xml:"worksheet"`
	SheetData   SheetData   `xml:"sheetData"`
	MergeCells  MergeCells  `xml:"mergeCells"`
}

type SheetData struct {
	Rows []Row `xml:"row"`
}

type Row struct {
	R    int    `xml:"r,attr"`
	Spans string `xml:"spans,attr"`
	Cells []Cell `xml:"c"`
}

type Cell struct {
	R       string    `xml:"r,attr"`
	T       string    `xml:"t,attr"`
	S       string    `xml:"s,attr"`
	V       string    `xml:"v"`
	Is      InlineStr `xml:"is"`
}

type InlineStr struct {
	T string `xml:"t"`
}

type MergeCells struct {
	Count      int         `xml:"count,attr"`
	MergeCell  []MergeCell `xml:"mergeCell"`
}

type MergeCell struct {
	Ref string `xml:"ref,attr"`
}

type SharedStrings struct {
	XMLName xml.Name `xml:"sst"`
	SI      []SI     `xml:"si"`
}

type SI struct {
	T string `xml:"t"`
}

type Styles struct {
	XMLName    xml.Name   `xml:"styleSheet"`
	NumFmts    NumFmts    `xml:"numFmts"`
	CellXfs    CellXfs    `xml:"cellXfs"`
}

type NumFmts struct {
	NumFmt []NumFmt `xml:"numFmt"`
}

type NumFmt struct {
	NumFmtID   string `xml:"numFmtId,attr"`
	FormatCode string `xml:"formatCode,attr"`
}

type CellXfs struct {
	Count int  `xml:"count,attr"`
	Xf    []Xf `xml:"xf"`
}

type Xf struct {
	NumFmtID string `xml:"numFmtId,attr"`
}

type XlsxReader struct {
	zipReader      *zip.ReadCloser
	workbook       *Workbook
	sharedStrings  []string
	styles         *Styles
	sheetMap       map[string]string
	customNumFmts  map[int]string
}

type ReadOptions struct {
	StartRow      int
	EndRow        int
	SkipEmptyRows bool
}

type SheetNotFoundError struct {
	RequestedSheet string
	AvailableSheets []string
}

func (e *SheetNotFoundError) Error() string {
	msg := "sheet not found: " + e.RequestedSheet
	if len(e.AvailableSheets) > 0 {
		msg += ". Available sheets: "
		for i, s := range e.AvailableSheets {
			if i > 0 {
				msg += ", "
			}
			msg += s
		}
	}
	return msg
}

type EncryptedFileError struct{}

func (e *EncryptedFileError) Error() string {
	return "the Excel file is encrypted and requires a password to open"
}
