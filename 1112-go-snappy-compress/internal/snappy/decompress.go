package snappy

import "errors"

var (
	ErrInvalidStream    = errors.New("snappy: invalid stream identifier")
	ErrCorrupted        = errors.New("snappy: corrupted data")
	ErrUnsupportedChunk = errors.New("snappy: unsupported chunk type")
	ErrOffsetOutOfRange = errors.New("snappy: copy offset out of range")
)

func Decode(data []byte) ([]byte, error) {
	if len(data) < 10 {
		return nil, ErrInvalidStream
	}

	if data[0] != chunkTypeStreamIdentifier {
		return nil, ErrInvalidStream
	}

	idLen := int(data[1]) | int(data[2])<<8 | int(data[3])<<16
	if idLen != len(streamIdentifier) {
		return nil, ErrInvalidStream
	}

	if string(data[4:4+idLen]) != streamIdentifier {
		return nil, ErrInvalidStream
	}

	pos := 4 + idLen
	var result []byte

	for pos < len(data) {
		if pos+4 > len(data) {
			return nil, ErrCorrupted
		}

		chunkType := data[pos]
		chunkLen := int(data[pos+1]) | int(data[pos+2])<<8 | int(data[pos+3])<<16
		pos += 4

		if pos+chunkLen > len(data) {
			return nil, ErrCorrupted
		}

		chunkData := data[pos : pos+chunkLen]
		pos += chunkLen

		switch chunkType {
		case chunkTypeCompressedData:
			uncompressed, err := decompressChunk(chunkData)
			if err != nil {
				return nil, err
			}
			result = append(result, uncompressed...)
		case chunkTypeUncompressedData:
			_, n := readVarint(chunkData)
			if n == 0 {
				return nil, ErrCorrupted
			}
			result = append(result, chunkData[n:]...)
		case chunkTypeStreamIdentifier, chunkTypeDictionary:
			continue
		default:
			return nil, ErrUnsupportedChunk
		}
	}

	return result, nil
}

func decompressChunk(chunkData []byte) ([]byte, error) {
	expectedLen, n := readVarint(chunkData)
	if n == 0 {
		return nil, ErrCorrupted
	}
	chunkData = chunkData[n:]

	result := make([]byte, 0, expectedLen)

	i := 0
	for i < len(chunkData) {
		tag := chunkData[i]
		elementType := tag & 0x03

		switch elementType {
		case elementTypeLiteral:
			length, extraBytes, err := readLiteralLength(tag, chunkData, i)
			if err != nil {
				return nil, err
			}
			i += 1 + extraBytes
			if i+length > len(chunkData) {
				return nil, ErrCorrupted
			}
			result = append(result, chunkData[i:i+length]...)
			i += length

		case elementTypeCopy1ByteOff:
			if i+2 > len(chunkData) {
				return nil, ErrCorrupted
			}
			length := int((tag>>2)&0x07) + 4
			offset := (int(tag&0xe0) << 3) | int(chunkData[i+1])
			if offset == 0 || offset > len(result) {
				return nil, ErrOffsetOutOfRange
			}
			result = appendCopy(result, offset, length)
			i += 2

		case elementTypeCopy2ByteOff:
			if i+3 > len(chunkData) {
				return nil, ErrCorrupted
			}
			length := int((tag>>2)&0x3f) + 4
			offset := int(chunkData[i+1]) | int(chunkData[i+2])<<8
			if offset == 0 || offset > len(result) {
				return nil, ErrOffsetOutOfRange
			}
			result = appendCopy(result, offset, length)
			i += 3

		case elementTypeCopy4ByteOff:
			if i+5 > len(chunkData) {
				return nil, ErrCorrupted
			}
			length := int((tag>>2)&0x3f) + 4
			offset := int(chunkData[i+1]) | int(chunkData[i+2])<<8 | int(chunkData[i+3])<<16 | int(chunkData[i+4])<<24
			if offset == 0 || offset > len(result) {
				return nil, ErrOffsetOutOfRange
			}
			result = appendCopy(result, offset, length)
			i += 5
		}
	}

	return result, nil
}

func readVarint(data []byte) (uint64, int) {
	var result uint64
	var shift uint
	for i, b := range data {
		if i > 10 {
			return 0, 0
		}
		result |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return result, i + 1
		}
		shift += 7
	}
	return 0, 0
}

func readLiteralLength(tag byte, chunkData []byte, pos int) (int, int, error) {
	lenMinus1 := int(tag >> 2)
	if lenMinus1 < 60 {
		return lenMinus1 + 1, 0, nil
	}

	extraBytes := lenMinus1 - 59
	if pos+1+extraBytes > len(chunkData) {
		return 0, 0, ErrCorrupted
	}

	var length int = 60
	for i := 0; i < extraBytes; i++ {
		b := chunkData[pos+1+i]
		length |= int(b) << (i * 8)
	}

	return length, extraBytes, nil
}

func appendCopy(result []byte, offset, length int) []byte {
	start := len(result) - offset
	for i := 0; i < length; i++ {
		result = append(result, result[start+i])
	}
	return result
}
