package zorder

var encode2DByteToWord [256]uint16
var decode2DWordToByteX [65536]uint8
var decode2DWordToByteY [65536]uint8

func init() {
	for x := 0; x < 256; x++ {
		encode2DByteToWord[x] = spreadByte(uint8(x))
	}

	for x := 0; x < 256; x++ {
		for y := 0; y < 256; y++ {
			word := encode2DByteToWord[x] | (encode2DByteToWord[y] << 1)
			decode2DWordToByteX[word] = uint8(x)
			decode2DWordToByteY[word] = uint8(y)
		}
	}
}

func spreadByte(b uint8) uint16 {
	x := uint16(b)
	x = (x | (x << 4)) & 0x0F0F
	x = (x | (x << 2)) & 0x3333
	x = (x | (x << 1)) & 0x5555
	return x
}

func Encode2DLookup(p Point2D) Code {
	x := p.X
	y := p.Y

	b0X := uint8(x)
	b1X := uint8(x >> 8)
	b2X := uint8(x >> 16)
	b3X := uint8(x >> 24)

	b0Y := uint8(y)
	b1Y := uint8(y >> 8)
	b2Y := uint8(y >> 16)
	b3Y := uint8(y >> 24)

	word0 := uint64(encode2DByteToWord[b0X]) | (uint64(encode2DByteToWord[b0Y]) << 1)
	word1 := uint64(encode2DByteToWord[b1X]) | (uint64(encode2DByteToWord[b1Y]) << 1)
	word2 := uint64(encode2DByteToWord[b2X]) | (uint64(encode2DByteToWord[b2Y]) << 1)
	word3 := uint64(encode2DByteToWord[b3X]) | (uint64(encode2DByteToWord[b3Y]) << 1)

	return Code(word0 | (word1 << 16) | (word2 << 32) | (word3 << 48))
}

func Decode2DLookup(c Code) Point2D {
	code := uint64(c)

	word0 := uint16(code & 0xFFFF)
	word1 := uint16((code >> 16) & 0xFFFF)
	word2 := uint16((code >> 32) & 0xFFFF)
	word3 := uint16((code >> 48) & 0xFFFF)

	x := uint32(decode2DWordToByteX[word0]) |
		(uint32(decode2DWordToByteX[word1]) << 8) |
		(uint32(decode2DWordToByteX[word2]) << 16) |
		(uint32(decode2DWordToByteX[word3]) << 24)

	y := uint32(decode2DWordToByteY[word0]) |
		(uint32(decode2DWordToByteY[word1]) << 8) |
		(uint32(decode2DWordToByteY[word2]) << 16) |
		(uint32(decode2DWordToByteY[word3]) << 24)

	return Point2D{X: x, Y: y}
}
