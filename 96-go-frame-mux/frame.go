package mux

import (
	"encoding/binary"
	"io"
)

const (
	FlagSYN byte = 1 << 0
	FlagFIN byte = 1 << 1
	FlagACK byte = 1 << 2
	FlagRST byte = 1 << 3

	FrameHeaderSize = 9
)

type Frame struct {
	StreamID uint32
	Flags    byte
	Length   uint32
	Payload  []byte
}

func ReadFrame(r io.Reader) (*Frame, error) {
	header := make([]byte, FrameHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	streamID := binary.BigEndian.Uint32(header[0:4])
	flags := header[4]
	length := binary.BigEndian.Uint32(header[5:9])

	frame := &Frame{
		StreamID: streamID,
		Flags:    flags,
		Length:   length,
	}

	if length > 0 {
		frame.Payload = make([]byte, length)
		if _, err := io.ReadFull(r, frame.Payload); err != nil {
			return nil, err
		}
	}

	return frame, nil
}

func WriteFrame(w io.Writer, frame *Frame) error {
	header := make([]byte, FrameHeaderSize)
	binary.BigEndian.PutUint32(header[0:4], frame.StreamID)
	header[4] = frame.Flags
	binary.BigEndian.PutUint32(header[5:9], uint32(len(frame.Payload)))

	if _, err := w.Write(header); err != nil {
		return err
	}

	if len(frame.Payload) > 0 {
		if _, err := w.Write(frame.Payload); err != nil {
			return err
		}
	}

	return nil
}
