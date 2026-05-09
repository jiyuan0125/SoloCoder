package lzw

import (
	"errors"
	"io"
)

type Decoder struct {
	dict []string
	nextCode uint16
	codeWidth int
	maxCode uint16
	clearCode uint16
	eoiCode uint16
	minCodeSize int
	bitsBuf uint32
	bitsCount int
	reader io.Reader
	prev string
}

func NewDecoder(minCodeSize int, r io.Reader) (*Decoder, error) {
	if minCodeSize < 2 || minCodeSize > 8 {
		return nil, errors.New("minCodeSize must be between 2 and 8")
	}

	clearCode := uint16(1 << minCodeSize)
	eoiCode := clearCode + 1

	dec := &Decoder{
		dict: make([]string, 4096),
		nextCode: eoiCode + 1,
		codeWidth: minCodeSize + 1,
		maxCode: 1 << (minCodeSize + 1),
		clearCode: clearCode,
		eoiCode: eoiCode,
		minCodeSize: minCodeSize,
		reader: r,
	}

	for i := 0; i < int(clearCode); i++ {
		dec.dict[i] = string([]byte{byte(i)})
	}

	return dec, nil
}

func (d *Decoder) reset() {
	clearCode := d.clearCode
	eoiCode := d.eoiCode

	d.dict = make([]string, 4096)
	d.nextCode = eoiCode + 1
	d.codeWidth = d.minCodeSize + 1
	d.maxCode = 1 << (d.minCodeSize + 1)
	d.prev = ""

	for i := 0; i < int(clearCode); i++ {
		d.dict[i] = string([]byte{byte(i)})
	}
}

func (d *Decoder) readCode() (uint16, bool, error) {
	for d.bitsCount < d.codeWidth {
		buf := make([]byte, 1)
		n, err := d.reader.Read(buf)
		if n == 0 {
			if err == io.EOF {
				return 0, false, nil
			}
			return 0, false, err
		}

		d.bitsBuf |= uint32(buf[0]) << d.bitsCount
		d.bitsCount += 8
	}

	mask := uint16((1 << d.codeWidth) - 1)
	code := uint16(d.bitsBuf) & mask
	d.bitsBuf >>= d.codeWidth
	d.bitsCount -= d.codeWidth

	return code, true, nil
}

func (d *Decoder) Decompress(w io.Writer) error {
	first := true
	for {
		code, ok, err := d.readCode()
		if !ok {
			return nil
		}
		if err != nil {
			return err
		}

		if code == d.clearCode {
			d.reset()
			first = true
			continue
		}

		if code == d.eoiCode {
			return nil
		}

		var current string
		if first {
			if code < uint16(len(d.dict)) && d.dict[code] != "" {
				current = d.dict[code]
			} else {
				return errors.New("invalid first code")
			}
			d.prev = current
			first = false
			_, err = w.Write([]byte(current))
			if err != nil {
				return err
			}
			continue
		}

		if code < d.nextCode && d.dict[code] != "" {
			current = d.dict[code]
			_, err = w.Write([]byte(current))
			if err != nil {
				return err
			}

			if d.nextCode < 4096 {
				newEntry := d.prev + string(current[0])
				d.dict[d.nextCode] = newEntry
				if d.nextCode == d.maxCode-1 {
					if d.codeWidth < 12 {
						d.codeWidth++
						d.maxCode = 1 << d.codeWidth
					}
				}
				d.nextCode++
			}
		} else if code == d.nextCode {
			current = d.prev + string(d.prev[0])
			_, err = w.Write([]byte(current))
			if err != nil {
				return err
			}

			if d.nextCode < 4096 {
				d.dict[d.nextCode] = current
				if d.nextCode == d.maxCode-1 {
					if d.codeWidth < 12 {
						d.codeWidth++
						d.maxCode = 1 << d.codeWidth
					}
				}
				d.nextCode++
			}
		} else {
			return errors.New("invalid code")
		}

		d.prev = current
	}
}
