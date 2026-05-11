package common

import "fmt"

type DeclarationType string

const (
	TypeFunction DeclarationType = "function"
	TypeType     DeclarationType = "type"
	TypeVariable DeclarationType = "variable"
	TypeConstant DeclarationType = "constant"
	TypeField    DeclarationType = "field"
)

type ConfidenceLevel string

const (
	ConfidenceCertain  ConfidenceLevel = "certain"
	ConfidencePossible ConfidenceLevel = "possible"
)

type ReferenceLocation struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	LineText string `json:"line_text"`
}

type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type Declaration struct {
	Name        string          `json:"name"`
	Type        DeclarationType `json:"type"`
	Location    Location        `json:"location"`
	Snippet     string          `json:"snippet"`
	IsExported  bool            `json:"is_exported"`
	IsAlias     bool            `json:"is_alias"`
	IsInterface bool            `json:"is_interface"`
}

type DeadCodeEntry struct {
	Declaration Declaration     `json:"declaration"`
	Reason      string          `json:"reason"`
	Confidence  ConfidenceLevel `json:"confidence"`
}

type Summary struct {
	UnusedFunctions int `json:"unused_functions"`
	UnusedTypes     int `json:"unused_types"`
	UnusedVariables int `json:"unused_variables"`
	UnusedConstants int `json:"unused_constants"`
	UnusedFields    int `json:"unused_fields"`
}

type AnalysisReport struct {
	PackageName string          `json:"package_name"`
	Entries     []DeadCodeEntry `json:"entries"`
	Summary     Summary         `json:"summary"`
}

type UploadFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type AnalyzeRequest struct {
	Files       []UploadFile `json:"files"`
	PackageName string       `json:"package_name"`
}

type AnalyzeResponse struct {
	Report   AnalysisReport `json:"report"`
	ReportID string         `json:"report_id"`
	Success  bool           `json:"success"`
	Error    string         `json:"error,omitempty"`
}

type GetReportRequest struct {
	ReportID string `json:"report_id"`
}

type GetReportResponse struct {
	Report  AnalysisReport `json:"report"`
	Success bool           `json:"success"`
	Error   string         `json:"error,omitempty"`
}

type FilterRequest struct {
	ReportID string          `json:"report_id"`
	Filter   DeclarationType `json:"filter"`
}

type FilterResponse struct {
	Report  AnalysisReport `json:"report"`
	Success bool           `json:"success"`
	Error   string         `json:"error,omitempty"`
}

type ReferencePathRequest struct {
	ReportID string `json:"report_id"`
	Name     string `json:"name"`
}

type ReferencePathResponse struct {
	Name        string              `json:"name"`
	References  []ReferenceLocation `json:"references"`
	Declaration Location            `json:"declaration"`
	Success     bool                `json:"success"`
	Error       string              `json:"error,omitempty"`
}

func (s Summary) String() string {
	return fmt.Sprintf(
		"Functions: %d, Types: %d, Variables: %d, Constants: %d, Fields: %d",
		s.UnusedFunctions, s.UnusedTypes, s.UnusedVariables, s.UnusedConstants, s.UnusedFields,
	)
}

func (c ConfidenceLevel) String() string {
	switch c {
	case ConfidenceCertain:
		return "确定是死代码"
	case ConfidencePossible:
		return "可能是死代码"
	default:
		return "未知"
	}
}

func (d DeclarationType) String() string {
	switch d {
	case TypeFunction:
		return "函数"
	case TypeType:
		return "类型"
	case TypeVariable:
		return "变量"
	case TypeConstant:
		return "常量"
	case TypeField:
		return "字段"
	default:
		return "未知"
	}
}
