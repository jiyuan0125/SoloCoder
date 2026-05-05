package csvjoin

import (
	"errors"
	"fmt"
)

var (
	// ErrInsufficientFiles is returned when fewer than 2 files are provided for joining.
	ErrInsufficientFiles = errors.New("at least 2 files required for join")
	// ErrFileNotFound is returned when a specified file cannot be found.
	ErrFileNotFound = errors.New("file not found")
	// ErrColumnNotFound is returned when the join key column is not found in a file.
	ErrColumnNotFound = errors.New("join key column not found")
	// ErrRead is returned when an error occurs while reading a CSV file.
	ErrRead = errors.New("error reading CSV file")
	// ErrWrite is returned when an error occurs while writing a CSV file.
	ErrWrite = errors.New("error writing CSV file")
	// ErrEncoding is returned when an error occurs during encoding detection or conversion.
	ErrEncoding = errors.New("encoding error")
)

// JoinError represents an error that occurred during a join operation.
type JoinError struct {
	Op  string // The operation that failed
	Err error  // The underlying error
}

// Error returns the error message.
func (e *JoinError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *JoinError) Unwrap() error {
	return e.Err
}

// ColumnNotFoundError is returned when a specified column is not found.
type ColumnNotFoundError struct {
	ColumnName string
	FilePath   string
}

// Error returns the error message.
func (e *ColumnNotFoundError) Error() string {
	if e.FilePath == "" {
		return fmt.Sprintf("column %q not found", e.ColumnName)
	}
	return fmt.Sprintf("column %q not found in file %q", e.ColumnName, e.FilePath)
}
