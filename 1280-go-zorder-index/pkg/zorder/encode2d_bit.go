package zorder

func Encode2DBit(p Point2D) Code {
	return Code(interleave2DBit(p.X) | (interleave2DBit(p.Y) << 1))
}

func Decode2DBit(c Code) Point2D {
	return Point2D{
		X: deinterleave2DBit(uint64(c)),
		Y: deinterleave2DBit(uint64(c) >> 1),
	}
}

func interleave2DBit(x uint32) uint64 {
	x64 := uint64(x)
	x64 = (x64 | (x64 << 16)) & 0x0000FFFF0000FFFF
	x64 = (x64 | (x64 << 8)) & 0x00FF00FF00FF00FF
	x64 = (x64 | (x64 << 4)) & 0x0F0F0F0F0F0F0F0F
	x64 = (x64 | (x64 << 2)) & 0x3333333333333333
	x64 = (x64 | (x64 << 1)) & 0x5555555555555555
	return x64
}

func deinterleave2DBit(x uint64) uint32 {
	x &= 0x5555555555555555
	x = (x | (x >> 1)) & 0x3333333333333333
	x = (x | (x >> 2)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF0000FFFF
	x = (x | (x >> 16)) & 0x00000000FFFFFFFF
	return uint32(x)
}
