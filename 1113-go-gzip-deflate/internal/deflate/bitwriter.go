package deflate

import "io"

type BitWriter struct {
	dst      io.Writer
	buffer   uint32
	bitCount int
}

func NewBitWriter(dst io.Writer) *BitWriter {
	return &BitWriter{dst: dst}
}

func (w *BitWriter) WriteBits(value uint, n int) {
	for n > 0 {
		remaining := 32 - w.bitCount
		if n > remaining {
			w.buffer |= uint32((value & ((1 << remaining) - 1)) << w.bitCount)
			w.bitCount += remaining
			value >>= remaining
			n -= remaining

			w.flushBytes()
		} else {
			w.buffer |= uint32((value & ((1 << n) - 1)) << w.bitCount)
			w.bitCount += n
			n = 0

			if w.bitCount >= 8 {
				w.flushBytes()
			}
		}
	}
}

func (w *BitWriter) WriteByte(b byte) {
	w.WriteBits(uint(b), 8)
}

func (w *BitWriter) WriteBytes(data []byte) {
	for _, b := range data {
		w.WriteByte(b)
	}
}

func (w *BitWriter) flushBytes() {
	for w.bitCount >= 8 {
		b := byte(w.buffer & 0xff)
		w.dst.Write([]byte{b})
		w.buffer >>= 8
		w.bitCount -= 8
	}
}

func (w *BitWriter) Flush() error {
	w.flushBytes()
	if w.bitCount > 0 {
		b := byte(w.buffer & 0xff)
		_, err := w.dst.Write([]byte{b})
		w.buffer = 0
		w.bitCount = 0
		return err
	}
	return nil
}

func (w *BitWriter) BitCount() int {
	return w.bitCount
}
