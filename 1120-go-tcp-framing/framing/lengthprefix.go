package framing

import (
	"encoding/binary"
	"errors"
)

var (
	ErrFrameTooLarge  = errors.New("framing: frame size exceeds maximum")
	ErrInvalidHeader  = errors.New("framing: invalid header size")
)

type LengthPrefixConfig struct {
	HeaderSize   int
	ByteOrder    binary.ByteOrder
	MaxFrameSize int
}

func NewLengthPrefixConfig(headerSize int, byteOrder binary.ByteOrder, maxFrameSize int) (*LengthPrefixConfig, error) {
	if headerSize != 2 && headerSize != 4 {
		return nil, ErrInvalidHeader
	}
	if maxFrameSize <= 0 {
		maxFrameSize = 16 * 1024 * 1024
	}
	return &LengthPrefixConfig{
		HeaderSize:   headerSize,
		ByteOrder:    byteOrder,
		MaxFrameSize: maxFrameSize,
	}, nil
}

func (c *LengthPrefixConfig) Encode(message []byte) ([]byte, error) {
	if len(message) > c.MaxFrameSize {
		return nil, ErrFrameTooLarge
	}
	frame := make([]byte, c.HeaderSize+len(message))
	switch c.HeaderSize {
	case 2:
		c.ByteOrder.PutUint16(frame[:2], uint16(len(message)))
	case 4:
		c.ByteOrder.PutUint32(frame[:4], uint32(len(message)))
	}
	copy(frame[c.HeaderSize:], message)
	return frame, nil
}

func (c *LengthPrefixConfig) Decode(buffer *RingBuffer) ([][]byte, error) {
	var messages [][]byte
	for {
		if buffer.Length() < c.HeaderSize {
			break
		}
		header, _ := buffer.Peek(c.HeaderSize)
		var payloadLen int
		switch c.HeaderSize {
		case 2:
			payloadLen = int(c.ByteOrder.Uint16(header))
		case 4:
			payloadLen = int(c.ByteOrder.Uint32(header))
		}
		if payloadLen > c.MaxFrameSize {
			return nil, ErrFrameTooLarge
		}
		totalLen := c.HeaderSize + payloadLen
		if buffer.Length() < totalLen {
			break
		}
		buffer.Discard(c.HeaderSize)
		message, _ := buffer.Read(payloadLen)
		messages = append(messages, message)
	}
	return messages, nil
}
