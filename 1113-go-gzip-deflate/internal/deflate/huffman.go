package deflate

import (
	"math/bits"
	"sort"
)

const (
	MaxLiteralCodes = 286
	MaxDistanceCodes = 30
	MaxCLCodes = 19
)

type HuffmanCode struct {
	Code uint
	Bits int
}

type Frequency struct {
	Symbol int
	Count  int
}

type huffmanNode struct {
	Symbol   int
	Count    int
	Left     *huffmanNode
	Right    *huffmanNode
	IsLeaf   bool
}

func BuildHuffmanCodes(frequencies []int, maxBits int) []HuffmanCode {
	numSymbols := len(frequencies)
	codes := make([]HuffmanCode, numSymbols)

	nodes := make([]*huffmanNode, 0)
	for i, count := range frequencies {
		if count > 0 {
			nodes = append(nodes, &huffmanNode{
				Symbol: i,
				Count:  count,
				IsLeaf: true,
			})
		}
	}

	if len(nodes) == 0 {
		return codes
	}

	for len(nodes) > 1 {
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Count == nodes[j].Count {
				return nodes[i].Symbol < nodes[j].Symbol
			}
			return nodes[i].Count < nodes[j].Count
		})

		left := nodes[0]
		right := nodes[1]
		nodes = nodes[2:]

		parent := &huffmanNode{
			Count: left.Count + right.Count,
			Left:  left,
			Right: right,
		}
		nodes = append(nodes, parent)
	}

	lengths := make([]int, numSymbols)
	calculateCodeLengths(nodes[0], 0, lengths, maxBits)

	generateCanonicalCodes(lengths, codes)

	return codes
}

func calculateCodeLengths(node *huffmanNode, depth int, lengths []int, maxBits int) {
	if node == nil {
		return
	}

	if node.IsLeaf {
		lengths[node.Symbol] = depth
		return
	}

	nextDepth := depth + 1
	if nextDepth > maxBits {
		nextDepth = maxBits
	}

	calculateCodeLengths(node.Left, nextDepth, lengths, maxBits)
	calculateCodeLengths(node.Right, nextDepth, lengths, maxBits)
}

func generateCanonicalCodes(lengths []int, codes []HuffmanCode) {
	lengthCounts := make([]int, 17)
	for _, l := range lengths {
		if l > 0 {
			lengthCounts[l]++
		}
	}

	nextCode := make([]int, 17)
	code := 0
	for bits := 1; bits <= 15; bits++ {
		code = (code + lengthCounts[bits-1]) << 1
		nextCode[bits] = code
	}

	for i := 0; i < len(lengths); i++ {
		l := lengths[i]
		if l > 0 {
			codes[i] = HuffmanCode{
				Code: uint(reverseBits(nextCode[l], l)),
				Bits: l,
			}
			nextCode[l]++
		}
	}
}

func reverseBits(code int, bits int) int {
	if bits == 0 {
		return 0
	}
	result := 0
	for i := 0; i < bits; i++ {
		result = (result << 1) | (code & 1)
		code >>= 1
	}
	return result
}

func GetFixedLiteralLengthCodes() []HuffmanCode {
	codes := make([]HuffmanCode, 288)

	for i := 0; i < 144; i++ {
		codes[i] = HuffmanCode{
			Code: uint(bits.Reverse16(uint16(0x030+i))) >> 7,
			Bits: 8,
		}
	}
	for i := 144; i < 256; i++ {
		codes[i] = HuffmanCode{
			Code: uint(bits.Reverse16(uint16(0x190+(i-144)))) >> 6,
			Bits: 9,
		}
	}
	for i := 256; i < 280; i++ {
		codes[i] = HuffmanCode{
			Code: uint(bits.Reverse16(uint16(0x000+(i-256)))) >> 7,
			Bits: 7,
		}
	}
	for i := 280; i < 288; i++ {
		codes[i] = HuffmanCode{
			Code: uint(bits.Reverse16(uint16(0x0C0+(i-280)))) >> 5,
			Bits: 5,
		}
	}

	return codes
}

func GetFixedDistanceCodes() []HuffmanCode {
	codes := make([]HuffmanCode, 30)

	for i := 0; i < 30; i++ {
		codes[i] = HuffmanCode{
			Code: uint(bits.Reverse16(uint16(i))) >> 11,
			Bits: 5,
		}
	}

	return codes
}

func LengthCode(length int) (code int, extraBits int, extraValue int) {
	switch {
	case length == 3:
		return 257, 0, 0
	case length == 4:
		return 258, 0, 0
	case length == 5:
		return 259, 0, 0
	case length == 6:
		return 260, 0, 0
	case length == 7:
		return 261, 0, 0
	case length == 8:
		return 262, 0, 0
	case length == 9:
		return 263, 0, 0
	case length == 10:
		return 264, 0, 0
	case length <= 12:
		return 265, 1, length - 11
	case length <= 14:
		return 266, 1, length - 13
	case length <= 16:
		return 267, 1, length - 15
	case length <= 18:
		return 268, 1, length - 17
	case length <= 22:
		return 269, 2, length - 19
	case length <= 26:
		return 270, 2, length - 23
	case length <= 30:
		return 271, 2, length - 27
	case length <= 34:
		return 272, 2, length - 31
	case length <= 42:
		return 273, 3, length - 35
	case length <= 50:
		return 274, 3, length - 43
	case length <= 58:
		return 275, 3, length - 51
	case length <= 66:
		return 276, 3, length - 59
	case length <= 82:
		return 277, 4, length - 67
	case length <= 98:
		return 278, 4, length - 83
	case length <= 114:
		return 279, 4, length - 99
	case length <= 130:
		return 280, 4, length - 115
	case length <= 162:
		return 281, 5, length - 131
	case length <= 194:
		return 282, 5, length - 163
	case length <= 226:
		return 283, 5, length - 195
	case length <= 257:
		return 284, 5, length - 227
	default:
		return 285, 0, 0
	}
}

func DistanceCode(dist int) (code int, extraBits int, extraValue int) {
	switch {
	case dist == 1:
		return 0, 0, 0
	case dist == 2:
		return 1, 0, 0
	case dist == 3:
		return 2, 0, 0
	case dist == 4:
		return 3, 0, 0
	case dist <= 6:
		return 4, 1, dist - 5
	case dist <= 8:
		return 5, 1, dist - 7
	case dist <= 12:
		return 6, 2, dist - 9
	case dist <= 16:
		return 7, 2, dist - 13
	case dist <= 24:
		return 8, 3, dist - 17
	case dist <= 32:
		return 9, 3, dist - 25
	case dist <= 48:
		return 10, 4, dist - 33
	case dist <= 64:
		return 11, 4, dist - 49
	case dist <= 96:
		return 12, 5, dist - 65
	case dist <= 128:
		return 13, 5, dist - 97
	case dist <= 192:
		return 14, 6, dist - 129
	case dist <= 256:
		return 15, 6, dist - 193
	case dist <= 384:
		return 16, 7, dist - 257
	case dist <= 512:
		return 17, 7, dist - 385
	case dist <= 768:
		return 18, 8, dist - 513
	case dist <= 1024:
		return 19, 8, dist - 769
	case dist <= 1536:
		return 20, 9, dist - 1025
	case dist <= 2048:
		return 21, 9, dist - 1537
	case dist <= 3072:
		return 22, 10, dist - 2049
	case dist <= 4096:
		return 23, 10, dist - 3073
	case dist <= 6144:
		return 24, 11, dist - 4097
	case dist <= 8192:
		return 25, 11, dist - 6145
	case dist <= 12288:
		return 26, 12, dist - 8193
	case dist <= 16384:
		return 27, 12, dist - 12289
	case dist <= 24576:
		return 28, 13, dist - 16385
	default:
		return 29, 13, dist - 24577
	}
}

var CLLengthOrder = []int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}
