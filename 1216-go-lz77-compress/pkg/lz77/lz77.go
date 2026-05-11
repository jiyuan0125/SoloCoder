package lz77

import (
	"bytes"
	"encoding/binary"
)

const (
	WindowSize     = 32 * 1024
	LookaheadSize  = 32 * 1024
	MinMatchLength = 3
	MaxMatchLength = 255
)

const (
	TagLiteral byte = 0x00
	TagTuple   byte = 0x01
)

func Compress(input []byte) []byte {
	if len(input) == 0 {
		return []byte{}
	}

	var output bytes.Buffer
	position := 0
	n := len(input)

	for position < n {
		lookaheadEnd := position + LookaheadSize
		if lookaheadEnd > n {
			lookaheadEnd = n
		}

		windowStart := position - WindowSize
		if windowStart < 0 {
			windowStart = 0
		}

		bestDistance := 0
		bestLength := 0

		if position > 0 && lookaheadEnd-position >= MinMatchLength {
			lookahead := input[position:lookaheadEnd]

			for i := position - 1; i >= windowStart; i-- {
				maxPossible := min(n-position, MaxMatchLength)

				matchLength := 0
				for j := 0; j < maxPossible; j++ {
					if input[i+j] == lookahead[j] {
						matchLength++
					} else {
						break
					}
				}

				if matchLength >= MinMatchLength && matchLength > bestLength {
					bestDistance = position - i
					bestLength = matchLength

					if matchLength == maxPossible {
						break
					}
				}
			}
		}

		if bestLength >= MinMatchLength {
			nextCharPos := position + bestLength

			if nextCharPos >= n {
				maxMatchWithNextChar := (n - 1) - position
				if maxMatchWithNextChar >= MinMatchLength {
					bestLength = maxMatchWithNextChar
					nextCharPos = position + bestLength
				} else {
					encodeLiteral(&output, input[position])
					position++
					continue
				}
			}

			nextChar := input[nextCharPos]
			encodeTuple(&output, bestDistance, bestLength, nextChar)
			position = nextCharPos + 1
		} else {
			encodeLiteral(&output, input[position])
			position++
		}
	}

	return output.Bytes()
}

func Decompress(encoded []byte) []byte {
	if len(encoded) == 0 {
		return []byte{}
	}

	var output bytes.Buffer
	position := 0
	n := len(encoded)

	for position < n {
		if position+1 > n {
			break
		}

		tag := encoded[position]
		position++

		if tag == TagLiteral {
			if position >= n {
				break
			}
			output.WriteByte(encoded[position])
			position++
		} else if tag == TagTuple {
			if position+4 > n {
				break
			}

			distance := binary.LittleEndian.Uint16(encoded[position : position+2])
			position += 2

			length := int(encoded[position])
			position++

			nextChar := encoded[position]
			position++

			currentLen := output.Len()
			windowStart := currentLen - int(distance)
			if windowStart < 0 {
				windowStart = 0
			}

			windowBytes := output.Bytes()
			for i := 0; i < length; i++ {
				idx := windowStart + (i % int(distance))
				output.WriteByte(windowBytes[idx])
			}

			output.WriteByte(nextChar)
		} else {
			break
		}
	}

	return output.Bytes()
}

func encodeLiteral(buf *bytes.Buffer, b byte) {
	buf.WriteByte(TagLiteral)
	buf.WriteByte(b)
}

func encodeTuple(buf *bytes.Buffer, distance, length int, nextChar byte) {
	if length < MinMatchLength {
		length = MinMatchLength
	}
	if length > MaxMatchLength {
		length = MaxMatchLength
	}

	buf.WriteByte(TagTuple)
	distBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(distBytes, uint16(distance))
	buf.Write(distBytes)
	buf.WriteByte(byte(length))
	buf.WriteByte(nextChar)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
