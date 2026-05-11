package mqtt

func EncodeRemainingLength(length uint32) ([]byte, error) {
	if length > MaxRemainingLength {
		return nil, ErrInvalidRemainingLength
	}

	var encoded []byte
	for {
		encodedByte := byte(length % 128)
		length = length / 128
		if length > 0 {
			encodedByte |= 0x80
		}
		encoded = append(encoded, encodedByte)
		if length == 0 {
			break
		}
	}
	return encoded, nil
}

func DecodeRemainingLength(data []byte, offset int) (uint32, int, error) {
	var multiplier uint32 = 1
	var value uint32 = 0
	var encodedByte byte
	var pos = offset

	for {
		if pos >= len(data) {
			return 0, offset, ErrPacketTooShort
		}
		encodedByte = data[pos]
		value += uint32(encodedByte&0x7F) * multiplier
		if multiplier > 128*128*128 {
			return 0, offset, ErrInvalidRemainingLength
		}
		multiplier *= 128
		pos++
		if (encodedByte & 0x80) == 0 {
			break
		}
	}

	return value, pos, nil
}

func EncodeFixedHeader(fh *FixedHeader) ([]byte, error) {
	firstByte := byte(fh.PacketType) << 4
	firstByte |= fh.EncodeFlags()

	rl, err := EncodeRemainingLength(fh.RemainingLength)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 1+len(rl))
	result[0] = firstByte
	copy(result[1:], rl)
	return result, nil
}

func DecodeFixedHeader(data []byte, offset int) (*FixedHeader, int, error) {
	if offset >= len(data) {
		return nil, offset, ErrPacketTooShort
	}

	firstByte := data[offset]
	fh := &FixedHeader{
		PacketType: PacketType((firstByte >> 4) & 0x0F),
	}
	fh.DecodeFlags(firstByte & 0x0F)

	rl, newOffset, err := DecodeRemainingLength(data, offset+1)
	if err != nil {
		return nil, offset, err
	}
	fh.RemainingLength = rl

	return fh, newOffset, nil
}
