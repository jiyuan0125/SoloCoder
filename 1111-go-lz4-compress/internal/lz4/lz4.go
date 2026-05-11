package lz4

import (
	"encoding/binary"
	"errors"
)

const (
	magicNumber       = 0x184D2204
	version           = 1
	minMatch          = 4
	blockSize         = 1 << 16
	hashTableSize     = 1 << 16
	hashTableShift    = 12
)

func CompressFrame(data []byte) ([]byte, error) {
	if len(data) == 0 {
		frame := make([]byte, 11)
		binary.LittleEndian.PutUint32(frame[0:4], magicNumber)
		frame[4] = byte(version << 6)
		frame[5] = 0x70
		frame[6] = 0
		binary.LittleEndian.PutUint32(frame[7:11], 0)
		return frame, nil
	}

	frame := make([]byte, 0, len(data)+20)
	frame = append(frame, []byte{
		0x04, 0x22, 0x4D, 0x18,
		0x60,
		0x70,
		0x00,
	}...)

	for i := 0; i < len(data); i += blockSize {
		end := i + blockSize
		if end > len(data) {
			end = len(data)
		}
		block := data[i:end]

		compressed, err := CompressBlock(block)
		if err != nil {
			return nil, err
		}

		if len(compressed) >= len(block) {
			size := uint32(len(block)) | 0x80000000
			sizeBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(sizeBytes, size)
			frame = append(frame, sizeBytes...)
			frame = append(frame, block...)
		} else {
			size := uint32(len(compressed))
			sizeBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(sizeBytes, size)
			frame = append(frame, sizeBytes...)
			frame = append(frame, compressed...)
		}
	}

	frame = append(frame, 0x00, 0x00, 0x00, 0x00)
	return frame, nil
}

func DecompressFrame(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty frame")
	}

	pos := 0
	result := make([]byte, 0, len(data)*2)

	for pos < len(data) {
		if pos+4 > len(data) {
			return nil, errors.New("incomplete frame")
		}

		magic := binary.LittleEndian.Uint32(data[pos : pos+4])
		if magic == magicNumber {
			pos += 4
			if pos+2 > len(data) {
				return nil, errors.New("incomplete frame descriptor")
			}
			flg := data[pos]
			pos++
			pos++

			if flg&(1<<3) != 0 {
				if pos+8 > len(data) {
					return nil, errors.New("incomplete content size")
				}
				pos += 8
			}
			if flg&(1<<0) != 0 {
				if pos+4 > len(data) {
					return nil, errors.New("incomplete dictionary ID")
				}
				pos += 4
			}
			if pos+1 > len(data) {
				return nil, errors.New("incomplete header checksum")
			}
			pos++

			for {
				if pos+4 > len(data) {
					return nil, errors.New("incomplete block size")
				}
				blockSize := binary.LittleEndian.Uint32(data[pos : pos+4])
				pos += 4

				if blockSize == 0 {
					break
				}

				uncompressed := (blockSize & 0x80000000) != 0
				size := blockSize & 0x7FFFFFFF

				if pos+int(size) > len(data) {
					return nil, errors.New("incomplete block")
				}

				blockData := data[pos : pos+int(size)]
				pos += int(size)

				if uncompressed {
					result = append(result, blockData...)
				} else {
					decompressed, err := DecompressBlock(blockData, blockSize)
					if err != nil {
						return nil, err
					}
					result = append(result, decompressed...)
				}

				if flg&(1<<4) != 0 {
					if pos+4 > len(data) {
						return nil, errors.New("incomplete block checksum")
					}
					pos += 4
				}
			}

			if flg&(1<<2) != 0 {
				if pos+4 > len(data) {
					return nil, errors.New("incomplete content checksum")
				}
				pos += 4
			}
		} else if magic >= 0x184D2A50 && magic <= 0x184D2A5F {
			pos += 4
			if pos+4 > len(data) {
				return nil, errors.New("incomplete skippable frame size")
			}
			size := binary.LittleEndian.Uint32(data[pos : pos+4])
			pos += 4
			pos += int(size)
		} else {
			return nil, errors.New("invalid magic number")
		}
	}

	return result, nil
}

func CompressBlock(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	dst := make([]byte, 0, len(src)+len(src)/255+16)
	srcSize := len(src)
	hashTable := make([]int, hashTableSize)
	for i := range hashTable {
		hashTable[i] = -1
	}

	anchor := 0

	for pos := 0; pos < srcSize-minMatch; {
		matchPos := pos
		matchLength := 0

		for i := 0; i < 3; i++ {
			if pos+i >= srcSize {
				break
			}
		}

		if pos+4 <= srcSize {
			hash := int(binary.LittleEndian.Uint32(src[pos:pos+4])>>hashTableShift) & (hashTableSize - 1)
			if hashTable[hash] >= 0 && pos-hashTable[hash] < 65536 {
				matchPos = hashTable[hash]
				matchLength = 0

				for pos+matchLength < srcSize && src[pos+matchLength] == src[matchPos+matchLength] {
					matchLength++
				}
			}
			hashTable[hash] = pos
		}

		if matchLength >= minMatch {
			offset := pos - matchPos
			if offset < 1 {
				pos++
				continue
			}

			literalLength := pos - anchor
			token := byte(min(literalLength, 15)) << 4
			matchToken := byte(min(matchLength-minMatch, 15))
			token |= matchToken

			dst = append(dst, token)

			if literalLength >= 15 {
				remaining := literalLength - 15
				for remaining >= 255 {
					dst = append(dst, 255)
					remaining -= 255
				}
				dst = append(dst, byte(remaining))
			}

			if literalLength > 0 {
				dst = append(dst, src[anchor:anchor+literalLength]...)
			}

			dst = append(dst, byte(offset&0xFF))
			dst = append(dst, byte((offset>>8)&0xFF))

			if matchLength-minMatch >= 15 {
				remaining := (matchLength - minMatch) - 15
				for remaining >= 255 {
					dst = append(dst, 255)
					remaining -= 255
				}
				dst = append(dst, byte(remaining))
			}

			pos += matchLength
			anchor = pos
		} else {
			pos++
		}
	}

	literalLength := srcSize - anchor
	if literalLength > 0 {
		token := byte(min(literalLength, 15)) << 4
		dst = append(dst, token)

		if literalLength >= 15 {
			remaining := literalLength - 15
			for remaining >= 255 {
				dst = append(dst, 255)
				remaining -= 255
			}
			dst = append(dst, byte(remaining))
		}

		dst = append(dst, src[anchor:]...)
	}

	return dst, nil
}

func DecompressBlock(src []byte, originalSize uint32) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	dst := make([]byte, 0, int(originalSize)*2)
	srcPos := 0
	srcEnd := len(src)

	for srcPos < srcEnd {
		if srcPos >= srcEnd {
			return nil, errors.New("incomplete sequence")
		}

		token := src[srcPos]
		srcPos++

		literalLength := int((token >> 4) & 0x0F)
		matchLength := int(token & 0x0F)

		if literalLength == 15 {
			for srcPos < srcEnd && src[srcPos] == 255 {
				literalLength += 255
				srcPos++
			}
			if srcPos < srcEnd {
				literalLength += int(src[srcPos])
				srcPos++
			}
		}

		if literalLength > 0 {
			if srcPos+literalLength > srcEnd {
				return nil, errors.New("incomplete literal")
			}
			dst = append(dst, src[srcPos:srcPos+literalLength]...)
			srcPos += literalLength
		}

		if srcPos >= srcEnd {
			break
		}

		if srcPos+2 > srcEnd {
			return nil, errors.New("incomplete match offset")
		}
		offset := int(binary.LittleEndian.Uint16(src[srcPos : srcPos+2]))
		srcPos += 2

		if offset < 1 {
			return nil, errors.New("invalid offset: offset cannot be 0")
		}

		if matchLength == 15 {
			for srcPos < srcEnd && src[srcPos] == 255 {
				matchLength += 255
				srcPos++
			}
			if srcPos < srcEnd {
				matchLength += int(src[srcPos])
				srcPos++
			}
		}
		matchLength += minMatch

		if offset > len(dst) {
			return nil, errors.New("offset exceeds decoded data")
		}

		start := len(dst) - offset
		if start < 0 {
			return nil, errors.New("invalid match start position")
		}

		end := start + matchLength
		if end > len(dst) {
			if start > len(dst) {
				return nil, errors.New("match start exceeds decoded data")
			}
			for i := start; i < end; i++ {
				dst = append(dst, dst[i])
			}
		} else {
			dst = append(dst, dst[start:end]...)
		}
	}

	return dst, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
