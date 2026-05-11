package brotli

import (
	"errors"
	"sort"
)

var errCorrupted = errors.New("brotli: corrupted input")

type huffmanCode struct {
	value uint32
	len   uint8
}

type huffmanNode struct {
	left   *huffmanNode
	right  *huffmanNode
	symbol int
	freq   int
}

type byFreq []*huffmanNode

func (s byFreq) Len() int           { return len(s) }
func (s byFreq) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byFreq) Less(i, j int) bool { return s[i].freq < s[j].freq }

func buildHuffmanCodes(freq []int, maxCodeLen int) ([]huffmanCode, error) {
	n := len(freq)
	if n == 0 {
		return nil, errCorrupted
	}

	nodes := make([]*huffmanNode, 0, n)
	for i, f := range freq {
		if f > 0 {
			nodes = append(nodes, &huffmanNode{symbol: i, freq: f})
		}
	}

	if len(nodes) == 0 {
		return make([]huffmanCode, n), nil
	}

	if len(nodes) == 1 {
		codes := make([]huffmanCode, n)
		codes[nodes[0].symbol] = huffmanCode{value: 0, len: 1}
		return codes, nil
	}

	sort.Sort(byFreq(nodes))

	i := 0
	j := 0
	pool := make([]huffmanNode, len(nodes)+2)
	poolIdx := 0

	for len(nodes)-i > 1 {
		var left, right *huffmanNode

		if j == 0 || (i < len(nodes) && nodes[i].freq <= pool[j-1].freq) {
			left = nodes[i]
			i++
		} else {
			left = &pool[j-1]
			j--
		}

		if j == 0 || (i < len(nodes) && nodes[i].freq <= pool[j-1].freq) {
			right = nodes[i]
			i++
		} else {
			right = &pool[j-1]
			j--
		}

		pool[poolIdx] = huffmanNode{
			left:  left,
			right: right,
			freq:  left.freq + right.freq,
		}
		j++
		poolIdx++
	}

	root := &pool[j-1]

	codes := make([]huffmanCode, n)
	assignCodes(root, 0, 0, codes, maxCodeLen)

	return codes, nil
}

func assignCodes(node *huffmanNode, code uint32, len int, codes []huffmanCode, maxCodeLen int) {
	if node == nil {
		return
	}

	if node.left == nil && node.right == nil {
		if len > maxCodeLen {
			len = maxCodeLen
		}
		codes[node.symbol] = huffmanCode{value: bitReverse(code, uint8(len)), len: uint8(len)}
		return
	}

	assignCodes(node.left, code, len+1, codes, maxCodeLen)
	assignCodes(node.right, code|(1<<len), len+1, codes, maxCodeLen)
}

func bitReverse(code uint32, len uint8) uint32 {
	var result uint32
	for i := uint8(0); i < len; i++ {
		if code&(1<<i) != 0 {
			result |= 1 << (len - 1 - i)
		}
	}
	return result
}

func reverseBits16(x uint16) uint16 {
	x = ((x & 0x5555) << 1) | ((x & 0xAAAA) >> 1)
	x = ((x & 0x3333) << 2) | ((x & 0xCCCC) >> 2)
	x = ((x & 0x0F0F) << 4) | ((x & 0xF0F0) >> 4)
	x = ((x & 0x00FF) << 8) | ((x & 0xFF00) >> 8)
	return x
}

func reverseBits8(x uint8) uint8 {
	x = ((x & 0x55) << 1) | ((x & 0xAA) >> 1)
	x = ((x & 0x33) << 2) | ((x & 0xCC) >> 2)
	x = ((x & 0x0F) << 4) | ((x & 0xF0) >> 4)
	return x
}

type huffmanDecoder struct {
	codes    []huffmanCode
	maxLen   uint8
	lookup   []uint16
	lookupBits uint8
}

func buildHuffmanDecoder(codes []huffmanCode) *huffmanDecoder {
	maxLen := uint8(0)
	for _, c := range codes {
		if c.len > maxLen {
			maxLen = c.len
		}
	}

	lookupBits := uint8(7)
	if maxLen < lookupBits {
		lookupBits = maxLen
	}

	lookupSize := 1 << lookupBits
	lookup := make([]uint16, lookupSize)

	for i := 0; i < lookupSize; i++ {
		lookup[i] = 0xffff
	}

	for symbol, code := range codes {
		if code.len == 0 {
			continue
		}

		if code.len <= lookupBits {
			key := reverseBits16(uint16(code.value))
			key >>= (16 - code.len)

			for i := 0; i < (1 << (lookupBits - code.len)); i++ {
				lookup[key|uint16(i)<<code.len] = uint16(code.len<<8) | uint16(symbol)
			}
		}
	}

	return &huffmanDecoder{
		codes:      codes,
		maxLen:     maxLen,
		lookup:     lookup,
		lookupBits: lookupBits,
	}
}

func (d *huffmanDecoder) decode(br *bitReader) (int, error) {
	bitsNeeded := int(d.lookupBits)
	if br.bitsLeft() < uint64(bitsNeeded) {
		bitsNeeded = int(br.bitsLeft())
	}

	if bitsNeeded == 0 {
		return 0, errCorrupted
	}

	bits, err := br.readBits(bitsNeeded)
	if err != nil {
		return 0, err
	}

	idx := int(bits)
	if idx < len(d.lookup) && d.lookup[idx] != 0xffff {
		symbol := int(d.lookup[idx] & 0xff)
		len := int(d.lookup[idx] >> 8)

		if len < bitsNeeded {
			unusedBits := bitsNeeded - len
			br.bitPos -= uint64(unusedBits)
		}

		return symbol, nil
	}

	for symbol, code := range d.codes {
		if code.len == 0 {
			continue
		}

		if int(code.len) <= bitsNeeded {
			key := uint32(bits) >> (bitsNeeded - int(code.len))
			key &= (1 << code.len) - 1
			key = bitReverse(key, code.len)

			if key == code.value {
				if int(code.len) < bitsNeeded {
					unusedBits := bitsNeeded - int(code.len)
					br.bitPos -= uint64(unusedBits)
				}
				return symbol, nil
			}
		}
	}

	extraBits := 1
	for ; extraBits+bitsNeeded <= int(d.maxLen); extraBits++ {
		if br.bitPos >= br.totalBits {
			break
		}

		if br.readBit() {
			bits |= 1 << (bitsNeeded + extraBits - 1)
		}

		totalBits := bitsNeeded + extraBits
		for symbol, code := range d.codes {
			if int(code.len) != totalBits {
				continue
			}

			key := bitReverse(uint32(bits), uint8(totalBits))
			if key == code.value {
				return symbol, nil
			}
		}
	}

	return 0, errCorrupted
}
