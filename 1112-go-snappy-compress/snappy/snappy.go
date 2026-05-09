package snappy

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	chunkTypeStreamIdentifier = 0xff
	chunkTypeCompressedData   = 0x00
	chunkTypeUncompressedData = 0x01
	chunkTypeDictionary       = 0x02

	elementTypeLiteral       = 0x00
	elementTypeCopy1Byte     = 0x01
	elementTypeCopy2Byte     = 0x02
	elementTypeCopy4Byte     = 0x03

	streamIdentifier = "sNaPpY"
)

var (
	errInvalidStream       = errors.New("invalid snappy stream")
	errInvalidChunkType    = errors.New("invalid chunk type")
	errInvalidChunkLength  = errors.New("invalid chunk length")
	errInvalidElement      = errors.New("invalid element")
	errInvalidCopyOffset   = errors.New("invalid copy offset")
	errInvalidCopyLength   = errors.New("invalid copy length")
	errCorruptedData       = errors.New("corrupted data")
)

func Encode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer

	if err := writeStreamIdentifier(&buf); err != nil {
		return nil, err
	}

	if err := writeCompressedChunk(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var result bytes.Buffer
	offset := 0

	if err := verifyStreamIdentifier(data, &offset); err != nil {
		return nil, err
	}

	for offset < len(data) {
		if err := readChunk(data, &offset, &result); err != nil {
			return nil, err
		}
	}

	return result.Bytes(), nil
}

func writeStreamIdentifier(buf *bytes.Buffer) error {
	if err := buf.WriteByte(chunkTypeStreamIdentifier); err != nil {
		return err
	}

	length := uint32(len(streamIdentifier))
	lengthBytes := []byte{
		byte(length & 0xff),
		byte((length >> 8) & 0xff),
		byte((length >> 16) & 0xff),
	}
	if _, err := buf.Write(lengthBytes); err != nil {
		return err
	}

	if _, err := buf.WriteString(streamIdentifier); err != nil {
		return err
	}

	return nil
}

func writeCompressedChunk(buf *bytes.Buffer, data []byte) error {
	chunkData := encodeElements(data)
	
	if err := buf.WriteByte(chunkTypeCompressedData); err != nil {
		return err
	}

	length := uint32(len(chunkData))
	lengthBytes := []byte{
		byte(length & 0xff),
		byte((length >> 8) & 0xff),
		byte((length >> 16) & 0xff),
	}
	if _, err := buf.Write(lengthBytes); err != nil {
		return err
	}

	if _, err := buf.Write(chunkData); err != nil {
		return err
	}

	return nil
}

func encodeElements(data []byte) []byte {
	var result bytes.Buffer
	i := 0
	n := len(data)

	for i < n {
		matchLen, matchOffset := findMatch(data, i)
		
		if matchLen >= 4 && matchOffset > 0 {
			if matchOffset <= 2048 {
				writeCopy1Byte(&result, matchOffset, matchLen)
			} else if matchOffset <= 65536 {
				writeCopy2Byte(&result, matchOffset, matchLen)
			} else {
				writeCopy4Byte(&result, matchOffset, matchLen)
			}
			i += matchLen
		} else {
			literalStart := i
			for i < n {
				nextMatchLen, _ := findMatch(data, i)
				if nextMatchLen >= 4 {
					break
				}
				i++
			}
			writeLiteral(&result, data[literalStart:i])
		}
	}

	return result.Bytes()
}

func findMatch(data []byte, pos int) (int, int) {
	if pos < 4 {
		return 0, 0
	}

	maxLen := 0
	maxOffset := 0
	maxSearch := min(pos, 65536)

	for offset := 1; offset <= maxSearch; offset++ {
		length := 0
		for pos+length < len(data) && data[pos+length] == data[pos-offset+length] {
			length++
		}
		if length > maxLen {
			maxLen = length
			maxOffset = offset
			if maxLen >= 64 {
				break
			}
		}
	}

	return maxLen, maxOffset
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func writeLiteral(buf *bytes.Buffer, data []byte) {
	length := len(data) - 1

	if length < 60 {
		byte1 := byte(length<<2) | elementTypeLiteral
		buf.WriteByte(byte1)
	} else {
		var lengthBytes []byte
		if length < 256 {
			lengthBytes = []byte{byte(60<<2) | elementTypeLiteral, byte(length)}
		} else if length < 65536 {
			lengthBytes = []byte{byte(61<<2) | elementTypeLiteral, byte(length & 0xff), byte((length >> 8) & 0xff)}
		} else {
			lengthBytes = []byte{
				byte(62<<2) | elementTypeLiteral,
				byte(length & 0xff),
				byte((length >> 8) & 0xff),
				byte((length >> 16) & 0xff),
			}
		}
		buf.Write(lengthBytes)
	}
	buf.Write(data)
}

func writeCopy1Byte(buf *bytes.Buffer, offset, length int) {
	encodedLen := min(length-4, 7)
	byte1 := byte(encodedLen<<2) | elementTypeCopy1Byte
	byte2 := byte(offset & 0xff)
	buf.WriteByte(byte1)
	buf.WriteByte(byte2)

	if length > 11 {
		extra := length - 11
		var extraBytes []byte
		for extra > 0 {
			extraBytes = append(extraBytes, byte(extra&0xff))
			extra >>= 8
		}
		buf.Write(extraBytes)
	}
}

func writeCopy2Byte(buf *bytes.Buffer, offset, length int) {
	encodedLen := min(length-4, 63)
	byte1 := byte(encodedLen<<2) | elementTypeCopy2Byte
	offsetBytes := []byte{byte(offset & 0xff), byte((offset >> 8) & 0xff)}
	buf.WriteByte(byte1)
	buf.Write(offsetBytes)

	if length > 67 {
		extra := length - 67
		var extraBytes []byte
		for extra > 0 {
			extraBytes = append(extraBytes, byte(extra&0xff))
			extra >>= 8
		}
		buf.Write(extraBytes)
	}
}

func writeCopy4Byte(buf *bytes.Buffer, offset, length int) {
	encodedLen := min(length-4, 63)
	byte1 := byte(encodedLen<<2) | elementTypeCopy4Byte
	offsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(offsetBytes, uint32(offset))
	buf.WriteByte(byte1)
	buf.Write(offsetBytes)

	if length > 67 {
		extra := length - 67
		var extraBytes []byte
		for extra > 0 {
			extraBytes = append(extraBytes, byte(extra&0xff))
			extra >>= 8
		}
		buf.Write(extraBytes)
	}
}

func verifyStreamIdentifier(data []byte, offset *int) error {
	if len(data) < 10 {
		return errInvalidStream
	}

	if data[*offset] != chunkTypeStreamIdentifier {
		return errInvalidStream
	}
	*offset++

	chunkLength := int(data[*offset]) | (int(data[*offset+1]) << 8) | (int(data[*offset+2]) << 16)
	*offset += 3

	if chunkLength != len(streamIdentifier) {
		return errInvalidStream
	}

	if string(data[*offset:*offset+chunkLength]) != streamIdentifier {
		return errInvalidStream
	}
	*offset += chunkLength

	return nil
}

func readChunk(data []byte, offset *int, result *bytes.Buffer) error {
	if *offset >= len(data) {
		return nil
	}

	chunkType := data[*offset]
	*offset++

	if *offset+3 > len(data) {
		return errInvalidChunkLength
	}
	chunkLength := int(data[*offset]) | (int(data[*offset+1]) << 8) | (int(data[*offset+2]) << 16)
	*offset += 3

	if *offset+chunkLength > len(data) {
		return errInvalidChunkLength
	}
	chunkData := data[*offset : *offset+chunkLength]
	*offset += chunkLength

	switch chunkType {
	case chunkTypeCompressedData:
		return decodeElements(chunkData, result)
	case chunkTypeUncompressedData:
		result.Write(chunkData)
		return nil
	case chunkTypeDictionary:
		return errors.New("dictionary chunks not supported")
	case chunkTypeStreamIdentifier:
		if string(chunkData) != streamIdentifier {
			return errInvalidStream
		}
		return nil
	default:
		return errInvalidChunkType
	}
}

func decodeElements(data []byte, result *bytes.Buffer) error {
	offset := 0
	n := len(data)

	for offset < n {
		byte1 := data[offset]
		elementType := byte1 & 0x03

		switch elementType {
		case elementTypeLiteral:
			length, bytesRead := readLiteralLength(data, offset)
			offset += bytesRead
			if offset+length > n {
				return errInvalidElement
			}
			result.Write(data[offset : offset+length])
			offset += length

		case elementTypeCopy1Byte:
			if offset+2 > n {
				return errInvalidElement
			}
			length := int((byte1>>2)&0x07) + 4
			offsetVal := int(data[offset+1])
			offset += 2

			if length == 11 {
				extra, extraBytes := readVarint(data, offset)
				length += extra
				offset += extraBytes
			}

			if err := copyData(result, offsetVal, length); err != nil {
				return err
			}

		case elementTypeCopy2Byte:
			if offset+3 > n {
				return errInvalidElement
			}
			length := int((byte1>>2)&0x3f) + 4
			offsetVal := int(data[offset+1]) | (int(data[offset+2]) << 8)
			offset += 3

			if length == 67 {
				extra, extraBytes := readVarint(data, offset)
				length += extra
				offset += extraBytes
			}

			if err := copyData(result, offsetVal, length); err != nil {
				return err
			}

		case elementTypeCopy4Byte:
			if offset+5 > n {
				return errInvalidElement
			}
			length := int((byte1>>2)&0x3f) + 4
			offsetVal := int(binary.LittleEndian.Uint32(data[offset+1 : offset+5]))
			offset += 5

			if length == 67 {
				extra, extraBytes := readVarint(data, offset)
				length += extra
				offset += extraBytes
			}

			if err := copyData(result, offsetVal, length); err != nil {
				return err
			}
		}
	}

	return nil
}

func readLiteralLength(data []byte, offset int) (int, int) {
	byte1 := data[offset]
	lengthType := (byte1 >> 2) & 0x3f

	if lengthType < 60 {
		return int(lengthType) + 1, 1
	}

	extraBytes := int(lengthType - 59)
	length := 1
	for i := 0; i < extraBytes; i++ {
		if offset+1+i >= len(data) {
			return 0, 0
		}
		length += int(data[offset+1+i]) << (uint(i) * 8)
	}
	return length, 1 + extraBytes
}

func readVarint(data []byte, offset int) (int, int) {
	result := 0
	shift := 0
	for i := 0; offset+i < len(data); i++ {
		b := data[offset+i]
		result |= int(b) << shift
		if b&0x80 == 0 {
			return result, i + 1
		}
		shift += 7
		if shift > 35 {
			return 0, 0
		}
	}
	return 0, 0
}

func copyData(result *bytes.Buffer, offset int, length int) error {
	resultBytes := result.Bytes()
	if offset > len(resultBytes) {
		return errInvalidCopyOffset
	}

	start := len(resultBytes) - offset
	for i := 0; i < length; i++ {
		result.WriteByte(resultBytes[start+i])
	}

	return nil
}
