package encoder

import (
	"bytes"
	"errors"
	"unicode/utf8"
)

const hexChars = "0123456789ABCDEF"

func isUnreserved(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~'
}

func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

func hexToByte(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'A' && c <= 'F' {
		return 10 + (c - 'A')
	}
	if c >= 'a' && c <= 'f' {
		return 10 + (c - 'a')
	}
	return 0
}

func isCompletePercentSequence(s string, i int) bool {
	if i+2 >= len(s) {
		return false
	}
	return isHexChar(s[i+1]) && isHexChar(s[i+2])
}

func Encode(input string) string {
	if input == "" {
		return ""
	}

	var buf bytes.Buffer
	for i := 0; i < len(input); {
		c := input[i]

		if c == '%' {
			if isCompletePercentSequence(input, i) {
				buf.WriteByte(c)
				buf.WriteByte(input[i+1])
				buf.WriteByte(input[i+2])
				i += 3
				continue
			}
		}

		if isUnreserved(c) {
			buf.WriteByte(c)
			i++
			continue
		}

		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			encodeByte(&buf, c)
			i++
		} else {
			for j := 0; j < size; j++ {
				encodeByte(&buf, input[i+j])
			}
			i += size
		}
	}

	return buf.String()
}

func encodeByte(buf *bytes.Buffer, c byte) {
	buf.WriteByte('%')
	buf.WriteByte(hexChars[c>>4])
	buf.WriteByte(hexChars[c&0x0F])
}

func Decode(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var buf bytes.Buffer
	for i := 0; i < len(input); {
		c := input[i]

		if c != '%' {
			buf.WriteByte(c)
			i++
			continue
		}

		if i+2 >= len(input) {
			return "", errors.New("incomplete percent-encoded sequence")
		}

		c1, c2 := input[i+1], input[i+2]
		if !isHexChar(c1) || !isHexChar(c2) {
			return "", errors.New("invalid hex character in percent-encoded sequence")
		}

		decodedByte := hexToByte(c1)<<4 | hexToByte(c2)
		buf.WriteByte(decodedByte)
		i += 3
	}

	return buf.String(), nil
}
