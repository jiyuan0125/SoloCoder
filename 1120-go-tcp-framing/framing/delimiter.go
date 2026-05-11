package framing

import (
	"errors"
)

const DefaultEscapeByte byte = 0x1B

var ErrEmptyDelimiter = errors.New("framing: delimiter cannot be empty")

type DelimiterConfig struct {
	Delimiter    []byte
	EscapeByte   byte
	MaxFrameSize int
	delimiterSet map[byte]struct{}
}

func NewDelimiterConfig(delimiter []byte, maxFrameSize int) (*DelimiterConfig, error) {
	return NewDelimiterConfigWithEscape(delimiter, DefaultEscapeByte, maxFrameSize)
}

func NewDelimiterConfigWithEscape(delimiter []byte, escapeByte byte, maxFrameSize int) (*DelimiterConfig, error) {
	if len(delimiter) == 0 {
		return nil, ErrEmptyDelimiter
	}
	if maxFrameSize <= 0 {
		maxFrameSize = 16 * 1024 * 1024
	}
	delimiterSet := make(map[byte]struct{})
	for _, b := range delimiter {
		delimiterSet[b] = struct{}{}
	}
	return &DelimiterConfig{
		Delimiter:    append([]byte{}, delimiter...),
		EscapeByte:   escapeByte,
		MaxFrameSize: maxFrameSize,
		delimiterSet: delimiterSet,
	}, nil
}

func (c *DelimiterConfig) needsEscape(b byte) bool {
	if b == c.EscapeByte {
		return true
	}
	_, ok := c.delimiterSet[b]
	return ok
}

func (c *DelimiterConfig) Encode(message []byte) ([]byte, error) {
	escapedLen := 0
	for _, b := range message {
		if c.needsEscape(b) {
			escapedLen += 2
		} else {
			escapedLen++
		}
	}
	if escapedLen > c.MaxFrameSize {
		return nil, ErrFrameTooLarge
	}
	frame := make([]byte, 0, escapedLen+len(c.Delimiter))
	for _, b := range message {
		if c.needsEscape(b) {
			frame = append(frame, c.EscapeByte)
		}
		frame = append(frame, b)
	}
	frame = append(frame, c.Delimiter...)
	return frame, nil
}

func (c *DelimiterConfig) Decode(buffer *RingBuffer) ([][]byte, error) {
	var messages [][]byte
	for {
		pos := buffer.IndexWithEscape(c.Delimiter, c.EscapeByte)
		if pos == -1 {
			if buffer.Length() > c.MaxFrameSize*2 {
				return nil, ErrFrameTooLarge
			}
			break
		}
		if pos > c.MaxFrameSize*2 {
			return nil, ErrFrameTooLarge
		}
		escapedMsg, _ := buffer.Read(pos)
		buffer.Discard(len(c.Delimiter))
		message := c.unescape(escapedMsg)
		if len(message) > c.MaxFrameSize {
			return nil, ErrFrameTooLarge
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (c *DelimiterConfig) unescape(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	result := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == c.EscapeByte && i+1 < len(data) {
			result = append(result, data[i+1])
			i++
		} else {
			result = append(result, data[i])
		}
	}
	return result
}
