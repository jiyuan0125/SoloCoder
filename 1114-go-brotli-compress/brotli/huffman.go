package brotli

import (
	"errors"
	"sort"
)

type huffmanNode struct {
	symbol  int
	count   int
	left    *huffmanNode
	right   *huffmanNode
	depth   int
}

type huffmanCode struct {
	symbol int
	code   uint32
	length int
}

func buildHuffmanTree(symbols []int, maxLen int) ([]huffmanCode, error) {
	if len(symbols) == 0 {
		return nil, errors.New("empty symbols")
	}

	counts := make(map[int]int)
	for _, s := range symbols {
		counts[s]++
	}

	var nodes []*huffmanNode
	for symbol, count := range counts {
		nodes = append(nodes, &huffmanNode{symbol: symbol, count: count})
	}

	for len(nodes) > 1 {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].count < nodes[j].count
		})

		first := nodes[0]
		second := nodes[1]
		nodes = nodes[2:]

		parent := &huffmanNode{
			symbol: -1,
			count:  first.count + second.count,
			left:   first,
			right:  second,
		}
		nodes = append(nodes, parent)
	}

	if len(nodes) == 0 {
		return nil, errors.New("no nodes")
	}

	codes := make([]huffmanCode, 0, len(counts))
	var traverse func(node *huffmanNode, code uint32, length int)
	traverse = func(node *huffmanNode, code uint32, length int) {
		if node == nil {
			return
		}

		if node.symbol >= 0 {
			reversedCode := reverseBits(code, length)
			codes = append(codes, huffmanCode{
				symbol: node.symbol,
				code:   reversedCode,
				length: length,
			})
			return
		}

		if length+1 > maxLen {
			return
		}

		traverse(node.left, code, length+1)
		traverse(node.right, code|(1<<length), length+1)
	}

	traverse(nodes[0], 0, 0)

	limitLengths(codes, maxLen)

	return codes, nil
}

func reverseBits(code uint32, length int) uint32 {
	var result uint32 = 0
	for i := 0; i < length; i++ {
		result <<= 1
		result |= code & 1
		code >>= 1
	}
	return result
}

func limitLengths(codes []huffmanCode, maxLen int) {
	needsAdjustment := false
	for _, c := range codes {
		if c.length > maxLen {
			needsAdjustment = true
			break
		}
	}

	if !needsAdjustment {
		return
	}

	for i := range codes {
		if codes[i].length > maxLen {
			codes[i].length = maxLen
		}
	}

	maxCode := 1 << maxLen
	used := make([]bool, maxCode)
	for i := range codes {
		code := uint32(0)
		for j := 0; j < codes[i].length; j++ {
			for used[code] {
				code++
			}
			codes[i].code = code
			used[code] = true
			code <<= 1
			if j < codes[i].length-1 {
				used[code] = true
			}
		}
	}
}

type huffmanDecoder struct {
	codes    map[uint32]huffmanCode
	maxLen   int
}

func buildHuffmanDecoder(codes []huffmanCode) *huffmanDecoder {
	decoder := &huffmanDecoder{
		codes:  make(map[uint32]huffmanCode),
		maxLen: 0,
	}

	for _, c := range codes {
		if c.length > decoder.maxLen {
			decoder.maxLen = c.length
		}
		decoder.codes[c.code] = c
	}

	return decoder
}

func (d *huffmanDecoder) decode(br *bitReader) (int, error) {
	var code uint32 = 0
	for i := 0; i <= d.maxLen; i++ {
		if i > 0 {
			bit, err := br.readBits(1)
			if err != nil {
				return 0, err
			}
			code |= uint32(bit) << (i - 1)
		}

		if c, ok := d.codes[code]; ok {
			return c.symbol, nil
		}
	}

	return 0, errors.New("invalid huffman code")
}
