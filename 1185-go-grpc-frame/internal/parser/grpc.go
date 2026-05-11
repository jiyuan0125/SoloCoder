package parser

import (
	"encoding/binary"
	"errors"
)

const (
	GRPCFrameHeaderSize = 5
	GRPCCompressionNone = 0
	GRPCCompressionGZIP = 1
)

var ErrGRPCInsufficientData = errors.New("insufficient data")

type GRPCFrame struct {
	Compressed bool
	Length     uint32
	Message    []byte
}

type GRPCStreamBuffer struct {
	buffer []byte
	frames []GRPCFrame
}

func NewGRPCStreamBuffer() *GRPCStreamBuffer {
	return &GRPCStreamBuffer{
		buffer: make([]byte, 0),
		frames: make([]GRPCFrame, 0),
	}
}

func (g *GRPCStreamBuffer) Write(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	g.buffer = append(g.buffer, data...)

	for {
		frame, consumed, err := g.tryParseFrame()
		if err != nil {
			if errors.Is(err, ErrGRPCInsufficientData) {
				break
			}
			return err
		}
		if consumed == 0 {
			break
		}

		g.frames = append(g.frames, frame)
	}

	return nil
}

func (g *GRPCStreamBuffer) tryParseFrame() (GRPCFrame, int, error) {
	if len(g.buffer) < GRPCFrameHeaderSize {
		return GRPCFrame{}, 0, ErrGRPCInsufficientData
	}

	compressionFlag := g.buffer[0]
	if compressionFlag != GRPCCompressionNone && compressionFlag != GRPCCompressionGZIP {
		return GRPCFrame{}, 0, errors.New("invalid compression flag")
	}

	length := binary.BigEndian.Uint32(g.buffer[1:5])

	totalNeeded := int(GRPCFrameHeaderSize + length)
	if len(g.buffer) < totalNeeded {
		return GRPCFrame{}, 0, ErrGRPCInsufficientData
	}

	frame := GRPCFrame{
		Compressed: compressionFlag == GRPCCompressionGZIP,
		Length:     length,
		Message:    g.buffer[GRPCFrameHeaderSize:totalNeeded],
	}

	remaining := g.buffer[totalNeeded:]
	g.buffer = remaining

	return frame, totalNeeded, nil
}

func (g *GRPCStreamBuffer) Frames() []GRPCFrame {
	frames := g.frames
	g.frames = make([]GRPCFrame, 0)
	return frames
}

func (g *GRPCStreamBuffer) Buffered() int {
	return len(g.buffer)
}

func (g *GRPCStreamBuffer) HasBufferedData() bool {
	return len(g.buffer) > 0
}

func ParseGRPCFrames(data []byte) ([]GRPCFrame, error) {
	buffer := NewGRPCStreamBuffer()
	if err := buffer.Write(data); err != nil {
		return nil, err
	}
	return buffer.Frames(), nil
}

func EncodeGRPCFrame(frame GRPCFrame) ([]byte, error) {
	if frame.Length != uint32(len(frame.Message)) {
		return nil, errors.New("message length mismatch")
	}

	buf := make([]byte, GRPCFrameHeaderSize+len(frame.Message))

	if frame.Compressed {
		buf[0] = GRPCCompressionGZIP
	} else {
		buf[0] = GRPCCompressionNone
	}

	binary.BigEndian.PutUint32(buf[1:5], frame.Length)
	copy(buf[GRPCFrameHeaderSize:], frame.Message)

	return buf, nil
}

func EncodeGRPCFrames(frames []GRPCFrame) ([]byte, error) {
	var result []byte
	for _, frame := range frames {
		encoded, err := EncodeGRPCFrame(frame)
		if err != nil {
			return nil, err
		}
		result = append(result, encoded...)
	}
	return result, nil
}
