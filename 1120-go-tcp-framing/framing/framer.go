package framing

import (
	"encoding/binary"
	"errors"

	"tcp-framing/protocol"
)

type FramerMode int

const (
	FramerModeLengthPrefix FramerMode = iota
	FramerModeDelimiter
)

var ErrUnknownMode = errors.New("framing: unknown mode")

type Framer struct {
	mode           FramerMode
	lengthPrefix   *LengthPrefixConfig
	delimiter      *DelimiterConfig
	buffer         *RingBuffer
}

type FramerOptions struct {
	Mode         protocol.FramerMode
	HeaderSize   int
	ByteOrder    protocol.ByteOrder
	MaxFrameSize int
	Delimiter    string
}

func DefaultFramerOptions() *FramerOptions {
	return &FramerOptions{
		Mode:         protocol.ModeLengthPrefix,
		HeaderSize:   4,
		ByteOrder:    protocol.ByteOrderBigEndian,
		MaxFrameSize: 16 * 1024 * 1024,
		Delimiter:    "\r\n",
	}
}

func NewFramer(opts *FramerOptions) (*Framer, error) {
	if opts == nil {
		opts = DefaultFramerOptions()
	}
	f := &Framer{
		buffer: NewRingBuffer(4096),
	}
	if err := f.Configure(opts); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Framer) Configure(opts *FramerOptions) error {
	var bo binary.ByteOrder
	switch opts.ByteOrder {
	case protocol.ByteOrderLittleEndian:
		bo = binary.LittleEndian
	default:
		bo = binary.BigEndian
	}
	switch opts.Mode {
	case protocol.ModeLengthPrefix:
		lp, err := NewLengthPrefixConfig(opts.HeaderSize, bo, opts.MaxFrameSize)
		if err != nil {
			return err
		}
		f.mode = FramerModeLengthPrefix
		f.lengthPrefix = lp
		f.delimiter = nil
	case protocol.ModeDelimiter:
		del, err := NewDelimiterConfig([]byte(opts.Delimiter), opts.MaxFrameSize)
		if err != nil {
			return err
		}
		f.mode = FramerModeDelimiter
		f.lengthPrefix = nil
		f.delimiter = del
	default:
		return ErrUnknownMode
	}
	return nil
}

func (f *Framer) Mode() FramerMode {
	return f.mode
}

func (f *Framer) Encode(message []byte) ([]byte, error) {
	switch f.mode {
	case FramerModeLengthPrefix:
		return f.lengthPrefix.Encode(message)
	case FramerModeDelimiter:
		return f.delimiter.Encode(message)
	default:
		return nil, ErrUnknownMode
	}
}

func (f *Framer) EncodeBatch(messages [][]byte) ([]byte, error) {
	var total []byte
	for _, msg := range messages {
		frame, err := f.Encode(msg)
		if err != nil {
			return nil, err
		}
		total = append(total, frame...)
	}
	return total, nil
}

func (f *Framer) Feed(data []byte) {
	f.buffer.Write(data)
}

func (f *Framer) Decode() ([][]byte, error) {
	switch f.mode {
	case FramerModeLengthPrefix:
		return f.lengthPrefix.Decode(f.buffer)
	case FramerModeDelimiter:
		return f.delimiter.Decode(f.buffer)
	default:
		return nil, ErrUnknownMode
	}
}

func (f *Framer) Reset() {
	f.buffer.Reset()
}
