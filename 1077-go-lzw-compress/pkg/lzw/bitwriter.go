package lzw

import (
	"io"
)

type BitWriter struct {
	writer io.Writer
	buffer byte
	offset int
}

func NewBitWriter(w io.Writer) *BitWriter {
	return &BitWriter{
		writer: w,
		buffer: 0,
		offset: 0,
	}
}

func (bw *BitWriter) WriteCode(code int, bitWidth int) error {
	for bitWidth > 0 {
		remainingBitsInByte := 8 - bw.offset
		if bitWidth <= remainingBitsInByte {
			bw.buffer |= byte(code&((1<<bitWidth)-1)) << bw.offset
			bw.offset += bitWidth
			if bw.offset == 8 {
				if err := bw.flush(); err != nil {
					return err
				}
			}
			return nil
		}
		mask := 1 << remainingBitsInByte
		bw.buffer |= byte(code&(mask-1)) << bw.offset
		bw.offset = 8
		if err := bw.flush(); err != nil {
			return err
		}
		code >>= remainingBitsInByte
		bitWidth -= remainingBitsInByte
	}
	return nil
}

func (bw *BitWriter) flush() error {
	if bw.offset == 0 {
		return nil
	}
	_, err := bw.writer.Write([]byte{bw.buffer})
	if err != nil {
		return err
	}
	bw.buffer = 0
	bw.offset = 0
	return nil
}

func (bw *BitWriter) Flush() error {
	if bw.offset > 0 {
		if err := bw.flush(); err != nil {
			return err
		}
	}
	return nil
}
