package qp

import (
	"bytes"
	"encoding/hex"
	"errors"
)

var (
	ErrIncompleteEscape = errors.New("quoted-printable: incomplete escape sequence")
	ErrInvalidHex       = errors.New("quoted-printable: invalid hex character in escape sequence")
)

func Decode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	var buf bytes.Buffer

	for i := 0; i < len(data); i++ {
		b := data[i]

		if b == eq {
			if i+1 >= len(data) {
				return nil, ErrIncompleteEscape
			}

			next := data[i+1]

			if next == '\r' {
				if i+2 >= len(data) || data[i+2] != '\n' {
					return nil, ErrIncompleteEscape
				}
				i += 2
			} else if next == '\n' {
				i += 1
			} else {
				if i+2 >= len(data) {
					return nil, ErrIncompleteEscape
				}

				hexChars := []byte{next, data[i+2]}
				decoded, err := hex.DecodeString(string(hexChars))
				if err != nil {
					return nil, ErrInvalidHex
				}
				buf.Write(decoded)
				i += 2
			}
		} else {
			buf.WriteByte(b)
		}
	}

	return buf.Bytes(), nil
}

func DecodeString(s string) (string, error) {
	data, err := Decode([]byte(s))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
