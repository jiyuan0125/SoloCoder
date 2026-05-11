package zorder

const max3DCoord = (1 << 21) - 1

func Encode3D(p Point3D) (Code, error) {
	if p.X > max3DCoord {
		return 0, &ErrCoordOverflow{Coord: p.X, Max: max3DCoord}
	}
	if p.Y > max3DCoord {
		return 0, &ErrCoordOverflow{Coord: p.Y, Max: max3DCoord}
	}
	if p.Z > max3DCoord {
		return 0, &ErrCoordOverflow{Coord: p.Z, Max: max3DCoord}
	}

	return Code(interleave3D(p.X) | (interleave3D(p.Y) << 1) | (interleave3D(p.Z) << 2)), nil
}

func Decode3D(c Code) Point3D {
	return Point3D{
		X: deinterleave3D(uint64(c)),
		Y: deinterleave3D(uint64(c) >> 1),
		Z: deinterleave3D(uint64(c) >> 2),
	}
}

func interleave3D(x uint32) uint64 {
	x64 := uint64(x) & 0x1FFFFF
	x64 = (x64 | (x64 << 32)) & 0x1F00000000FFFF
	x64 = (x64 | (x64 << 16)) & 0x1F0000FF0000FF
	x64 = (x64 | (x64 << 8)) & 0x100F00F00F00F00F
	x64 = (x64 | (x64 << 4)) & 0x10C30C30C30C30C3
	x64 = (x64 | (x64 << 2)) & 0x1249249249249249
	return x64
}

func deinterleave3D(x uint64) uint32 {
	x &= 0x1249249249249249
	x = (x | (x >> 2)) & 0x10C30C30C30C30C3
	x = (x | (x >> 4)) & 0x100F00F00F00F00F
	x = (x | (x >> 8)) & 0x1F0000FF0000FF
	x = (x | (x >> 16)) & 0x1F00000000FFFF
	x = (x | (x >> 32)) & 0x00000000001FFFFF
	return uint32(x)
}
