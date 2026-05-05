// Package api defines the request and response types for the CSV join service.
package api

// JoinType specifies the type of join operation.
type JoinType string

const (
	// InnerJoin returns only rows with matching keys in all files.
	InnerJoin JoinType = "inner"
	// LeftJoin returns all rows from the left file, with matching rows from the right files.
	LeftJoin JoinType = "left"
)

// FileConfig represents the configuration for a single CSV file.
type FileConfig struct {
	// Path is the file path to the CSV file.
	Path string `json:"path"`
	// SkipHeader indicates if the first row should be skipped (treated as data, not header).
	SkipHeader bool `json:"skip_header"`
	// ColumnNames specifies custom column names when SkipHeader is true.
	// If SkipHeader is true but ColumnNames is empty, default names (col0, col1, ...) are used.
	ColumnNames []string `json:"column_names,omitempty"`
	// FileSuffix is the suffix added to column names to distinguish them from other files.
	// If empty, a default suffix (_f0, _f1, ...) is used.
	FileSuffix string `json:"file_suffix,omitempty"`
}

// JoinRequest represents a request to join multiple CSV files.
type JoinRequest struct {
	// Files is a list of CSV file configurations to join.
	Files []FileConfig `json:"files"`
	// JoinKey is the column name used for joining (after applying any suffixes).
	// For multiple files, the join key is expected to exist in all files.
	JoinKey string `json:"join_key"`
	// JoinType specifies the type of join operation (inner or left).
	JoinType JoinType `json:"join_type"`
	// TrimKeys indicates if whitespace should be trimmed from join key values.
	// When true, "ABC" and " ABC" are considered the same key.
	TrimKeys bool `json:"trim_keys"`
}

// JoinResponse represents the response from a CSV join operation.
type JoinResponse struct {
	// Success indicates if the join operation was successful.
	Success bool `json:"success"`
	// Error contains the error message if Success is false.
	Error string `json:"error,omitempty"`
	// OutputPath is the path to the resulting CSV file if Success is true.
	OutputPath string `json:"output_path,omitempty"`
	// RowCount is the number of rows in the resulting CSV.
	RowCount int `json:"row_count,omitempty"`
	// Columns is the list of column names in the resulting CSV.
	Columns []string `json:"columns,omitempty"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}
