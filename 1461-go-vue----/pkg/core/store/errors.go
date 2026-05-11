package store

import "errors"

var (
	ErrBatchExists            = errors.New("batch already exists")
	ErrBatchNotFound          = errors.New("batch not found")
	ErrBatchImmutable         = errors.New("batch id or product name cannot be modified")
	ErrProcessFlowNotFound    = errors.New("process flow not found for product")
	ErrProcessRecordNotFound  = errors.New("process record not found")
	ErrInspectionSpecNotFound = errors.New("inspection spec not found")
	ErrInvalidOperation       = errors.New("invalid operation")
)
