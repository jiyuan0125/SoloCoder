package ascii85

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Mode string

const (
	ModeAdobe Mode = "adobe"
	ModeBtoa  Mode = "btoa"
)

func Encode(data []byte, mode Mode) (string, error) {
	if mode != ModeAdobe && mode != ModeBtoa {
		return "", errors.New("invalid mode: must be 'adobe' or 'btoa'")
	}

	if len(data) == 0 {
		if mode == ModeAdobe {
			return "<~~>", nil
		}
		return "", nil
	}

	var buf bytes.Buffer

	if mode == ModeAdobe {
		buf.WriteString("<~")
	} else {
		buf.WriteString("xbtoa Begin\n")
	}

	n := len(data)
	fullGroups := n / 4
	remaining := n % 4

	for i := 0; i < fullGroups; i++ {
		offset := i * 4
		b0 := data[offset]
		b1 := data[offset+1]
		b2 := data[offset+2]
		b3 := data[offset+3]

		if b0 == 0 && b1 == 0 && b2 == 0 && b3 == 0 {
			buf.WriteByte('z')
			continue
		}

		val := uint32(b0)<<24 | uint32(b1)<<16 | uint32(b2)<<8 | uint32(b3)
		encodeGroup(&buf, val)
	}

	if remaining > 0 {
		var b0, b1, b2, b3 byte = 0, 0, 0, 0
		offset := fullGroups * 4

		if remaining >= 1 {
			b0 = data[offset]
		}
		if remaining >= 2 {
			b1 = data[offset+1]
		}
		if remaining >= 3 {
			b2 = data[offset+2]
		}

		val := uint32(b0)<<24 | uint32(b1)<<16 | uint32(b2)<<8 | uint32(b3)
		chars := encodeGroupToString(val)
		buf.WriteString(chars[:remaining+1])
	}

	if mode == ModeAdobe {
		buf.WriteString("~>")
	} else {
		buf.WriteString("\nxbtoa End\n")
	}

	return buf.String(), nil
}

func encodeGroup(buf *bytes.Buffer, val uint32) {
	pow85_4 := uint32(85 * 85 * 85 * 85)
	pow85_3 := uint32(85 * 85 * 85)
	pow85_2 := uint32(85 * 85)
	pow85_1 := uint32(85)

	c0 := val / pow85_4
	rem := val % pow85_4
	c1 := rem / pow85_3
	rem = rem % pow85_3
	c2 := rem / pow85_2
	rem = rem % pow85_2
	c3 := rem / pow85_1
	c4 := rem % pow85_1

	buf.WriteByte(byte(c0) + '!')
	buf.WriteByte(byte(c1) + '!')
	buf.WriteByte(byte(c2) + '!')
	buf.WriteByte(byte(c3) + '!')
	buf.WriteByte(byte(c4) + '!')
}

func encodeGroupToString(val uint32) string {
	pow85_4 := uint32(85 * 85 * 85 * 85)
	pow85_3 := uint32(85 * 85 * 85)
	pow85_2 := uint32(85 * 85)
	pow85_1 := uint32(85)

	c0 := val / pow85_4
	rem := val % pow85_4
	c1 := rem / pow85_3
	rem = rem % pow85_3
	c2 := rem / pow85_2
	rem = rem % pow85_2
	c3 := rem / pow85_1
	c4 := rem % pow85_1

	return string([]byte{
		byte(c0) + '!',
		byte(c1) + '!',
		byte(c2) + '!',
		byte(c3) + '!',
		byte(c4) + '!',
	})
}

func Decode(input string) ([]byte, Mode, error) {
	if len(input) == 0 {
		return nil, "", errors.New("empty input")
	}

	var mode Mode
	var encodedData string

	if strings.HasPrefix(input, "<~") {
		mode = ModeAdobe
		var err error
		encodedData, err = extractAdobeData(input)
		if err != nil {
			return nil, "", err
		}
	} else {
		mode = ModeBtoa
		var err error
		encodedData, err = extractBtoaData(input)
		if err != nil {
			return nil, "", err
		}
	}

	if len(encodedData) == 0 {
		return []byte{}, mode, nil
	}

	result, err := decodeEncodedData(encodedData)
	if err != nil {
		return nil, "", err
	}

	return result, mode, nil
}

func extractAdobeData(input string) (string, error) {
	trimmed := strings.TrimSpace(input)

	if !strings.HasPrefix(trimmed, "<~") {
		return "", errors.New("invalid Adobe format: missing <~")
	}
	if !strings.HasSuffix(trimmed, "~>") {
		return "", errors.New("invalid Adobe format: missing ~>")
	}

	content := trimmed[2 : len(trimmed)-2]
	return removeWhitespace(content), nil
}

func extractBtoaData(input string) (string, error) {
	lines := strings.Split(input, "\n")

	startIdx := -1
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "xbtoa Begin") {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return "", errors.New("invalid btoa format: missing 'xbtoa Begin'")
	}

	endIdx := -1
	for i := startIdx + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "xbtoa End") {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return "", errors.New("invalid btoa format: missing 'xbtoa End'")
	}

	firstDataLine := startIdx + 1
	if firstDataLine < endIdx {
		firstLine := strings.TrimSpace(lines[firstDataLine])
		if !isAscii85Data(firstLine) {
			firstDataLine++
		}
	}

	var dataBuf bytes.Buffer
	for i := firstDataLine; i < endIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		dataBuf.WriteString(line)
	}

	return removeWhitespace(dataBuf.String()), nil
}

func isAscii85Data(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !isAscii85Char(c) {
			return false
		}
	}
	return true
}

func isAscii85Char(c rune) bool {
	return (c >= '!' && c <= 'u') || c == 'z'
}

func removeWhitespace(s string) string {
	var builder strings.Builder
	builder.Grow(len(s))
	for _, c := range s {
		if !isWhitespace(c) {
			builder.WriteRune(c)
		}
	}
	return builder.String()
}

func isWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func decodeEncodedData(encoded string) ([]byte, error) {
	if len(encoded) == 0 {
		return []byte{}, nil
	}

	var result bytes.Buffer
	i := 0
	n := len(encoded)

	for i < n {
		if encoded[i] == 'z' {
			result.Write([]byte{0, 0, 0, 0})
			i++
			continue
		}

		var groupChars [5]byte
		groupLen := 0

		for j := 0; j < 5 && i < n; j++ {
			if encoded[i] == 'z' {
				if groupLen > 0 {
					return nil, errors.New("z compression not allowed in partial group")
				}
				return nil, errors.New("unexpected z in data stream")
			}
			if !isValidChar(encoded[i]) {
				return nil, fmt.Errorf("invalid character: %c", encoded[i])
			}
			groupChars[j] = encoded[i]
			groupLen++
			i++
		}

		if groupLen == 1 {
			return nil, errors.New("invalid encoded data: single character at end")
		}

		var val uint32
		powers := []uint32{85 * 85 * 85 * 85, 85 * 85 * 85, 85 * 85, 85, 1}
		for j := 0; j < groupLen; j++ {
			c := groupChars[j] - '!'
			if c > 84 {
				return nil, fmt.Errorf("character out of range: %c", groupChars[j])
			}
			if j < 5 {
				val += uint32(c) * powers[j]
			}
		}

		if groupLen == 5 {
			result.WriteByte(byte(val >> 24))
			result.WriteByte(byte(val >> 16))
			result.WriteByte(byte(val >> 8))
			result.WriteByte(byte(val))
		} else if groupLen == 2 {
			val += powers[1] + powers[2] + powers[3] + powers[4]
			result.WriteByte(byte(val >> 24))
		} else if groupLen == 3 {
			val += powers[2] + powers[3] + powers[4]
			result.WriteByte(byte(val >> 24))
			result.WriteByte(byte(val >> 16))
		} else if groupLen == 4 {
			val += powers[3] + powers[4]
			result.WriteByte(byte(val >> 24))
			result.WriteByte(byte(val >> 16))
			result.WriteByte(byte(val >> 8))
		}
	}

	return result.Bytes(), nil
}

func isValidChar(b byte) bool {
	return b >= '!' && b <= 'u'
}

func NewEncoder(mode Mode) io.WriteCloser {
	return &encoder{mode: mode, buf: &bytes.Buffer{}}
}

type encoder struct {
	mode Mode
	buf  *bytes.Buffer
	wroteHeader bool
}

func (e *encoder) Write(p []byte) (n int, err error) {
	if !e.wroteHeader {
		if e.mode == ModeAdobe {
			e.buf.WriteString("<~")
		} else {
			e.buf.WriteString("xbtoa Begin\n")
		}
		e.wroteHeader = true
	}
	n, err = e.buf.Write(p)
	return n, err
}

func (e *encoder) Close() error {
	return nil
}

func NewDecoder() io.Reader {
	return &decoder{}
}

type decoder struct{}

func (d *decoder) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
