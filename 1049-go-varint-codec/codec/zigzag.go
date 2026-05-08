package codec

const (
	uint32Max = 0xFFFFFFFF
	uint64Max = 0xFFFFFFFFFFFFFFFF
	int32Max  = 0x7FFFFFFF
	int32Min  = -0x80000000
	int64Max  = 0x7FFFFFFFFFFFFFFF
	int64Min  = -0x8000000000000000
)

func ZigZagEncodeInt32(v int32) uint32 {
	return uint32((v << 1) ^ (v >> 31))
}

func ZigZagDecodeUint32(v uint32) int32 {
	return int32((v >> 1) ^ -(v & 1))
}

func ZigZagEncodeInt64(v int64) uint64 {
	return uint64((v << 1) ^ (v >> 63))
}

func ZigZagDecodeUint64(v uint64) int64 {
	sign := int64(v & 1)
	return int64(v>>1) ^ -sign
}
