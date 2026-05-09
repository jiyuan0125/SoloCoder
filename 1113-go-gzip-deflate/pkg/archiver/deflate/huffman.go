package deflate

import (
	"sort"
)

type HuffmanNode struct {
	Symbol   int
	Count    int
	Left     *HuffmanNode
	Right    *HuffmanNode
	Depth    int
}

type HuffmanCode struct {
	Code   uint32
	Length int
}

func BuildHuffmanTree(freq map[int]int, maxBits int) []HuffmanCode {
	if len(freq) == 0 {
		return nil
	}

	nodes := make([]*HuffmanNode, 0, len(freq))
	for sym, count := range freq {
		if count > 0 {
			nodes = append(nodes, &HuffmanNode{Symbol: sym, Count: count})
		}
	}

	if len(nodes) == 0 {
		return nil
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

		parent := &HuffmanNode{
			Symbol: -1,
			Count:  left.Count + right.Count,
			Left:   left,
			Right:  right,
		}
		nodes = append(nodes, parent)
	}

	root := nodes[0]
	maxSymbol := 0
	for sym := range freq {
		if sym > maxSymbol {
			maxSymbol = sym
		}
	}

	lengths := make([]int, maxSymbol+1)
	assignLengths(root, 0, lengths)

	canonical := make([]HuffmanCode, maxSymbol+1)
	generateCanonical(lengths, canonical, maxBits)

	return canonical
}

func assignLengths(node *HuffmanNode, depth int, lengths []int) {
	if node == nil {
		return
	}
	if node.Symbol >= 0 {
		lengths[node.Symbol] = depth
		return
	}
	assignLengths(node.Left, depth+1, lengths)
	assignLengths(node.Right, depth+1, lengths)
}

func generateCanonical(lengths []int, codes []HuffmanCode, maxBits int) {
	type symLen struct {
		sym   int
		length int
	}

	var list []symLen
	for sym, length := range lengths {
		if length > 0 {
			list = append(list, symLen{sym, length})
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].length == list[j].length {
			return list[i].sym < list[j].sym
		}
		return list[i].length < list[j].length
	})

	var code uint32
	lastLen := 0
	for _, item := range list {
		if item.length > lastLen {
			code <<= uint(item.length - lastLen)
			lastLen = item.length
		}
		codes[item.sym] = HuffmanCode{
			Code:   reverseBits(code, item.length),
			Length: item.length,
		}
		code++
	}
}

func reverseBits(code uint32, n int) uint32 {
	var result uint32
	for i := 0; i < n; i++ {
		result |= ((code >> uint(i)) & 1) << uint(n-1-i)
	}
	return result
}

var fixedLiteralCodes = buildFixedLiteralCodes()
var fixedDistanceCodes = buildFixedDistanceCodes()

func buildFixedLiteralCodes() []HuffmanCode {
	codes := make([]HuffmanCode, 288)
	var code uint16
	for i := 0; i < 144; i++ {
		codes[i] = HuffmanCode{Code: uint32(code), Length: 8}
		code++
	}
	code = 0b110010000
	for i := 144; i < 256; i++ {
		codes[i] = HuffmanCode{Code: uint32(code), Length: 9}
		code++
	}
	code = 0b0000000
	for i := 256; i < 280; i++ {
		codes[i] = HuffmanCode{Code: uint32(code), Length: 7}
		code++
	}
	code = 0b11000000
	for i := 280; i < 288; i++ {
		codes[i] = HuffmanCode{Code: uint32(code), Length: 8}
		code++
	}
	return codes
}

func buildFixedDistanceCodes() []HuffmanCode {
	codes := make([]HuffmanCode, 32)
	var code uint8
	for i := 0; i < 32; i++ {
		codes[i] = HuffmanCode{Code: uint32(code), Length: 5}
		code++
	}
	return codes
}

func GetFixedLiteralCode(symbol int) HuffmanCode {
	if symbol >= 0 && symbol < len(fixedLiteralCodes) {
		return fixedLiteralCodes[symbol]
	}
	return HuffmanCode{}
}

func GetFixedDistanceCode(symbol int) HuffmanCode {
	if symbol >= 0 && symbol < len(fixedDistanceCodes) {
		return fixedDistanceCodes[symbol]
	}
	return HuffmanCode{}
}
