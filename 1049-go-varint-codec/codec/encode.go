package codec

import "errors"

var (
	ErrOverflow         = errors.New("value overflow")
	ErrInvalidEncoding  = errors.New("invalid encoding")
	ErrTruncated        = errors.New("truncated data")
)

const (
	maxVarintBytes32  = 5
	maxVarintBytes64  = 10
	maxLEB128Bytes32  = 5
	maxLEB128Bytes64  = 9
)

func EncodeVarintUint32(v uint32) []byte {
	return encodeVarintUint64(uint64(v))
}

func EncodeVarintInt32(v int32) []byte {
	return EncodeVarintUint32(ZigZagEncodeInt32(v))
}

func EncodeVarintUint64(v uint64) []byte {
	return encodeVarintUint64(v)
}

func EncodeVarintInt64(v int64) []byte {
	return EncodeVarintUint64(ZigZagEncodeInt64(v))
}

func encodeVarintUint64(v uint64) []byte {
	var buf []byte
	for v >= 0x80 {
		buf = append(buf, byte(v)|0x80)
		v >>= 7
	}
	buf = append(buf, byte(v))
	return buf
}

func EncodeLEB128Uint32(v uint32) []byte {
	return encodeLEB128Uint64(uint64(v))
}

func EncodeLEB128Int32(v int32) []byte {
	return EncodeLEB128Uint32(ZigZagEncodeInt32(v))
}

func EncodeLEB128Uint64(v uint64) []byte {
	return encodeLEB128Uint64(v)
}

func EncodeLEB128Int64(v int64) []byte {
	return EncodeLEB128Uint64(ZigZagEncodeInt64(v))
}

func encodeLEB128Uint64(v uint64) []byte {
	var buf []byte
	for {
		byteVal := byte(v & 0x7F)
		v >>= 7
		if v != 0 {
			byteVal |= 0x80
		}
		buf = append(buf, byteVal)
		if v == 0 {
			break
		}
	}
	return buf
}
