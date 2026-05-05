// Package csvjoin provides functionality to join multiple CSV files based on a common key column.
// It supports inner joins and left joins, with memory-efficient processing for large files.
package csvjoin

// JoinType specifies the type of join operation.
type JoinType int

const (
	// InnerJoin returns only rows with matching keys in all files.
	InnerJoin JoinType = iota
	// LeftJoin returns all rows from the left file, with matching rows from the right files.
	// Non-matching rows from right files are filled with empty values.
	LeftJoin
)

// FileOptions represents the options for reading a single CSV file.
type FileOptions struct {
	// Path is the file path to the CSV file.
	Path string
	// SkipHeader indicates if the first row should be treated as data (not header).
	// When true, the first row is included in the data, and ColumnNames should be provided.
	SkipHeader bool
	// ColumnNames specifies custom column names when SkipHeader is true.
	// If SkipHeader is true but ColumnNames is empty, default names (col0, col1, ...) are used.
	ColumnNames []string
	// FileSuffix is the suffix added to column names to distinguish them from other files.
	// If empty, a default suffix (_f0, _f1, ...) is used based on file order.
	FileSuffix string
}

// JoinOptions represents the options for a join operation.
type JoinOptions struct {
	// Files is a list of file options for each CSV file to join.
	Files []FileOptions
	// JoinKey is the column name used for joining.
	// This should be the original column name (before suffixes are applied).
	JoinKey string
	// JoinType specifies the type of join operation.
	JoinType JoinType
	// TrimKeys indicates if whitespace should be trimmed from join key values.
	// When true, "ABC" and " ABC" are considered the same key.
	TrimKeys bool
}

// Row represents a single row of CSV data with column names.
type Row struct {
	// Columns is the list of column names in order.
	Columns []string
	// Values is the list of values corresponding to Columns.
	Values []string
}

// CSVReader wraps the standard CSV reader with additional functionality.
type CSVReader interface {
	// Read reads a single row from the CSV file.
	// Returns io.EOF when there are no more rows.
	Read() ([]string, error)
	// ReadAll reads all remaining rows from the CSV file.
	ReadAll() ([][]string, error)
	// Columns returns the column names of the CSV file.
	Columns() []string
	// Close closes the underlying reader.
	Close() error
}

// CSVWriter wraps the standard CSV writer with additional functionality.
type CSVWriter interface {
	// Write writes a single row to the CSV file.
	Write(record []string) error
	// WriteAll writes all rows to the CSV file.
	WriteAll(records [][]string) error
	// Flush flushes any buffered data to the underlying writer.
	Flush()
	// Error reports any error that occurred during a previous Write or Flush.
	Error() error
	// Close closes the underlying writer.
	Close() error
}

// Joiner performs the join operation on CSV files.
type Joiner struct {
	options JoinOptions
}

// NewJoiner creates a new Joiner with the given options.
func NewJoiner(options JoinOptions) (*Joiner, error) {
	if len(options.Files) < 2 {
		return nil, &JoinError{
			Op:  "NewJoiner",
			Err: ErrInsufficientFiles,
		}
	}
	return &Joiner{options: options}, nil
}

// Options returns the join options.
func (j *Joiner) Options() JoinOptions {
	return j.options
}
