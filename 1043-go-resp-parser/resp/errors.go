package resp

import (
	"fmt"
)

type ProtocolError struct {
	Pos     int64
	Message string
}

func NewProtocolError(pos int64, format string, args ...interface{}) *ProtocolError {
	return &ProtocolError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	}
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("protocol error at byte %d: %s", e.Pos, e.Message)
}
