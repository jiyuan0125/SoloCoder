package deflate

import (
	"bytes"
	"io"
)

type Writer struct {
	out      io.Writer
	bw       *BitWriter
	inputBuf []byte
}

func NewWriter(out io.Writer) *Writer {
	return &Writer{
		out: out,
		bw:  NewBitWriter(out),
	}
}

func (w *Writer) Write(data []byte) (int, error) {
	w.inputBuf = append(w.inputBuf, data...)
	return len(data), nil
}

func (w *Writer) Flush() error {
	for len(w.inputBuf) > 0 {
		blockSize := min(len(w.inputBuf), 65535)
		block := w.inputBuf[:blockSize]
		w.inputBuf = w.inputBuf[blockSize:]

		if err := w.writeBlock(block, len(w.inputBuf) == 0); err != nil {
			return err
		}
	}
	return w.bw.Flush()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (w *Writer) writeBlock(data []byte, final bool) error {
	if final {
		if err := w.bw.WriteBit(true); err != nil {
			return err
		}
	} else {
		if err := w.bw.WriteBit(false); err != nil {
			return err
		}
	}

	if len(data) < 10 {
		return w.writeStoredBlock(data)
	}
	return w.writeFixedBlock(data)
}

func (w *Writer) writeStoredBlock(data []byte) error {
	if err := w.bw.WriteBits(0, 2); err != nil {
		return err
	}
	if err := w.bw.Flush(); err != nil {
		return err
	}

	length := uint16(len(data))
	if _, err := w.out.Write([]byte{byte(length), byte(length >> 8)}); err != nil {
		return err
	}
	lengthComplement := ^length
	if _, err := w.out.Write([]byte{byte(lengthComplement), byte(lengthComplement >> 8)}); err != nil {
		return err
	}

	if len(data) > 0 {
		if _, err := w.out.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeFixedBlock(data []byte) error {
	if err := w.bw.WriteBits(1, 2); err != nil {
		return err
	}

	tokens := LZ77Encode(data)
	for _, token := range tokens {
		var sym int
		var code HuffmanCode
		if token.Type == TokenLiteral {
			sym = int(token.Literal)
			code = GetFixedLiteralCode(sym)
		} else {
			sym, extra, extraBits := GetLengthCode(token.Length)
			code = GetFixedLiteralCode(sym)
			if err := w.writeHuffmanCode(code); err != nil {
				return err
			}
			if extraBits > 0 {
				if err := w.bw.WriteBits(uint32(extra), extraBits); err != nil {
					return err
				}
			}
			distSym, distExtra, distExtraBits := GetDistanceCode(token.Distance)
			distCode := GetFixedDistanceCode(distSym)
			if err := w.writeHuffmanCode(distCode); err != nil {
				return err
			}
			if distExtraBits > 0 {
				if err := w.bw.WriteBits(uint32(distExtra), distExtraBits); err != nil {
					return err
				}
			}
			continue
		}
		if err := w.writeHuffmanCode(code); err != nil {
			return err
		}
	}

	eob := GetFixedLiteralCode(256)
	return w.writeHuffmanCode(eob)
}

func (w *Writer) writeHuffmanCode(code HuffmanCode) error {
	for i := 0; i < code.Length; i++ {
		bit := (code.Code >> uint(code.Length-1-i)) & 1
		if err := w.bw.WriteBit(bit == 1); err != nil {
			return err
		}
	}
	return nil
}

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
