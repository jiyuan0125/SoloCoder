package core

import "errors"

var (
	ErrVarintTruncated  = errors.New("varint truncated")
	ErrVarintTooLong    = errors.New("varint too long (max 10 bytes)")
	ErrVarintOverflow   = errors.New("varint overflow")
)

func ReadVarint(data []byte) (uint64, int, error) {
	var result uint64
	var shift uint
	var i int

	for i = 0; i < len(data); i++ {
		b := data[i]
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, i + 1, nil
		}
		shift += 7
		if shift >= 64 {
			return 0, 0, ErrVarintOverflow
		}
	}

	return 0, 0, ErrVarintTruncated
}

func ZigzagDecode(n uint64) int64 {
	return int64((n >> 1) ^ -(n & 1))
}

func ZigzagDecode32(n uint32) int32 {
	return int32((n >> 1) ^ -(n & 1))
}
