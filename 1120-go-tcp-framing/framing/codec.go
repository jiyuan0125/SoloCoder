package framing

import (
	"errors"
)

type FrameMode string

const (
	ModeLengthPrefix FrameMode = "length_prefix"
	ModeDelimiter    FrameMode = "delimiter"
)

type ByteOrder string

const (
	ByteOrderBigEndian    ByteOrder = "big_endian"
	ByteOrderLittleEndian ByteOrder = "little_endian"
)

type Config struct {
	Mode           FrameMode
	HeaderSize     int
	ByteOrder      ByteOrder
	MaxFrameLength int
	Delimiter      []byte
}

var ErrFrameTooLarge = errors.New("frame too large")

type Codec interface {
	Encode(msg []byte) ([]byte, error)
	Decode(data []byte) ([][]byte, error)
	Reset()
	Config() Config
	UpdateConfig(cfg Config) error
}
