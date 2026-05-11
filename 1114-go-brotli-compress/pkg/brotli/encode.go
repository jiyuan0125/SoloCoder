package brotli

import (
	"bytes"
)

const (
	formatMarker     = 0xB0
	minMatchLength   = 2
	maxMatchLength   = 65535
)

type lz77Match struct {
	position int
	length   int
	distance int
}

type command struct {
	isLiteral bool
	literal   byte
	insertLen int
	copyLen   int
	distance  int
}

type simpleHeader struct {
	formatVersion uint8
	windowBits    uint8
	quality       uint8
	originalLen   uint32
}

type decoderState struct {
	output     []byte
	outputPos  int
	windowSize int
}

func newDecoderState() *decoderState {
	return &decoderState{
		output:     make([]byte, 0, 4096),
		outputPos:  0,
		windowSize: 1 << 16,
	}
}

func (ds *decoderState) appendByte(b byte) {
	ds.output = append(ds.output, b)
	ds.outputPos++
}

func (ds *decoderState) copyFromDistance(distance, length int) error {
	if distance <= 0 || length <= 0 {
		return errCorrupted
	}

	if ds.outputPos < distance {
		return errCorrupted
	}

	start := ds.outputPos - distance
	for i := 0; i < length; i++ {
		ds.appendByte(ds.output[start+i])
	}

	return nil
}

func writeUvarint(n uint32) []byte {
	buf := make([]byte, 0, 5)
	for n >= 0x80 {
		buf = append(buf, byte(n)|0x80)
		n >>= 7
	}
	buf = append(buf, byte(n))
	return buf
}

func readUvarint(br *bitReader) (uint32, error) {
	var result uint32
	var shift uint

	for {
		if br.bitsLeft() < 8 {
			return 0, errCorrupted
		}
		b, err := br.readBits(8)
		if err != nil {
			return 0, err
		}

		result |= uint32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift >= 35 {
			return 0, errCorrupted
		}
	}

	return result, nil
}

func (h *simpleHeader) encode() []byte {
	buf := make([]byte, 0, 8)
	buf = append(buf, formatMarker)

	info := (h.windowBits & 0x1F)
	info |= (h.quality & 0x0F) << 5
	buf = append(buf, info)

	buf = append(buf, writeUvarint(h.originalLen)...)
	return buf
}

func decodeHeader(br *bitReader) (*simpleHeader, error) {
	if br.bitsLeft() < 8 {
		return nil, errCorrupted
	}
	marker, err := br.readBits(8)
	if err != nil {
		return nil, err
	}

	if marker != formatMarker {
		return nil, errCorrupted
	}

	if br.bitsLeft() < 8 {
		return nil, errCorrupted
	}
	info, err := br.readBits(8)
	if err != nil {
		return nil, err
	}

	windowBits := uint8(info & 0x1F)
	quality := uint8((info >> 5) & 0x0F)

	originalLen, err := readUvarint(br)
	if err != nil {
		return nil, err
	}

	return &simpleHeader{
		formatVersion: 1,
		windowBits:    windowBits,
		quality:       quality,
		originalLen:   originalLen,
	}, nil
}

func findLongestMatch(
	data []byte,
	pos int,
	minDistance int,
	maxDistance int,
	searchDepth int,
	minMatch int,
	useDistanceCache bool,
	recentDistances []int,
) (bestMatch lz77Match) {
	bestMatch.length = 0
	maxLen := len(data) - pos
	if maxLen < minMatch {
		return
	}

	if maxLen > maxMatchLength {
		maxLen = maxMatchLength
	}

	end := pos - minDistance
	if end < 0 {
		return
	}

	start := pos - maxDistance
	if start < 0 {
		start = 0
	}

	if useDistanceCache {
		for _, d := range recentDistances {
			if d <= 0 || d > pos {
				continue
			}

			matchStart := pos - d
			if matchStart < 0 {
				continue
			}

			matchLen := 0
			for matchLen < maxLen && data[matchStart+matchLen] == data[pos+matchLen] {
				matchLen++
			}

			if matchLen >= minMatch && matchLen > bestMatch.length {
				bestMatch.length = matchLen
				bestMatch.distance = d
				bestMatch.position = pos
			}
		}
	}

	searched := 0
	for i := start; i <= end && searched < searchDepth; i++ {
		matchLen := 0
		for matchLen < maxLen && data[i+matchLen] == data[pos+matchLen] {
			matchLen++
		}

		if matchLen > bestMatch.length && matchLen >= minMatch {
			bestMatch.length = matchLen
			bestMatch.position = pos
			bestMatch.distance = pos - i

			if matchLen == maxLen {
				break
			}
		}

		searched++
	}

	return
}

func encodeLZ77(data []byte, quality int) []command {
	windowBits := getWindowBits(quality)
	windowSize := getWindowSize(windowBits)
	searchDepth := getSearchDepth(quality)
	minMatch := getMinMatch(quality)

	useDistanceCache := quality >= 4

	commands := make([]command, 0, len(data)/2)
	pos := 0

	literalRunStart := -1

	recentDistances := make([]int, 4)

	for pos < len(data) {
		minDistance := 1
		maxDistance := windowSize

		if quality <= 2 {
			maxDistance = windowSize / 2
		}

		match := findLongestMatch(data, pos, minDistance, maxDistance, searchDepth, minMatch, useDistanceCache, recentDistances)

		if match.length >= minMatch {
			if useDistanceCache && match.distance > 0 {
				copy(recentDistances[1:], recentDistances)
				recentDistances[0] = match.distance
			}

			if literalRunStart >= 0 {
				for i := literalRunStart; i < pos; i++ {
					commands = append(commands, command{
						isLiteral: true,
						literal:   data[i],
					})
				}
				literalRunStart = -1
			}

			commands = append(commands, command{
				isLiteral: false,
				copyLen:   match.length,
				distance:  match.distance,
			})
			pos += match.length
		} else {
			if literalRunStart < 0 {
				literalRunStart = pos
			}
			pos++
		}
	}

	if literalRunStart >= 0 {
		for i := literalRunStart; i < pos; i++ {
			commands = append(commands, command{
				isLiteral: true,
				literal:   data[i],
			})
		}
	}

	return commands
}

func writeCommand(cmd command, bw *bitWriter) {
	if cmd.isLiteral {
		bw.writeBits(0, 1)
		bw.writeBits(uint64(cmd.literal), 8)
	} else {
		bw.writeBits(1, 1)

		lenVal := uint32(cmd.copyLen - 2)
		if lenVal < 15 {
			bw.writeBits(uint64(lenVal), 4)
		} else {
			bw.writeBits(15, 4)
			bw.writeBits(uint64(lenVal), 16)
		}

		distVal := uint32(cmd.distance - 1)
		if distVal < 15 {
			bw.writeBits(uint64(distVal), 4)
		} else if distVal < 256 {
			bw.writeBits(15, 4)
			bw.writeBits(0, 1)
			bw.writeBits(uint64(distVal), 8)
		} else {
			bw.writeBits(15, 4)
			bw.writeBits(1, 1)
			bw.writeBits(uint64(distVal), 24)
		}
	}
}

func readCommand(br *bitReader) (command, error) {
	if br.bitsLeft() < 1 {
		return command{}, errCorrupted
	}

	isCopy := br.readBit()

	if !isCopy {
		if br.bitsLeft() < 8 {
			return command{}, errCorrupted
		}
		lit, err := br.readBits(8)
		if err != nil {
			return command{}, err
		}
		return command{isLiteral: true, literal: byte(lit)}, nil
	}

	if br.bitsLeft() < 4 {
		return command{}, errCorrupted
	}
	lenCode, err := br.readBits(4)
	if err != nil {
		return command{}, err
	}

	var copyLen int
	if lenCode < 15 {
		copyLen = int(lenCode) + 2
	} else {
		if br.bitsLeft() < 16 {
			return command{}, errCorrupted
		}
		lenExtra, err := br.readBits(16)
		if err != nil {
			return command{}, err
		}
		copyLen = int(lenExtra) + 2
	}

	if br.bitsLeft() < 4 {
		return command{}, errCorrupted
	}
	distCode, err := br.readBits(4)
	if err != nil {
		return command{}, err
	}

	var distance int
	if distCode < 15 {
		distance = int(distCode) + 1
	} else {
		if br.bitsLeft() < 1 {
			return command{}, errCorrupted
		}
		distFlag := br.readBit()
		if !distFlag {
			if br.bitsLeft() < 8 {
				return command{}, errCorrupted
			}
			distExtra, err := br.readBits(8)
			if err != nil {
				return command{}, err
			}
			distance = int(distExtra) + 1
		} else {
			if br.bitsLeft() < 24 {
				return command{}, errCorrupted
			}
			distExtra, err := br.readBits(24)
			if err != nil {
				return command{}, err
			}
			distance = int(distExtra) + 1
		}
	}

	return command{isLiteral: false, copyLen: copyLen, distance: distance}, nil
}

func encodeSimple(data []byte, quality int) []byte {
	windowBits := getWindowBits(quality)
	commands := encodeLZ77(data, quality)

	bw := newBitWriter()

	header := simpleHeader{
		formatVersion: 1,
		windowBits:    uint8(windowBits),
		quality:       uint8(quality),
		originalLen:   uint32(len(data)),
	}

	headerBytes := header.encode()
	for _, b := range headerBytes {
		bw.writeBits(uint64(b), 8)
	}

	for _, cmd := range commands {
		writeCommand(cmd, bw)
	}

	return bw.bytes()
}

func decodeSimple(br *bitReader) ([]byte, error) {
	header, err := decodeHeader(br)
	if err != nil {
		return nil, err
	}

	ds := newDecoderState()
	targetLen := int(header.originalLen)

	for ds.outputPos < targetLen {
		cmd, err := readCommand(br)
		if err != nil {
			return nil, err
		}

		if cmd.isLiteral {
			ds.appendByte(cmd.literal)
		} else {
			if err := ds.copyFromDistance(cmd.distance, cmd.copyLen); err != nil {
				return nil, err
			}
		}
	}

	return ds.output, nil
}

func Encode(input []byte, quality int) ([]byte, error) {
	if quality < MinQuality {
		quality = MinQuality
	}
	if quality > MaxQuality {
		quality = MaxQuality
	}

	if len(input) == 0 {
		return []byte{0x03}, nil
	}

	if len(input) <= 16 {
		output := make([]byte, 0, len(input)+5)
		output = append(output, 0x01)

		nibbles := len(input) - 1
		if nibbles >= 12 {
			output = append(output, byte(0x30|(nibbles-12)))
		} else {
			output = append(output, byte(nibbles<<4))
		}

		output = append(output, input...)

		if len(output)%2 == 1 {
			output = append(output, 0x00)
		}

		return output, nil
	}

	return encodeSimple(input, quality), nil
}

func MustEncode(input []byte, quality int) []byte {
	result, err := Encode(input, quality)
	if err != nil {
		panic(err)
	}
	return result
}

func EncodeString(input string, quality int) ([]byte, error) {
	return Encode([]byte(input), quality)
}

func EncodeBuffer(buf *bytes.Buffer, quality int) ([]byte, error) {
	return Encode(buf.Bytes(), quality)
}

func Decode(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, errCorrupted
	}

	if bytes.Equal(input, []byte{0x03}) {
		return []byte{}, nil
	}

	if len(input) >= 2 && input[0] == 0x01 {
		header := input[1]

		var length int
		if (header & 0x30) == 0x30 {
			length = int(header&0x0f) + 13
		} else {
			length = int(header>>4) + 1
		}

		if 2+length > len(input) {
			return nil, errCorrupted
		}

		return input[2 : 2+length], nil
	}

	if len(input) >= 1 && input[0] == formatMarker {
		br := newBitReader(input)
		return decodeSimple(br)
	}

	return nil, errCorrupted
}

func DecodeToString(input []byte) (string, error) {
	result, err := Decode(input)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func MustDecode(input []byte) []byte {
	result, err := Decode(input)
	if err != nil {
		panic(err)
	}
	return result
}
