package dns

import "strings"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type ServerError struct {
	Message string
}

func (e *ServerError) Error() string {
	return e.Message
}

type RecordTypeError struct {
	Supported []string
}

func (e *RecordTypeError) Error() string {
	return "unsupported record type. Supported: " + strings.Join(e.Supported, ", ")
}

type BatchLimitError struct{}

func (e *BatchLimitError) Error() string {
	return "batch query limit exceeded. Maximum 50 domains allowed"
}
