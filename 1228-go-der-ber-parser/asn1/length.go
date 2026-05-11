package asn1

import (
	"errors"
)

const IndefiniteLength = 0x80

func EncodeLength(length int, definite bool) []byte {
	if !definite {
		return []byte{IndefiniteLength}
	}

	if length < 0x80 {
		return []byte{byte(length)}
	}

	var bytes []byte
	n := length
	for n > 0 {
		bytes = append([]byte{byte(n & 0xFF)}, bytes...)
		n >>= 8
	}

	firstByte := byte(len(bytes)) | 0x80
	return append([]byte{firstByte}, bytes...)
}

func ParseLength(data []byte) (length int, lenLen int, indefinite bool, err error) {
	if len(data) < 1 {
		return 0, 0, false, errors.New("insufficient data for length")
	}

	firstByte := data[0]
	if firstByte == IndefiniteLength {
		return 0, 1, true, nil
	}

	if (firstByte & 0x80) == 0 {
		return int(firstByte), 1, false, nil
	}

	numBytes := int(firstByte & 0x7F)
	if numBytes == 0 {
		return 0, 0, false, errors.New("invalid length encoding")
	}
	if numBytes > 4 {
		return 0, 0, false, errors.New("length too large")
	}

	if len(data) < 1+numBytes {
		return 0, 0, false, errors.New("insufficient data for length")
	}

	var lengthVal int
	for i := 0; i < numBytes; i++ {
		lengthVal = (lengthVal << 8) | int(data[1+i])
	}

	return lengthVal, 1 + numBytes, false, nil
}
