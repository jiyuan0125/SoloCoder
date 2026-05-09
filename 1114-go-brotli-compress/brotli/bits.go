package brotli

import (
	"errors"
	"io"
)

type bitReader struct {
	input    io.Reader
	buffer   []byte
	bitPos   uint
	bytePos  int
	bitsLeft int
}

func newBitReader(r io.Reader) *bitReader {
	return &bitReader{
		input:    r,
		buffer:   make([]byte, 4096),
		bitPos:   0,
		bytePos:  0,
		bitsLeft: 0,
	}
}

func (br *bitReader) ensureBits(n int) error {
	for br.bitsLeft < n {
		if br.bytePos >= len(br.buffer) {
			br.bytePos = 0
		}
		read, err := br.input.Read(br.buffer[br.bytePos:])
		if err != nil {
			if err == io.EOF {
				return errors.New("unexpected EOF")
			}
			return err
		}
		if read == 0 {
			return errors.New("unexpected EOF")
		}
		br.bitsLeft += read * 8
	}
	return nil
}

func (br *bitReader) readBits(n int) (uint64, error) {
	if n == 0 {
		return 0, nil
	}
	if err := br.ensureBits(n); err != nil {
		return 0, err
	}

	var result uint64
	remaining := n

	for remaining > 0 {
		byteIndex := br.bitPos / 8
		bitInByte := br.bitPos % 8
		bitsAvailable := 8 - bitInByte
		bitsToTake := remaining
		if bitsToTake > int(bitsAvailable) {
			bitsToTake = int(bitsAvailable)
		}

		mask := uint8((1 << bitsToTake) - 1)
		value := (br.buffer[byteIndex] >> bitInByte) & mask
		result |= uint64(value) << (n - remaining)

		br.bitPos += uint(bitsToTake)
		br.bitsLeft -= bitsToTake
		remaining -= bitsToTake

		if bitInByte+uint8(bitsToTake) == 8 {
			br.bytePos++
		}
	}

	return result, nil
}

func (br *bitReader) alignToByte() {
	offset := br.bitPos % 8
	if offset > 0 {
		br.bitPos += 8 - offset
		br.bitsLeft -= int(8 - offset)
		if (br.bitPos % 8) == 0 {
			br.bytePos = int(br.bitPos / 8)
		}
	}
}

type bitWriter struct {
	output   io.Writer
	buffer   []byte
	bitPos   uint
	bytePos  int
}

func newBitWriter(w io.Writer) *bitWriter {
	return &bitWriter{
		output:   w,
		buffer:   make([]byte, 4096),
		bitPos:   0,
		bytePos:  0,
	}
}

func (bw *bitWriter) writeBits(value uint64, n int) error {
	if n == 0 {
		return nil
	}

	remaining := n

	for remaining > 0 {
		byteIndex := bw.bitPos / 8
		bitInByte := bw.bitPos % 8
		bitsAvailable := 8 - bitInByte
		bitsToWrite := remaining
		if bitsToWrite > int(bitsAvailable) {
			bitsToWrite = int(bitsAvailable)
		}

		mask := uint8((1 << bitsToWrite) - 1)
		shift := n - remaining
		byteValue := uint8((value >> shift) & uint64(mask))
		bw.buffer[byteIndex] |= byteValue << bitInByte

		bw.bitPos += uint(bitsToWrite)
		remaining -= bitsToWrite

		if bitInByte+uint8(bitsToWrite) == 8 {
			bw.bytePos = int(byteIndex) + 1
		}

		if bw.bytePos >= len(bw.buffer)-1 {
			if err := bw.flush(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (bw *bitWriter) alignToByte() error {
	offset := bw.bitPos % 8
	if offset > 0 {
		if err := bw.writeBits(0, int(8-offset)); err != nil {
			return err
		}
	}
	return nil
}

func (bw *bitWriter) flush() error {
	if bw.bytePos > 0 {
		_, err := bw.output.Write(bw.buffer[:bw.bytePos])
		if err != nil {
			return err
		}
	}
	bw.bitPos = 0
	bw.bytePos = 0
	for i := range bw.buffer {
		bw.buffer[i] = 0
	}
	return nil
}

func (bw *bitWriter) close() error {
	if err := bw.alignToByte(); err != nil {
		return err
	}
	return bw.flush()
}
