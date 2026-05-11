package qp

import (
	"bytes"
	"encoding/hex"
)

const maxLineLength = 76

var (
	crlf = []byte{'\r', '\n'}
	eq   = byte('=')
)

func Encode(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	var line bytes.Buffer
	var pendingSpace bool
	var pendingTab bool
	spaceCount := 0
	tabCount := 0

	flushPending := func() {
		for i := 0; i < spaceCount; i++ {
			writeByte(&buf, &line, ' ', false)
		}
		for i := 0; i < tabCount; i++ {
			writeByte(&buf, &line, '\t', false)
		}
		spaceCount = 0
		tabCount = 0
		pendingSpace = false
		pendingTab = false
	}

	for i := 0; i < len(data); i++ {
		b := data[i]

		if b == '\r' {
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
			flushLineEndSpaces(&buf, &line, spaceCount, tabCount)
			spaceCount = 0
			tabCount = 0
			pendingSpace = false
			pendingTab = false
			buf.Write(crlf)
			line.Reset()
		} else if b == '\n' {
			flushLineEndSpaces(&buf, &line, spaceCount, tabCount)
			spaceCount = 0
			tabCount = 0
			pendingSpace = false
			pendingTab = false
			buf.Write(crlf)
			line.Reset()
		} else if b == ' ' {
			flushPending()
			pendingSpace = true
			spaceCount++
		} else if b == '\t' {
			flushPending()
			pendingTab = true
			tabCount++
		} else {
			flushPending()
			writeByte(&buf, &line, b, false)
		}
	}

	if pendingSpace || pendingTab {
		for i := 0; i < spaceCount; i++ {
			writeByte(&buf, &line, ' ', true)
		}
		for i := 0; i < tabCount; i++ {
			writeByte(&buf, &line, '\t', true)
		}
	}

	if line.Len() > 0 {
		buf.Write(line.Bytes())
	}

	return buf.String(), nil
}

func flushLineEndSpaces(buf, line *bytes.Buffer, spaceCount, tabCount int) {
	for i := 0; i < spaceCount; i++ {
		encoded := encodeByte(' ')
		writeEncoded(buf, line, encoded)
	}
	for i := 0; i < tabCount; i++ {
		encoded := encodeByte('\t')
		writeEncoded(buf, line, encoded)
	}
	if line.Len() > 0 {
		buf.Write(line.Bytes())
	}
}

func writeByte(buf, line *bytes.Buffer, b byte, mustEncode bool) {
	if !mustEncode && canWriteAsIs(b) {
		if line.Len() >= maxLineLength {
			buf.Write(line.Bytes())
			buf.WriteByte(eq)
			buf.Write(crlf)
			line.Reset()
		}
		line.WriteByte(b)
	} else {
		encoded := encodeByte(b)
		writeEncoded(buf, line, encoded)
	}
}

func writeEncoded(buf, line *bytes.Buffer, encoded []byte) {
	if line.Len()+len(encoded) > maxLineLength {
		buf.Write(line.Bytes())
		buf.WriteByte(eq)
		buf.Write(crlf)
		line.Reset()
	}
	line.Write(encoded)
}

func canWriteAsIs(b byte) bool {
	return b >= 33 && b <= 126 && b != 61
}

func encodeByte(b byte) []byte {
	result := make([]byte, 3)
	result[0] = eq
	hex.Encode(result[1:], []byte{b})
	result[1] = toUpper(result[1])
	result[2] = toUpper(result[2])
	return result
}

func toUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

func EncodeString(s string) (string, error) {
	return Encode([]byte(s))
}
