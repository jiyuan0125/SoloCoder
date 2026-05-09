package lzw

import (
	"io"
)

type BitReader struct {
	reader io.Reader
	buffer byte
	offset int
	eof    bool
}

func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{
		reader: r,
		buffer: 0,
		offset: 8,
		eof:    false,
	}
}

func (br *BitReader) ReadCode(bitWidth int) (int, bool, error) {
	var code int
	var bitsRead int
	for bitsRead < bitWidth {
		if br.offset >= 8 {
			buf := make([]byte, 1)
			n, err := br.reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					if bitsRead == 0 {
						return 0, true, nil
					}
					return code, true, nil
				}
				return 0, false, err
			}
			if n == 0 {
				if bitsRead == 0 {
					return 0, true, nil
				}
				return code, true, nil
			}
			br.buffer = buf[0]
			br.offset = 0
		}
		remainingBitsInByte := 8 - br.offset
		bitsNeeded := bitWidth - bitsRead
		if bitsNeeded <= remainingBitsInByte {
			mask := (1 << bitsNeeded) - 1
			code |= (int(br.buffer>>br.offset) & mask) << bitsRead
			br.offset += bitsNeeded
			bitsRead = bitWidth
		} else {
			mask := (1 << remainingBitsInByte) - 1
			code |= (int(br.buffer>>br.offset) & mask) << bitsRead
			bitsRead += remainingBitsInByte
			br.offset = 8
		}
	}
	return code, false, nil
}
