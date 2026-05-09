package lz4

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidMagic        = errors.New("lz4: invalid magic number")
	ErrUnsupportedVersion  = errors.New("lz4: unsupported version")
	ErrInvalidBlockSize    = errors.New("lz4: invalid block size")
	ErrCorruptedFrame      = errors.New("lz4: corrupted frame")
	ErrTruncatedFrame      = errors.New("lz4: truncated frame")
)

func Compress(src []byte) ([]byte, error) {
	srcLen := len(src)

	dst := make([]byte, 0)

	dst = binary.LittleEndian.AppendUint32(dst, MagicNumber)

	flg := byte((VersionCurrent << 6) | 0x40)
	dst = append(dst, flg)

	bd := byte(0x70)
	dst = append(dst, bd)

	dst = append(dst, checksumDescriptor(flg, bd))

	blockSize := srcLen
	if blockSize > MaxBlockSize {
		blockSize = MaxBlockSize
	}

	compressedBound := compressBound(blockSize)
	compressedBuf := make([]byte, compressedBound)

	pos := 0
	for pos < srcLen {
		end := pos + blockSize
		if end > srcLen {
			end = srcLen
		}
		blockData := src[pos:end]

		compressedLen, err := CompressBlock(blockData, compressedBuf)
		if err != nil {
			return nil, err
		}

		var blockSizeField uint32
		var blockContent []byte

		if compressedLen > 0 && compressedLen < len(blockData) {
			blockSizeField = uint32(compressedLen)
			blockContent = compressedBuf[:compressedLen]
		} else {
			blockSizeField = uint32(len(blockData)) | UncompressedFlag
			blockContent = blockData
		}

		dst = binary.LittleEndian.AppendUint32(dst, blockSizeField)
		dst = append(dst, blockContent...)

		pos = end
	}

	dst = binary.LittleEndian.AppendUint32(dst, 0)

	return dst, nil
}

func Decompress(src []byte) ([]byte, error) {
	srcLen := len(src)
	if srcLen < 4 {
		return nil, ErrTruncatedFrame
	}

	pos := 0

	for pos+4 <= srcLen {
		magic := binary.LittleEndian.Uint32(src[pos:])
		pos += 4

		if magic >= MagicNumberSkippableMin && magic <= MagicNumberSkippableMax {
			if pos+4 > srcLen {
				return nil, ErrTruncatedFrame
			}
			skipLen := int(binary.LittleEndian.Uint32(src[pos:]))
			pos += 4
			if pos+skipLen > srcLen {
				return nil, ErrTruncatedFrame
			}
			pos += skipLen
			continue
		}

		if magic != MagicNumber {
			return nil, ErrInvalidMagic
		}

		if pos+2 > srcLen {
			return nil, ErrTruncatedFrame
		}
		flg := src[pos]
		pos++
		bd := src[pos]
		pos++

		version := (flg >> 6) & 0x03
		if version != VersionCurrent {
			return nil, ErrUnsupportedVersion
		}

		if pos > srcLen {
			return nil, ErrTruncatedFrame
		}
		expectedChecksum := checksumDescriptor(flg, bd)
		if src[pos] != expectedChecksum {
			return nil, ErrCorruptedFrame
		}
		pos++

		result := make([]byte, 0, 65536)
		decompressBuf := make([]byte, 65536)

		for {
			if pos+4 > srcLen {
				return nil, ErrTruncatedFrame
			}
			blockSizeField := binary.LittleEndian.Uint32(src[pos:])
			pos += 4

			blockSize := int(blockSizeField & BlockSizeMask)
			isUncompressed := (blockSizeField & UncompressedFlag) != 0

			if blockSize == 0 {
				return result, nil
			}

			if pos+blockSize > srcLen {
				return nil, ErrTruncatedFrame
			}
			blockData := src[pos : pos+blockSize]
			pos += blockSize

			if isUncompressed {
				result = append(result, blockData...)
			} else {
				for {
					decompressedLen, err := DecompressBlock(blockData, decompressBuf, len(decompressBuf))
					if err == ErrOutputTooSmall {
						decompressBuf = make([]byte, len(decompressBuf)*2)
						continue
					}
					if err != nil {
						return nil, err
					}
					result = append(result, decompressBuf[:decompressedLen]...)
					break
				}
			}
		}
	}

	return []byte{}, nil
}

func checksumDescriptor(flg, bd byte) byte {
	const hashPrime uint32 = 2654435761
	h := uint32(flg)<<8 | uint32(bd)
	h *= hashPrime
	h >>= 8
	return byte(h)
}
