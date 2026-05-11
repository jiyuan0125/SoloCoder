package lzw

import (
	"errors"
	"io"
)

type Encoder struct {
	dict map[string]uint16
	nextCode uint16
	codeWidth int
	maxCode uint16
	clearCode uint16
	eoiCode uint16
	minCodeSize int
	bitsBuf uint32
	bitsCount int
	writer io.Writer
}

func NewEncoder(minCodeSize int, w io.Writer) (*Encoder, error) {
	if minCodeSize < 2 || minCodeSize > 8 {
		return nil, errors.New("minCodeSize must be between 2 and 8")
	}

	clearCode := uint16(1 << minCodeSize)
	eoiCode := clearCode + 1

	enc := &Encoder{
		dict: make(map[string]uint16, 4096),
		nextCode: eoiCode + 1,
		codeWidth: minCodeSize + 1,
		maxCode: 1 << (minCodeSize + 1),
		clearCode: clearCode,
		eoiCode: eoiCode,
		minCodeSize: minCodeSize,
		writer: w,
	}

	for i := 0; i < 256; i++ {
		enc.dict[string([]byte{byte(i)})] = uint16(i)
	}

	err := enc.writeCode(clearCode)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

func (e *Encoder) reset() error {
	eoiCode := e.eoiCode

	e.dict = make(map[string]uint16, 4096)
	e.nextCode = eoiCode + 1
	e.codeWidth = e.minCodeSize + 1
	e.maxCode = 1 << (e.minCodeSize + 1)

	for i := 0; i < 256; i++ {
		e.dict[string([]byte{byte(i)})] = uint16(i)
	}

	return e.writeCode(e.clearCode)
}

func (e *Encoder) writeCode(code uint16) error {
	e.bitsBuf |= uint32(code) << e.bitsCount
	e.bitsCount += e.codeWidth

	for e.bitsCount >= 8 {
		b := byte(e.bitsBuf & 0xff)
		_, err := e.writer.Write([]byte{b})
		if err != nil {
			return err
		}
		e.bitsBuf >>= 8
		e.bitsCount -= 8
	}

	return nil
}

func (e *Encoder) flush() error {
	for e.bitsCount > 0 {
		b := byte(e.bitsBuf & 0xff)
		_, err := e.writer.Write([]byte{b})
		if err != nil {
			return err
		}
		e.bitsBuf >>= 8
		e.bitsCount -= 8
	}
	return nil
}

func (e *Encoder) Compress(r io.Reader) error {
	var current string
	buf := make([]byte, 1)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			c := buf[0]
			extended := current + string([]byte{c})

			if _, ok := e.dict[extended]; ok {
				current = extended
			} else {
				code := e.dict[current]
				err = e.writeCode(code)
				if err != nil {
					return err
				}

				if e.nextCode < 4096 {
					e.dict[extended] = e.nextCode
					if e.nextCode == e.maxCode {
						if e.codeWidth < 12 {
							e.codeWidth++
							e.maxCode = 1 << e.codeWidth
						}
					}
					e.nextCode++
				} else {
					err = e.reset()
					if err != nil {
						return err
					}
				}

				current = string([]byte{c})
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	if current != "" {
		code := e.dict[current]
		err := e.writeCode(code)
		if err != nil {
			return err
		}
	}

	err := e.writeCode(e.eoiCode)
	if err != nil {
		return err
	}

	return e.flush()
}
