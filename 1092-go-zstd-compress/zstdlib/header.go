package zstdlib

import (
	"encoding/binary"
	"errors"
	"io"

	"zstd-tool/protocol"
)

var (
	ErrInvalidMagic    = errors.New("invalid magic number, not a zstd compressed file")
	ErrTruncatedHeader = errors.New("truncated file header")
	ErrHeaderMismatch  = errors.New("header size information mismatch")
)

func WriteHeader(w io.Writer, originalSize, compressedSize uint32, level protocol.CompressionLevel) error {
	header := protocol.FileHeader{
		MagicNumber:    protocol.MagicNumber,
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
		Level:          level,
	}

	if err := binary.Write(w, binary.LittleEndian, header.MagicNumber); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, header.OriginalSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, header.CompressedSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, int16(header.Level)); err != nil {
		return err
	}

	return nil
}

func ReadHeader(r io.Reader) (*protocol.FileHeader, error) {
	var header protocol.FileHeader

	if err := binary.Read(r, binary.LittleEndian, &header.MagicNumber); err != nil {
		if err == io.EOF {
			return nil, ErrTruncatedHeader
		}
		return nil, err
	}

	if header.MagicNumber != protocol.MagicNumber {
		return nil, ErrInvalidMagic
	}

	if err := binary.Read(r, binary.LittleEndian, &header.OriginalSize); err != nil {
		return nil, ErrTruncatedHeader
	}
	if err := binary.Read(r, binary.LittleEndian, &header.CompressedSize); err != nil {
		return nil, ErrTruncatedHeader
	}

	var level int16
	if err := binary.Read(r, binary.LittleEndian, &level); err != nil {
		return nil, ErrTruncatedHeader
	}
	header.Level = protocol.CompressionLevel(level)

	return &header, nil
}
