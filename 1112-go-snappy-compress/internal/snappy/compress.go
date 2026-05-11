package snappy

const (
	chunkTypeStreamIdentifier = 0xff
	chunkTypeCompressedData   = 0x00
	chunkTypeUncompressedData = 0x01
	chunkTypeDictionary       = 0x02

	elementTypeLiteral      = 0x00
	elementTypeCopy1ByteOff = 0x01
	elementTypeCopy2ByteOff = 0x02
	elementTypeCopy4ByteOff = 0x03

	streamIdentifier = "sNaPpY"

	maxUncompressedChunkSize = 65536
)

func Encode(data []byte) []byte {
	result := encodeStreamIdentifier()

	if len(data) == 0 {
		return result
	}

	for i := 0; i < len(data); i += maxUncompressedChunkSize {
		end := i + maxUncompressedChunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		compressed := compressChunkSimple(chunk)

		chunkData := make([]byte, 0, len(compressed)+7)
		chunkData = appendVarint(chunkData, uint64(len(chunk)))
		chunkData = append(chunkData, compressed...)

		chunkHeader := encodeChunkHeader(chunkTypeCompressedData, uint32(len(chunkData)))
		result = append(result, chunkHeader...)
		result = append(result, chunkData...)
	}

	return result
}

func encodeStreamIdentifier() []byte {
	id := []byte(streamIdentifier)
	header := encodeChunkHeader(chunkTypeStreamIdentifier, uint32(len(id)))
	return append(header, id...)
}

func encodeChunkHeader(chunkType byte, length uint32) []byte {
	header := make([]byte, 4)
	header[0] = chunkType
	header[1] = byte(length)
	header[2] = byte(length >> 8)
	header[3] = byte(length >> 16)
	return header
}

func compressChunkSimple(data []byte) []byte {
	var result []byte

	if len(data) == 0 {
		return result
	}

	literalStart := 0
	i := 0

	for i < len(data) {
		offset, matchLen := findMatch(data, i)

		if matchLen >= 4 {
			if i > literalStart {
				result = appendLiteral(result, data[literalStart:i])
			}

			for matchLen > 0 {
				var copyLen int
				var use2Byte bool

				if offset <= 2048 && matchLen <= 11 {
					copyLen = matchLen
					use2Byte = false
				} else {
					copyLen = matchLen
					if copyLen > 67 {
						copyLen = 67
					}
					use2Byte = true
				}

				if use2Byte {
					if offset <= 65535 {
						result = appendCopy2Byte(result, offset, copyLen)
					} else {
						result = appendCopy4Byte(result, offset, copyLen)
					}
				} else {
					result = appendCopy1Byte(result, offset, copyLen)
				}

				matchLen -= copyLen
				i += copyLen
			}
			literalStart = i
		} else {
			i++
			if i-literalStart >= 60 {
				result = appendLiteral(result, data[literalStart:i])
				literalStart = i
			}
		}
	}

	if literalStart < len(data) {
		result = appendLiteral(result, data[literalStart:])
	}

	return result
}

func findMatch(data []byte, pos int) (int, int) {
	if pos < 4 {
		return 0, 0
	}

	bestOffset := 0
	bestLen := 0
	maxOffset := pos
	if maxOffset > 65535 {
		maxOffset = 65535
	}
	maxLen := len(data) - pos
	if maxLen > 64 {
		maxLen = 64
	}

	for offset := 1; offset <= maxOffset; offset++ {
		l := 0
		for l < maxLen && data[pos+l] == data[pos-offset+l] {
			l++
		}
		if l > bestLen && l >= 4 {
			bestLen = l
			bestOffset = offset
			if bestLen == maxLen {
				break
			}
		}
	}

	return bestOffset, bestLen
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func appendLiteral(result, literal []byte) []byte {
	length := len(literal)
	if length <= 60 {
		header := byte(elementTypeLiteral) | byte((length-1)<<2)
		result = append(result, header)
	} else {
		header := byte(elementTypeLiteral) | byte(59<<2)
		result = append(result, header)
		remaining := length - 60
		for {
			if remaining < 128 {
				result = append(result, byte(remaining))
				break
			}
			result = append(result, byte(remaining&0x7f)|0x80)
			remaining >>= 7
		}
	}
	result = append(result, literal...)
	return result
}

func appendCopy1Byte(result []byte, offset, length int) []byte {
	lengthMinus4 := length - 4
	header := byte(elementTypeCopy1ByteOff) | byte(lengthMinus4<<2) | byte((offset>>3)&0xe0)
	result = append(result, header, byte(offset&0xff))
	return result
}

func appendCopy2Byte(result []byte, offset, length int) []byte {
	lengthMinus4 := length - 4
	header := byte(elementTypeCopy2ByteOff) | byte(lengthMinus4<<2)
	result = append(result, header, byte(offset), byte(offset>>8))
	return result
}

func appendCopy4Byte(result []byte, offset, length int) []byte {
	lengthMinus4 := length - 4
	header := byte(elementTypeCopy4ByteOff) | byte(lengthMinus4<<2)
	result = append(result, header, byte(offset), byte(offset>>8), byte(offset>>16), byte(offset>>24))
	return result
}

func appendVarint(result []byte, value uint64) []byte {
	for value >= 0x80 {
		result = append(result, byte(value&0x7f)|0x80)
		value >>= 7
	}
	result = append(result, byte(value))
	return result
}
