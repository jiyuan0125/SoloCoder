package api

type CoverageMode string

const (
	ModeSet    CoverageMode = "set"
	ModeCount  CoverageMode = "count"
	ModeAtomic CoverageMode = "atomic"
)

type UploadRequest struct {
	Content string `json:"content"`
}

type UploadResponse struct {
	ID      string `json:"id"`
	Mode    string `json:"mode"`
	Message string `json:"message"`
}

type FunctionCoverage struct {
	Package      string  `json:"package"`
	Function     string  `json:"function"`
	CoveredLines int     `json:"covered_lines"`
	TotalLines   int     `json:"total_lines"`
	Percentage   float64 `json:"percentage"`
}

type FunctionCoverageReport struct {
	Functions []FunctionCoverage `json:"functions"`
}

type PackageCoverage struct {
	Package      string  `json:"package"`
	CoveredLines int     `json:"covered_lines"`
	TotalLines   int     `json:"total_lines"`
	Percentage   float64 `json:"percentage"`
}

type PackageCoverageReport struct {
	Packages []PackageCoverage `json:"packages"`
}

type LineCoverage struct {
	LineNumber   int    `json:"line_number"`
	Status       string `json:"status"`
	ExecutionCount int  `json:"execution_count,omitempty"`
}

type FileLineCoverage struct {
	File  string         `json:"file"`
	Lines []LineCoverage `json:"lines"`
}

type LineCoverageReport struct {
	Files []FileLineCoverage `json:"files"`
}

type SummaryResponse struct {
	TotalLines        int              `json:"total_lines"`
	CoveredLines      int              `json:"covered_lines"`
	UncoveredLines    int              `json:"uncovered_lines"`
	Percentage        float64          `json:"percentage"`
	ConfidenceLower   float64          `json:"confidence_lower,omitempty"`
	ConfidenceUpper   float64          `json:"confidence_upper,omitempty"`
	LowestPackages    []PackageCoverage `json:"lowest_packages"`
	HighestPackages   []PackageCoverage `json:"highest_packages"`
}

type DiffFunction struct {
	Package          string  `json:"package"`
	Function         string  `json:"function"`
	OldPercentage    float64 `json:"old_percentage"`
	NewPercentage    float64 `json:"new_percentage"`
	Change           float64 `json:"change"`
	IsDecrease       bool    `json:"is_decrease"`
}

type DiffReport struct {
	OldID       string         `json:"old_id"`
	NewID       string         `json:"new_id"`
	OldSummary  SummaryResponse `json:"old_summary"`
	NewSummary  SummaryResponse `json:"new_summary"`
	Changed     []DiffFunction `json:"changed"`
	Decreased   []DiffFunction `json:"decreased"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Line    int    `json:"line,omitempty"`
	Details string `json:"details,omitempty"`
}
