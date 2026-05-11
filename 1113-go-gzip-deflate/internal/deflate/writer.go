package deflate

import (
	"io"
)

const (
	BlockTypeStored  = 0
	BlockTypeFixed   = 1
	BlockTypeDynamic = 2
	DefaultBlockSize = 32768
)

type Writer struct {
	dst         io.Writer
	bitWriter   *BitWriter
	lz77        *LZ77Encoder
	buffer      []byte
	closed      bool
	totalInput  int64
	adler32     uint32
}

func NewWriter(dst io.Writer) *Writer {
	return &Writer{
		dst:       dst,
		bitWriter: NewBitWriter(dst),
		lz77:      NewLZ77Encoder(),
		adler32:   1,
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}

	w.buffer = append(w.buffer, p...)
	w.updateAdler32(p)
	w.totalInput += int64(len(p))

	for len(w.buffer) >= DefaultBlockSize {
		if err := w.writeBlock(false); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

func (w *Writer) Flush() error {
	if w.closed {
		return io.ErrClosedPipe
	}

	for len(w.buffer) > 0 {
		if err := w.writeBlock(len(w.buffer) < 100); err != nil {
			return err
		}
	}

	return w.bitWriter.Flush()
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}

	for len(w.buffer) > 0 {
		if err := w.writeBlock(true); err != nil {
			return err
		}
	}

	if err := w.bitWriter.Flush(); err != nil {
		return err
	}

	w.closed = true
	return nil
}

func (w *Writer) updateAdler32(data []byte) {
	const mod = 65521
	s1 := uint32(w.adler32 & 0xffff)
	s2 := uint32(w.adler32 >> 16)

	for _, b := range data {
		s1 = (s1 + uint32(b)) % mod
		s2 = (s2 + s1) % mod
	}

	w.adler32 = (s2 << 16) | s1
}

func (w *Writer) writeBlock(useStored bool) error {
	var blockData []byte
	if len(w.buffer) > DefaultBlockSize {
		blockData = w.buffer[:DefaultBlockSize]
		w.buffer = w.buffer[DefaultBlockSize:]
	} else {
		blockData = w.buffer
		w.buffer = nil
	}

	final := len(w.buffer) == 0

	if useStored || len(blockData) < 100 {
		return w.writeStoredBlock(blockData, final)
	}

	return w.writeCompressedBlock(blockData, final)
}

func (w *Writer) writeStoredBlock(data []byte, final bool) error {
	if final {
		w.bitWriter.WriteBits(1, 1)
	} else {
		w.bitWriter.WriteBits(0, 1)
	}
	w.bitWriter.WriteBits(BlockTypeStored, 2)

	for w.bitWriter.BitCount() > 0 {
		w.bitWriter.WriteBits(0, 1)
	}

	length := len(data)
	w.dst.Write([]byte{
		byte(length & 0xff),
		byte((length >> 8) & 0xff),
		byte((^length) & 0xff),
		byte((^length >> 8) & 0xff),
	})
	w.dst.Write(data)

	return nil
}

func (w *Writer) writeCompressedBlock(data []byte, final bool) error {
	tokens := w.lz77.Encode(data)

	if final {
		w.bitWriter.WriteBits(1, 1)
	} else {
		w.bitWriter.WriteBits(0, 1)
	}
	w.bitWriter.WriteBits(BlockTypeFixed, 2)

	litLenCodes := GetFixedLiteralLengthCodes()
	distCodes := GetFixedDistanceCodes()

	for _, tok := range tokens {
		if tok.IsMatch {
			lCode, lExtraBits, lExtra := LengthCode(tok.Length)
			w.writeCode(litLenCodes[lCode])
			if lExtraBits > 0 {
				w.bitWriter.WriteBits(uint(lExtra), lExtraBits)
			}

			dCode, dExtraBits, dExtra := DistanceCode(tok.Distance)
			w.writeCode(distCodes[dCode])
			if dExtraBits > 0 {
				w.bitWriter.WriteBits(uint(dExtra), dExtraBits)
			}
		} else {
			w.writeCode(litLenCodes[tok.Literal])
		}
	}

	w.writeCode(litLenCodes[256])

	return nil
}

func (w *Writer) writeCode(code HuffmanCode) {
	if code.Bits > 0 {
		w.bitWriter.WriteBits(code.Code, code.Bits)
	}
}

func (w *Writer) TotalInput() int64 {
	return w.totalInput
}

func (w *Writer) Adler32() uint32 {
	return w.adler32
}
