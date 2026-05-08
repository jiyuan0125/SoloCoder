package wsframe

import "errors"

var (
	ErrInsufficientData      = errors.New("insufficient data to parse frame")
	ErrInvalidOpcode         = errors.New("invalid opcode")
	ErrControlFrameFragmented = errors.New("control frame must not be fragmented")
	ErrControlFrameTooLarge  = errors.New("control frame payload exceeds 125 bytes")
	ErrReservedBitSet        = errors.New("reserved bits set but no extension negotiated")
	ErrInvalidUTF8           = errors.New("invalid UTF-8 encoding in text frame")
	ErrMessageTooLarge       = errors.New("message exceeds maximum size")
	ErrUnexpectedContinuation = errors.New("unexpected continuation frame")
	ErrMissingContinuation   = errors.New("expected continuation frame but received data frame")
)
