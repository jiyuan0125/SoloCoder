package server

import "errors"

var (
	ErrListNotFound           = errors.New("mailing list not found")
	ErrSubscriberNotFound     = errors.New("subscriber not found")
	ErrTemplateNotFound       = errors.New("template not found")
	ErrTaskNotFound           = errors.New("task not found")
	ErrInvalidEmailFormat     = errors.New("invalid email format")
	ErrDuplicateSend          = errors.New("duplicate send within 24 hours")
	ErrListPaused             = errors.New("mailing list is paused")
	ErrNoActiveSubscribers    = errors.New("no active subscribers in list")
	ErrABTestRequiresSubjects = errors.New("A/B test requires both subject_a and subject_b")
	ErrInvalidBatchSize       = errors.New("invalid batch size")
	ErrNameRequired           = errors.New("name is required")
	ErrMethodNotAllowed       = errors.New("method not allowed")
)
