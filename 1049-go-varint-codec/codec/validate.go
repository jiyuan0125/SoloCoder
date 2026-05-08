package codec

func ValidateVarint(data []byte, numType NumberType) (bool, string) {
	offset := 0
	maxBytes := maxVarintBytes64
	if numType == TypeInt32 || numType == TypeUint32 {
		maxBytes = maxVarintBytes32
	}
	for offset < len(data) {
		n := 0
		for i := offset; i < len(data) && i-offset < maxBytes; i++ {
			if data[i]&0x80 == 0 {
				n = i - offset + 1
				break
			}
		}
		if n == 0 {
			if len(data)-offset >= maxBytes {
				return false, "overflow: value too large"
			}
			return false, "truncated data"
		}
		offset += n
	}
	return true, "valid"
}

func ValidateLEB128(data []byte, numType NumberType) (bool, string) {
	offset := 0
	maxBytes := maxLEB128Bytes64
	if numType == TypeInt32 || numType == TypeUint32 {
		maxBytes = maxLEB128Bytes32
	}
	for offset < len(data) {
		n := 0
		for i := offset; i < len(data) && i-offset < maxBytes; i++ {
			if data[i]&0x80 == 0 {
				n = i - offset + 1
				break
			}
		}
		if n == 0 {
			if len(data)-offset >= maxBytes {
				return false, "overflow: value too large"
			}
			return false, "truncated data"
		}
		offset += n
	}
	return true, "valid"
}
