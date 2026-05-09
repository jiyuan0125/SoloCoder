package deflate

import "io"

type BitWriter struct {
	out   io.Writer
	buf   byte
	count int
}

func NewBitWriter(out io.Writer) *BitWriter {
	return &BitWriter{out: out}
}

func (bw *BitWriter) WriteBit(b bool) error {
	if b {
		bw.buf |= 1 << uint(bw.count)
	}
	bw.count++
	if bw.count == 8 {
		if _, err := bw.out.Write([]byte{bw.buf}); err != nil {
			return err
		}
		bw.buf = 0
		bw.count = 0
	}
	return nil
}

func (bw *BitWriter) WriteBits(val uint32, n int) error {
	for i := 0; i < n; i++ {
		if err := bw.WriteBit((val>>uint(i))&1 == 1); err != nil {
			return err
		}
	}
	return nil
}

func (bw *BitWriter) Flush() error {
	if bw.count > 0 {
		if _, err := bw.out.Write([]byte{bw.buf}); err != nil {
			return err
		}
		bw.buf = 0
		bw.count = 0
	}
	return nil
}

func (bw *BitWriter) Bytes() int {
	return 0
}

type BitReader struct {
	in    io.Reader
	buf   byte
	count int
}

func NewBitReader(in io.Reader) *BitReader {
	return &BitReader{in: in}
}

func (br *BitReader) ReadBit() (bool, error) {
	if br.count == 0 {
		var b [1]byte
		n, err := br.in.Read(b[:])
		if n == 0 {
			return false, io.EOF
		}
		if err != nil {
			return false, err
		}
		br.buf = b[0]
		br.count = 8
	}
	br.count--
	bit := br.buf&1 == 1
	br.buf >>= 1
	return bit, nil
}

func (br *BitReader) ReadBits(n int) (uint32, error) {
	var val uint32
	for i := 0; i < n; i++ {
		bit, err := br.ReadBit()
		if err != nil {
			return 0, err
		}
		if bit {
			val |= 1 << uint(i)
		}
	}
	return val, nil
}
