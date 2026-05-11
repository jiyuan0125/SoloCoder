package brotli

type bitWriter struct {
	buf      []byte
	offset   uint64
	bitCount uint64
}

func newBitWriter() *bitWriter {
	return &bitWriter{
		buf:      make([]byte, 0, 256),
		offset:   0,
		bitCount: 0,
	}
}

func (bw *bitWriter) writeBits(value uint64, numBits int) {
	for numBits > 0 {
		byteIndex := bw.bitCount / 8
		bitIndex := bw.bitCount % 8

		if uint64(len(bw.buf)) <= byteIndex {
			bw.buf = append(bw.buf, 0)
		}

		bitsToWrite := 8 - int(bitIndex)
		if bitsToWrite > numBits {
			bitsToWrite = numBits
		}

		mask := uint64((1 << bitsToWrite) - 1)
		bw.buf[byteIndex] |= uint8((value & mask) << bitIndex)

		value >>= bitsToWrite
		bw.bitCount += uint64(bitsToWrite)
		numBits -= bitsToWrite
	}
}

func (bw *bitWriter) bytes() []byte {
	result := make([]byte, len(bw.buf))
	copy(result, bw.buf)
	return result
}

type bitReader struct {
	buf      []byte
	bitPos   uint64
	totalBits uint64
}

func newBitReader(data []byte) *bitReader {
	return &bitReader{
		buf:       data,
		bitPos:    0,
		totalBits: uint64(len(data)) * 8,
	}
}

func (br *bitReader) bitsLeft() uint64 {
	if br.totalBits <= br.bitPos {
		return 0
	}
	return br.totalBits - br.bitPos
}

func (br *bitReader) readBits(numBits int) (uint64, error) {
	if numBits == 0 {
		return 0, nil
	}

	if br.bitPos+uint64(numBits) > br.totalBits {
		return 0, errCorrupted
	}

	var result uint64
	for i := 0; i < numBits; i++ {
		if br.readBit() {
			result |= 1 << i
		}
	}
	return result, nil
}

func (br *bitReader) readBit() bool {
	byteIndex := br.bitPos / 8
	bitIndex := br.bitPos % 8
	br.bitPos++
	return (br.buf[byteIndex] & (1 << bitIndex)) != 0
}
