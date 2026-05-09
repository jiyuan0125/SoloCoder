package lz4

import (
	"encoding/binary"
	"errors"
	"math"
)

var (
	ErrInvalidOffset     = errors.New("lz4: invalid offset (must be >= 1)")
	ErrOffsetOutOfBounds = errors.New("lz4: offset out of bounds")
	ErrOutputTooSmall    = errors.New("lz4: output buffer too small")
)

func compressBound(srcLen int) int {
	if srcLen > MaxBlockSize {
		return srcLen + (srcLen/255) + 16
	}
	return srcLen + (srcLen/255) + 16
}

func writeVarLen(dst []byte, value int) ([]byte, int) {
	written := 0
	for value >= 255 {
		if len(dst) <= written {
			return dst, written
		}
		dst[written] = 255
		written++
		value -= 255
	}
	if len(dst) <= written {
		return dst, written
	}
	dst[written] = byte(value)
	written++
	return dst, written
}

func readVarLen(src []byte) (int, int, error) {
	value := 0
	read := 0
	for read < len(src) {
		b := src[read]
		read++
		if b < 255 {
			value += int(b)
			return value, read, nil
		}
		value += 255
		if value > math.MaxInt32 {
			return 0, 0, errors.New("lz4: variable length too large")
		}
	}
	return 0, 0, errors.New("lz4: truncated variable length")
}

func findMatch(src []byte, pos int, srcEnd int, hashTable []int) (matchPos int, matchLen int) {
	if pos+MinMatch > srcEnd {
		return 0, 0
	}

	h := int(src[pos]) | (int(src[pos+1]) << 8) | (int(src[pos+2]) << 16) | (int(src[pos+3]) << 24)
	h = (h * 2654435761) >> 16
	h = h & 0xFFFF

	matchPos = hashTable[h]
	hashTable[h] = pos

	if matchPos == 0 || pos-matchPos > MaxOffset {
		return 0, 0
	}

	if src[matchPos] != src[pos] || src[matchPos+1] != src[pos+1] ||
		src[matchPos+2] != src[pos+2] || src[matchPos+3] != src[pos+3] {
		return 0, 0
	}

	matchLen = MinMatch
	for pos+matchLen < srcEnd && matchPos+matchLen < pos && src[matchPos+matchLen] == src[pos+matchLen] {
		matchLen++
	}

	return matchPos, matchLen
}

func CompressBlock(src, dst []byte) (int, error) {
	srcLen := len(src)
	if srcLen == 0 {
		return 0, nil
	}

	dstLen := len(dst)
	if dstLen < 1 {
		return 0, ErrOutputTooSmall
	}

	hashTable := make([]int, 65536)

	pos := 0
	dstPos := 0
	literalStart := 0

	for pos+MinMatch < srcLen {
		matchPos, matchLen := findMatch(src, pos, srcLen, hashTable)

		if matchLen == 0 {
			pos++
			continue
		}

		literalLen := pos - literalStart

		tokenLiteral := literalLen
		if tokenLiteral > 15 {
			tokenLiteral = 15
		}

		matchLenCode := matchLen - MinMatch
		if matchLenCode > 15 {
			matchLenCode = 15
		}

		token := byte((tokenLiteral << TokenLiteralShift) | matchLenCode)

		if dstPos+1 > dstLen {
			return 0, ErrOutputTooSmall
		}
		dst[dstPos] = token
		dstPos++

		if literalLen >= 15 {
			remaining := literalLen - 15
			dst, written := writeVarLen(dst[dstPos:], remaining)
			if written < 0 || dstPos+written > dstLen {
				return 0, ErrOutputTooSmall
			}
			dstPos += written
		}

		if literalLen > 0 {
			if dstPos+literalLen > dstLen {
				return 0, ErrOutputTooSmall
			}
			copy(dst[dstPos:], src[literalStart:literalStart+literalLen])
			dstPos += literalLen
		}

		offset := pos - matchPos
		if offset < 1 || offset > MaxOffset {
			return 0, ErrInvalidOffset
		}

		if dstPos+2 > dstLen {
			return 0, ErrOutputTooSmall
		}
		binary.LittleEndian.PutUint16(dst[dstPos:], uint16(offset))
		dstPos += 2

		if matchLen >= MinMatch+15 {
			remaining := matchLen - (MinMatch + 15)
			dst, written := writeVarLen(dst[dstPos:], remaining)
			if written < 0 || dstPos+written > dstLen {
				return 0, ErrOutputTooSmall
			}
			dstPos += written
		}

		pos += matchLen
		literalStart = pos
	}

	literalLen := srcLen - literalStart
	if literalLen > 0 {
		tokenLiteral := literalLen
		if tokenLiteral > 15 {
			tokenLiteral = 15
		}

		token := byte(tokenLiteral << TokenLiteralShift)

		if dstPos+1 > dstLen {
			return 0, ErrOutputTooSmall
		}
		dst[dstPos] = token
		dstPos++

		if literalLen >= 15 {
			remaining := literalLen - 15
			dst, written := writeVarLen(dst[dstPos:], remaining)
			if written < 0 || dstPos+written > dstLen {
				return 0, ErrOutputTooSmall
			}
			dstPos += written
		}

		if dstPos+literalLen > dstLen {
			return 0, ErrOutputTooSmall
		}
		copy(dst[dstPos:], src[literalStart:])
		dstPos += literalLen
	}

	return dstPos, nil
}

func DecompressBlock(src, dst []byte, dstCapacity int) (int, error) {
	srcLen := len(src)
	dstLen := 0

	if srcLen == 0 {
		return 0, nil
	}

	pos := 0

	for pos < srcLen {
		if pos >= srcLen {
			break
		}

		token := src[pos]
		pos++

		literalLen := int((token & TokenLiteralMask) >> TokenLiteralShift)
		matchLenCode := int(token & TokenMatchMask)

		if literalLen == 15 {
			additional, read, err := readVarLen(src[pos:])
			if err != nil {
				return 0, err
			}
			pos += read
			literalLen += additional
		}

		if literalLen > 0 {
			if pos+literalLen > srcLen {
				return 0, errors.New("lz4: literal extends beyond source")
			}
			if dstLen+literalLen > dstCapacity {
				return 0, ErrOutputTooSmall
			}
			copy(dst[dstLen:], src[pos:pos+literalLen])
			dstLen += literalLen
			pos += literalLen
		}

		if pos >= srcLen {
			break
		}

		if pos+2 > srcLen {
			return 0, errors.New("lz4: truncated offset")
		}
		offset := int(binary.LittleEndian.Uint16(src[pos:]))
		pos += 2

		if offset < 1 {
			return 0, ErrInvalidOffset
		}

		matchLen := matchLenCode + MinMatch
		if matchLenCode == 15 {
			additional, read, err := readVarLen(src[pos:])
			if err != nil {
				return 0, err
			}
			pos += read
			matchLen += additional
		}

		matchStart := dstLen - offset
		if matchStart < 0 {
			return 0, ErrOffsetOutOfBounds
		}

		if dstLen+matchLen > dstCapacity {
			return 0, ErrOutputTooSmall
		}

		for i := 0; i < matchLen; i++ {
			dst[dstLen+i] = dst[matchStart+i]
		}
		dstLen += matchLen
	}

	return dstLen, nil
}
