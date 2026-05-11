package core

import "errors"

var (
	ErrInvalidISBN       = errors.New("invalid ISBN")
	ErrInvalidYear       = errors.New("invalid publication year")
	ErrEmptyTitle        = errors.New("title cannot be empty")
	ErrEmptyAuthor       = errors.New("author cannot be empty")
	ErrInvalidLineFormat = errors.New("invalid line format")
)
