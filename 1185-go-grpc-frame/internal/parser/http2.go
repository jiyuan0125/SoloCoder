package parser

import (
	"encoding/binary"
	"errors"
)

const (
	HTTP2FrameHeaderSize = 9
	HTTP2MaxFrameSize    = 16384
	HTTP2StreamReserved  = 0x7FFFFFFF

	FrameTypeData    = 0x0
	FrameTypeHeaders = 0x1
	FrameTypePriority = 0x2
	FrameTypeRstStream = 0x3
	FrameTypeSettings = 0x4
	FrameTypePushPromise = 0x5
	FrameTypePing = 0x6
	FrameTypeGoAway = 0x7
	FrameTypeWindowUpdate = 0x8
	FrameTypeContinuation = 0x9

	FlagDataEndStream    = 0x1
	FlagDataPadded       = 0x8

	FlagHeadersEndStream  = 0x1
	FlagHeadersEndHeaders = 0x4
	FlagHeadersPadded     = 0x8
	FlagHeadersPriority   = 0x20

	FlagSettingsAck = 0x1

	FlagPingAck = 0x1
)

func FrameTypeName(frameType uint8) string {
	switch frameType {
	case FrameTypeData:
		return "DATA"
	case FrameTypeHeaders:
		return "HEADERS"
	case FrameTypePriority:
		return "PRIORITY"
	case FrameTypeRstStream:
		return "RST_STREAM"
	case FrameTypeSettings:
		return "SETTINGS"
	case FrameTypePushPromise:
		return "PUSH_PROMISE"
	case FrameTypePing:
		return "PING"
	case FrameTypeGoAway:
		return "GOAWAY"
	case FrameTypeWindowUpdate:
		return "WINDOW_UPDATE"
	case FrameTypeContinuation:
		return "CONTINUATION"
	default:
		return "UNKNOWN"
	}
}

func FrameTypeFlags(frameType uint8, flags uint8) []string {
	var flagNames []string

	switch frameType {
	case FrameTypeData:
		if flags&FlagDataEndStream != 0 {
			flagNames = append(flagNames, "END_STREAM")
		}
		if flags&FlagDataPadded != 0 {
			flagNames = append(flagNames, "PADDED")
		}
	case FrameTypeHeaders:
		if flags&FlagHeadersEndStream != 0 {
			flagNames = append(flagNames, "END_STREAM")
		}
		if flags&FlagHeadersEndHeaders != 0 {
			flagNames = append(flagNames, "END_HEADERS")
		}
		if flags&FlagHeadersPadded != 0 {
			flagNames = append(flagNames, "PADDED")
		}
		if flags&FlagHeadersPriority != 0 {
			flagNames = append(flagNames, "PRIORITY")
		}
	case FrameTypeSettings:
		if flags&FlagSettingsAck != 0 {
			flagNames = append(flagNames, "ACK")
		}
	case FrameTypePing:
		if flags&FlagPingAck != 0 {
			flagNames = append(flagNames, "ACK")
		}
	}

	return flagNames
}

type HTTP2Frame struct {
	Length     uint32
	Type       uint8
	Flags      uint8
	StreamID   uint32
	Payload    []byte
	TotalSize  int
}

type HTTP2Parser struct {
	buffer []byte
}

func NewHTTP2Parser() *HTTP2Parser {
	return &HTTP2Parser{
		buffer: make([]byte, 0),
	}
}

func (h *HTTP2Parser) Write(data []byte) {
	if len(data) == 0 {
		return
	}
	h.buffer = append(h.buffer, data...)
}

func (h *HTTP2Parser) HasFrame() bool {
	if len(h.buffer) < HTTP2FrameHeaderSize {
		return false
	}

	length := h.readFrameLength()
	return len(h.buffer) >= HTTP2FrameHeaderSize+int(length)
}

func (h *HTTP2Parser) readFrameLength() uint32 {
	return uint32(h.buffer[0])<<16 |
		uint32(h.buffer[1])<<8 |
		uint32(h.buffer[2])
}

func (h *HTTP2Parser) ReadFrame() (*HTTP2Frame, error) {
	if len(h.buffer) < HTTP2FrameHeaderSize {
		return nil, errors.New("insufficient data for frame header")
	}

	length := h.readFrameLength()
	if length > HTTP2MaxFrameSize {
		return nil, errors.New("frame size exceeds maximum allowed")
	}

	totalSize := HTTP2FrameHeaderSize + int(length)
	if len(h.buffer) < totalSize {
		return nil, errors.New("insufficient data for complete frame")
	}

	frame := &HTTP2Frame{
		Length:    length,
		Type:      h.buffer[3],
		Flags:     h.buffer[4],
		StreamID:  binary.BigEndian.Uint32(h.buffer[5:9]) & HTTP2StreamReserved,
		Payload:   h.buffer[HTTP2FrameHeaderSize:totalSize],
		TotalSize: totalSize,
	}

	h.buffer = h.buffer[totalSize:]

	return frame, nil
}

func (h *HTTP2Parser) Buffered() int {
	return len(h.buffer)
}

func ParseHTTP2Frames(data []byte) ([]HTTP2Frame, error) {
	parser := NewHTTP2Parser()
	parser.Write(data)

	var frames []HTTP2Frame
	for parser.HasFrame() {
		frame, err := parser.ReadFrame()
		if err != nil {
			return frames, err
		}
		frames = append(frames, *frame)
	}

	return frames, nil
}

func EncodeHTTP2Frame(frame HTTP2Frame) ([]byte, error) {
	if frame.Length > HTTP2MaxFrameSize {
		return nil, errors.New("frame payload exceeds maximum size")
	}

	if len(frame.Payload) != int(frame.Length) {
		return nil, errors.New("payload length mismatch")
	}

	buf := make([]byte, HTTP2FrameHeaderSize+int(frame.Length))

	buf[0] = byte((frame.Length >> 16) & 0xFF)
	buf[1] = byte((frame.Length >> 8) & 0xFF)
	buf[2] = byte(frame.Length & 0xFF)
	buf[3] = frame.Type
	buf[4] = frame.Flags

	binary.BigEndian.PutUint32(buf[5:9], frame.StreamID)

	copy(buf[HTTP2FrameHeaderSize:], frame.Payload)

	return buf, nil
}

func EncodeHTTP2Frames(frames []HTTP2Frame) ([]byte, error) {
	var result []byte
	for _, frame := range frames {
		encoded, err := EncodeHTTP2Frame(frame)
		if err != nil {
			return nil, err
		}
		result = append(result, encoded...)
	}
	return result, nil
}
