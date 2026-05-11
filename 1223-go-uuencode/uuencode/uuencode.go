package uuencode

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const lineLength = 45

func Encode(data []byte, filename string, mode int) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("begin %d %s\n", mode, filename))

	for i := 0; i < len(data); i += lineLength {
		end := i + lineLength
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		lengthChar := byte(len(chunk) + 32)
		buf.WriteByte(lengthChar)

		encodedLine := encodeChunk(chunk)
		buf.WriteString(encodedLine)
		buf.WriteByte('\n')
	}

	buf.WriteString("`\n")
	buf.WriteString("end\n")

	return buf.String()
}

func encodeChunk(chunk []byte) string {
	var result bytes.Buffer

	for i := 0; i < len(chunk); i += 3 {
		var b1, b2, b3 byte
		if i < len(chunk) {
			b1 = chunk[i]
		}
		if i+1 < len(chunk) {
			b2 = chunk[i+1]
		}
		if i+2 < len(chunk) {
			b3 = chunk[i+2]
		}

		combined := (uint32(b1) << 16) | (uint32(b2) << 8) | uint32(b3)

		c1 := (combined >> 18) & 0x3F
		c2 := (combined >> 12) & 0x3F
		c3 := (combined >> 6) & 0x3F
		c4 := combined & 0x3F

		result.WriteByte(byte(c1 + 32))
		result.WriteByte(byte(c2 + 32))
		result.WriteByte(byte(c3 + 32))
		result.WriteByte(byte(c4 + 32))
	}

	return result.String()
}

type DecodeResult struct {
	Data     []byte
	Filename string
	Mode     int
}

func Decode(encoded string) (*DecodeResult, error) {
	lines := strings.Split(encoded, "\n")

	beginLineIndex := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "begin ") {
			beginLineIndex = i
			break
		}
	}

	if beginLineIndex == -1 {
		return nil, errors.New("begin line not found")
	}

	beginParts := strings.SplitN(lines[beginLineIndex], " ", 3)
	if len(beginParts) < 3 {
		return nil, errors.New("invalid begin line format")
	}

	mode, err := strconv.Atoi(beginParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid mode: %w", err)
	}

	filename := beginParts[2]

	endLineIndex := -1
	for i := beginLineIndex + 1; i < len(lines); i++ {
		if lines[i] == "end" {
			endLineIndex = i
			break
		}
	}

	if endLineIndex == -1 {
		return nil, errors.New("end line not found")
	}

	var data bytes.Buffer

	for i := beginLineIndex + 1; i < endLineIndex; i++ {
		line := lines[i]
		if len(line) == 0 {
			continue
		}

		lengthChar := line[0]
		var length int
		if lengthChar == 96 {
			length = 0
		} else {
			length = int(lengthChar - 32)
		}

		if length == 0 {
			continue
		}

		if length > lineLength {
			continue
		}

		if len(line) > 1 {
			decoded, err := decodeLine(line[1:], length)
			if err != nil {
				return nil, err
			}
			data.Write(decoded)
		}
	}

	return &DecodeResult{
		Data:     data.Bytes(),
		Filename: filename,
		Mode:     mode,
	}, nil
}

func decodeLine(encodedLine string, expectedLength int) ([]byte, error) {
	var result bytes.Buffer

	for i := 0; i < len(encodedLine); i += 4 {
		var c1, c2, c3, c4 byte
		if i < len(encodedLine) {
			c1 = encodedLine[i]
		}
		if i+1 < len(encodedLine) {
			c2 = encodedLine[i+1]
		}
		if i+2 < len(encodedLine) {
			c3 = encodedLine[i+2]
		}
		if i+3 < len(encodedLine) {
			c4 = encodedLine[i+3]
		}

		toValue := func(b byte) uint32 {
			if b == 96 {
				return 0
			}
			return uint32(b - 32)
		}

		n1 := toValue(c1)
		n2 := toValue(c2)
		n3 := toValue(c3)
		n4 := toValue(c4)

		combined := (n1 << 18) | (n2 << 12) | (n3 << 6) | n4

		b1 := byte((combined >> 16) & 0xFF)
		b2 := byte((combined >> 8) & 0xFF)
		b3 := byte(combined & 0xFF)

		result.WriteByte(b1)
		result.WriteByte(b2)
		result.WriteByte(b3)
	}

	decoded := result.Bytes()
	if len(decoded) > expectedLength {
		decoded = decoded[:expectedLength]
	}

	return decoded, nil
}
