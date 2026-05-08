package huffman

import (
	"bytes"
	"container/heap"
	"errors"
	"fmt"
)

type FreqMap map[rune]int

type Node struct {
	Char     rune
	Freq     int
	Left     *Node
	Right    *Node
	IsLeaf   bool
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Freq == pq[j].Freq {
		if pq[i].IsLeaf && pq[j].IsLeaf {
			return pq[i].Char < pq[j].Char
		}
		if pq[i].IsLeaf {
			return true
		}
		if pq[j].IsLeaf {
			return false
		}
		return true
	}
	return pq[i].Freq < pq[j].Freq
}

func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

func CountFrequencies(text string) FreqMap {
	freq := make(FreqMap)
	for _, ch := range text {
		freq[ch]++
	}
	return freq
}

func BuildHuffmanTree(freq FreqMap) *Node {
	if len(freq) == 0 {
		return nil
	}

	pq := make(PriorityQueue, 0, len(freq))
	for ch, f := range freq {
		pq = append(pq, &Node{
			Char:   ch,
			Freq:   f,
			IsLeaf: true,
		})
	}
	heap.Init(&pq)

	for len(pq) > 1 {
		left := heap.Pop(&pq).(*Node)
		right := heap.Pop(&pq).(*Node)

		parent := &Node{
			Freq:   left.Freq + right.Freq,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}
		heap.Push(&pq, parent)
	}

	return heap.Pop(&pq).(*Node)
}

func BuildCodeTable(root *Node) map[rune]string {
	codeTable := make(map[rune]string)
	if root == nil {
		return codeTable
	}

	var traverse func(*Node, string)
	traverse = func(node *Node, code string) {
		if node.IsLeaf {
			if code == "" {
				code = "0"
			}
			codeTable[node.Char] = code
			return
		}
		if node.Left != nil {
			traverse(node.Left, code+"0")
		}
		if node.Right != nil {
			traverse(node.Right, code+"1")
		}
	}

	traverse(root, "")
	return codeTable
}

func Encode(text string, codeTable map[rune]string) ([]byte, int, error) {
	if text == "" {
		return nil, 0, nil
	}

	if len(codeTable) == 0 {
		return nil, 0, errors.New("empty code table")
	}

	var bitString bytes.Buffer
	for _, ch := range text {
		code, ok := codeTable[ch]
		if !ok {
			return nil, 0, fmt.Errorf("character %q not found in code table", ch)
		}
		bitString.WriteString(code)
	}

	bits := bitString.String()
	numBits := len(bits)
	paddingBits := (8 - (numBits % 8)) % 8

	if paddingBits > 0 {
		padding := make([]byte, paddingBits)
		for i := range padding {
			padding[i] = '0'
		}
		bits += string(padding)
	}

	numBytes := len(bits) / 8
	result := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		byteVal := 0
		for j := 0; j < 8; j++ {
			if bits[i*8+j] == '1' {
				byteVal |= 1 << (7 - j)
			}
		}
		result[i] = byte(byteVal)
	}

	return result, paddingBits, nil
}

func Decode(data []byte, paddingBits int, root *Node) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if root == nil {
		return "", errors.New("empty huffman tree")
	}

	var bitString bytes.Buffer
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			if (b >> i) & 1 == 1 {
				bitString.WriteByte('1')
			} else {
				bitString.WriteByte('0')
			}
		}
	}

	bits := bitString.String()
	totalBits := len(bits)
	effectiveBits := totalBits - paddingBits

	var result bytes.Buffer
	node := root
	index := 0

	for index < effectiveBits {
		if node.IsLeaf {
			result.WriteRune(node.Char)
			node = root
		} else {
			if bits[index] == '0' {
				node = node.Left
			} else {
				node = node.Right
			}
			index++
		}
	}

	if node.IsLeaf {
		result.WriteRune(node.Char)
	}

	return result.String(), nil
}

func SerializeTree(root *Node) []byte {
	if root == nil {
		return nil
	}

	var buf bytes.Buffer

	var serialize func(*Node)
	serialize = func(node *Node) {
		if node == nil {
			return
		}
		if node.IsLeaf {
			buf.WriteByte(1)
			charBytes := []byte(string(node.Char))
			lenBytes := byte(len(charBytes))
			buf.WriteByte(lenBytes)
			buf.Write(charBytes)
		} else {
			buf.WriteByte(0)
			serialize(node.Left)
			serialize(node.Right)
		}
	}

	serialize(root)
	return buf.Bytes()
}

func DeserializeTree(data []byte) (*Node, error) {
	if len(data) == 0 {
		return nil, nil
	}

	index := 0

	var deserialize func() (*Node, error)
	deserialize = func() (*Node, error) {
		if index >= len(data) {
			return nil, errors.New("unexpected end of data")
		}

		tag := data[index]
		index++

		if tag == 1 {
			if index >= len(data) {
				return nil, errors.New("unexpected end of data")
			}
			charLen := int(data[index])
			index++
			if index+charLen > len(data) {
				return nil, errors.New("unexpected end of data")
			}
			charBytes := data[index : index+charLen]
			index += charLen

			runes := []rune(string(charBytes))
			if len(runes) != 1 {
				return nil, errors.New("invalid character data")
			}

			return &Node{
				Char:   runes[0],
				IsLeaf: true,
			}, nil
		} else if tag == 0 {
			left, err := deserialize()
			if err != nil {
				return nil, err
			}
			right, err := deserialize()
			if err != nil {
				return nil, err
			}
			return &Node{
				Left:  left,
				Right: right,
			}, nil
		} else {
			return nil, fmt.Errorf("invalid tag: %d", tag)
		}
	}

	return deserialize()
}

func (f FreqMap) Copy() FreqMap {
	copy := make(FreqMap, len(f))
	for k, v := range f {
		copy[k] = v
	}
	return copy
}
