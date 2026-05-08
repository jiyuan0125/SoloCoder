package codec

func DecodeVarintUint32(data []byte) (uint32, int, error) {
	val, n, err := DecodeVarintUint64(data)
	if err != nil {
		return 0, 0, err
	}
	if val > uint64(uint32Max) {
		return 0, 0, ErrOverflow
	}
	return uint32(val), n, nil
}

func DecodeVarintInt32(data []byte) (int32, int, error) {
	val, n, err := DecodeVarintUint32(data)
	if err != nil {
		return 0, 0, err
	}
	return ZigZagDecodeUint32(val), n, nil
}

func DecodeVarintUint64(data []byte) (uint64, int, error) {
	var result uint64
	var shift uint
	for i, b := range data {
		if i >= maxVarintBytes64 {
			return 0, 0, ErrOverflow
		}
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, i + 1, nil
		}
		shift += 7
	}
	return 0, 0, ErrTruncated
}

func DecodeVarintInt64(data []byte) (int64, int, error) {
	val, n, err := DecodeVarintUint64(data)
	if err != nil {
		return 0, 0, err
	}
	return ZigZagDecodeUint64(val), n, nil
}

func DecodeLEB128Uint32(data []byte) (uint32, int, error) {
	val, n, err := DecodeLEB128Uint64(data)
	if err != nil {
		return 0, 0, err
	}
	if val > uint64(uint32Max) {
		return 0, 0, ErrOverflow
	}
	return uint32(val), n, nil
}

func DecodeLEB128Int32(data []byte) (int32, int, error) {
	val, n, err := DecodeLEB128Uint32(data)
	if err != nil {
		return 0, 0, err
	}
	return ZigZagDecodeUint32(val), n, nil
}

func DecodeLEB128Uint64(data []byte) (uint64, int, error) {
	var result uint64
	var shift uint
	for i, b := range data {
		if i >= maxLEB128Bytes64 {
			return 0, 0, ErrOverflow
		}
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, i + 1, nil
		}
		shift += 7
	}
	return 0, 0, ErrTruncated
}

func DecodeLEB128Int64(data []byte) (int64, int, error) {
	val, n, err := DecodeLEB128Uint64(data)
	if err != nil {
		return 0, 0, err
	}
	return ZigZagDecodeUint64(val), n, nil
}
