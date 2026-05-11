package yenc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"strconv"
)

const (
	DefaultLineSize = 128
	escapeChar      = 61
	ybeginLine      = "=ybegin"
	ypartLine       = "=ypart"
	yendLine        = "=yend"
)

type Options struct {
	LineSize int
}

type EncodeResult struct {
	Text     string
	TotalCRC uint32
	TotalSize int
}

func Encode(data []byte) string {
	result, _ := EncodeWithOptions(data, nil)
	return result.Text
}

func EncodeWithOptions(data []byte, opts *Options) (*EncodeResult, error) {
	lineSize := DefaultLineSize
	if opts != nil && opts.LineSize > 0 {
		lineSize = opts.LineSize
	}

	totalSize := len(data)
	totalCRC := crc32.ChecksumIEEE(data)

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("=ybegin size=%d line=%d\n", totalSize, lineSize))

	if totalSize <= lineSize {
		partCRC := crc32.ChecksumIEEE(data)
		encodedPart := encodeDataLine(data)
		buf.WriteString(encodedPart)
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("=yend size=%d crc32=%s\n", totalSize, hex.EncodeToString([]byte{byte(partCRC >> 24), byte(partCRC >> 16), byte(partCRC >> 8), byte(partCRC)})))
	} else {
		for start := 0; start < totalSize; start += lineSize {
			end := start + lineSize
			if end > totalSize {
				end = totalSize
			}
			partData := data[start:end]
			partCRC := crc32.ChecksumIEEE(partData)

			buf.WriteString(fmt.Sprintf("=ypart begin=%d end=%d\n", start+1, end))
			encodedPart := encodeDataLine(partData)
			buf.WriteString(encodedPart)
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("=yend size=%d crc32=%s\n", len(partData), hex.EncodeToString([]byte{byte(partCRC >> 24), byte(partCRC >> 16), byte(partCRC >> 8), byte(partCRC)})))
		}
	}

	buf.WriteString(fmt.Sprintf("=yend size=%d crc32=%s\n", totalSize, hex.EncodeToString([]byte{byte(totalCRC >> 24), byte(totalCRC >> 16), byte(totalCRC >> 8), byte(totalCRC)})))

	return &EncodeResult{
		Text:     buf.String(),
		TotalCRC: totalCRC,
		TotalSize: totalSize,
	}, nil
}

func encodeDataLine(data []byte) string {
	var buf bytes.Buffer
	for _, b := range data {
		encoded := (int(b) + 42) % 256

		if b == 0 || b == 10 {
			buf.WriteByte(byte(encoded))
			continue
		}

		if encoded < 33 || encoded > 126 || encoded == escapeChar {
			buf.WriteByte(escapeChar)
			buf.WriteByte(byte((encoded + 42) % 256))
		} else {
			buf.WriteByte(byte(encoded))
		}
	}
	return buf.String()
}

func Decode(text string) ([]byte, error) {
	rawLines := splitLines(text)
	if len(rawLines) == 0 {
		return nil, io.EOF
	}

	var lines []string
	var i int
	for i < len(rawLines) {
		line := rawLines[i]
		if len(line) > 0 && line[len(line)-1] == '=' && i < len(rawLines)-1 {
			lines = append(lines, line+"\n"+rawLines[i+1])
			i += 2
		} else {
			lines = append(lines, line)
			i++
		}
	}

	var result bytes.Buffer
	var expectedTotalSize int
	var foundYBegin bool
	var collectingData bool
	var partData bytes.Buffer

	for _, line := range lines {
		line = trimLine(line)
		if line == "" {
			continue
		}

		if len(line) >= 7 && line[:7] == ybeginLine {
			foundYBegin = true
			expectedTotalSize = parseSizeFromYBegin(line)
			continue
		}

		if len(line) >= 6 && line[:6] == ypartLine {
			collectingData = true
			partData.Reset()
			continue
		}

		if len(line) >= 5 && line[:5] == yendLine {
			if collectingData {
				if partData.Len() > 0 {
					decoded, err := decodeDataLine(partData.String())
					if err != nil {
						return nil, err
					}
					result.Write(decoded)
				}
				collectingData = false
				continue
			} else if foundYBegin {
				_, _ = parseYEnd(line)
				continue
			}
		}

		if collectingData {
			partData.WriteString(line)
		} else if foundYBegin {
			decoded, err := decodeDataLine(line)
			if err != nil {
				return nil, err
			}
			result.Write(decoded)
		}
	}

	if expectedTotalSize > 0 && result.Len() != expectedTotalSize {
		return nil, fmt.Errorf("decoded size %d does not match expected %d", result.Len(), expectedTotalSize)
	}

	return result.Bytes(), nil
}

func decodeDataLine(line string) ([]byte, error) {
	var result []byte
	i := 0
	for i < len(line) {
		if line[i] == escapeChar {
			if i+1 >= len(line) {
				return nil, fmt.Errorf("invalid escape sequence at position %d", i)
			}
			second := int(line[i+1])
			firstDecoded := (second - 42) % 256
			if firstDecoded < 0 {
				firstDecoded += 256
			}
			original := (firstDecoded - 42) % 256
			if original < 0 {
				original += 256
			}
			result = append(result, byte(original))
			i += 2
		} else {
			original := (int(line[i]) - 42) % 256
			if original < 0 {
				original += 256
			}
			result = append(result, byte(original))
			i++
		}
	}
	return result, nil
}

func splitLines(text string) []string {
	var lines []string
	var start int
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			if i > start {
				lines = append(lines, text[start:i])
			}
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

func trimLine(line string) string {
	return line
}

func parseSizeFromYBegin(line string) int {
	parts := splitAndParseParams(line)
	if size, ok := parts["size"]; ok {
		if n, err := strconv.Atoi(size); err == nil {
			return n
		}
	}
	return 0
}

func parseYPart(line string) (int, int) {
	parts := splitAndParseParams(line)
	begin := 0
	end := 0
	if s, ok := parts["begin"]; ok {
		if n, err := strconv.Atoi(s); err == nil {
			begin = n
		}
	}
	if s, ok := parts["end"]; ok {
		if n, err := strconv.Atoi(s); err == nil {
			end = n
		}
	}
	return begin, end
}

func parseYEnd(line string) (int, uint32) {
	parts := splitAndParseParams(line)
	size := 0
	crc := uint32(0)
	if s, ok := parts["size"]; ok {
		if n, err := strconv.Atoi(s); err == nil {
			size = n
		}
	}
	if s, ok := parts["crc32"]; ok {
		if len(s) == 8 {
			if bytes, err := hex.DecodeString(s); err == nil && len(bytes) == 4 {
				crc = uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
			}
		}
	}
	return size, crc
}

func splitAndParseParams(line string) map[string]string {
	result := make(map[string]string)
	fields := splitByWhitespace(line)
	for _, field := range fields {
		idx := indexOfEquals(field)
		if idx > 0 && idx < len(field)-1 {
			key := field[:idx]
			value := field[idx+1:]
			result[key] = value
		}
	}
	return result
}

func splitByWhitespace(s string) []string {
	var result []string
	var buf bytes.Buffer
	for _, ch := range s {
		if ch == ' ' || ch == '\t' {
			if buf.Len() > 0 {
				result = append(result, buf.String())
				buf.Reset()
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	if buf.Len() > 0 {
		result = append(result, buf.String())
	}
	return result
}

func indexOfEquals(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return i
		}
	}
	return -1
}
