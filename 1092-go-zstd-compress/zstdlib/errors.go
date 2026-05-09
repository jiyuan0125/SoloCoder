package zstdlib

import "errors"

var (
	ErrEmptyDictionary   = errors.New("dictionary cannot be empty")
	ErrDictionaryNotFound = errors.New("dictionary not found")
	ErrInvalidDictionary = errors.New("invalid dictionary")
	ErrDictionaryMismatch = errors.New("dictionary mismatch")
	ErrCorruptedData     = errors.New("corrupted compressed data")
	ErrFileNotFound      = errors.New("file not found")
	ErrEmptyFile         = errors.New("empty file")
)
