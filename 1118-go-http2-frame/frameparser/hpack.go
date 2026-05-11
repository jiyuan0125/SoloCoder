package frameparser

import (
	"bytes"
	"fmt"
)

type HPACKDecoder struct {
	dynamicTable []HeaderField
	maxSize      int
	currentSize  int
}

func NewHPACKDecoder() *HPACKDecoder {
	return &HPACKDecoder{
		dynamicTable: make([]HeaderField, 0),
		maxSize:      DefaultHeaderTableSize,
		currentSize:  0,
	}
}

func (d *HPACKDecoder) SetMaxSize(size int) {
	d.maxSize = size
	d.evict()
}

func (d *HPACKDecoder) Decode(data []byte) ([]HeaderField, error) {
	var headers []HeaderField

	pos := 0
	for pos < len(data) {
		field, consumed, err := d.decodeHeaderField(data[pos:])
		if err != nil {
			return nil, err
		}
		if field != nil {
			headers = append(headers, *field)
		}
		pos += consumed
	}

	return headers, nil
}

func (d *HPACKDecoder) decodeHeaderField(data []byte) (*HeaderField, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("empty data")
	}

	firstByte := data[0]

	if firstByte&0x80 != 0 {
		return d.decodeIndexed(data)
	} else if firstByte&0x40 != 0 {
		return d.decodeLiteralWithIncremental(data)
	} else if firstByte&0x20 != 0 {
		return d.decodeDynamicTableSizeUpdate(data)
	} else {
		return d.decodeLiteralNeverIndexed(data)
	}
}

func (d *HPACKDecoder) decodeIndexed(data []byte) (*HeaderField, int, error) {
	index, consumed, err := decodeInteger(data, 7)
	if err != nil {
		return nil, 0, err
	}

	if index == 0 {
		return nil, 0, fmt.Errorf("invalid index 0")
	}

	field, err := d.getField(index)
	if err != nil {
		return nil, 0, err
	}

	return &field, consumed, nil
}

func (d *HPACKDecoder) decodeLiteralWithIncremental(data []byte) (*HeaderField, int, error) {
	index, consumed, err := decodeInteger(data, 6)
	if err != nil {
		return nil, 0, err
	}

	pos := consumed
	var name string

	if index == 0 {
		var consumedName int
		name, consumedName, err = d.decodeString(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		pos += consumedName
	} else {
		field, err := d.getField(index)
		if err != nil {
			return nil, 0, err
		}
		name = field.Name
	}

	var value string
	var consumedValue int
	value, consumedValue, err = d.decodeString(data[pos:])
	if err != nil {
		return nil, 0, err
	}
	pos += consumedValue

	field := HeaderField{Name: name, Value: value}
	d.addToDynamicTable(field)

	return &field, pos, nil
}

func (d *HPACKDecoder) decodeLiteralNeverIndexed(data []byte) (*HeaderField, int, error) {
	index, consumed, err := decodeInteger(data, 4)
	if err != nil {
		return nil, 0, err
	}

	pos := consumed
	var name string

	if index == 0 {
		var consumedName int
		name, consumedName, err = d.decodeString(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		pos += consumedName
	} else {
		field, err := d.getField(index)
		if err != nil {
			return nil, 0, err
		}
		name = field.Name
	}

	var value string
	var consumedValue int
	value, consumedValue, err = d.decodeString(data[pos:])
	if err != nil {
		return nil, 0, err
	}
	pos += consumedValue

	return &HeaderField{Name: name, Value: value}, pos, nil
}

func (d *HPACKDecoder) decodeDynamicTableSizeUpdate(data []byte) (*HeaderField, int, error) {
	newSize, consumed, err := decodeInteger(data, 5)
	if err != nil {
		return nil, 0, err
	}

	if newSize > d.maxSize {
		return nil, 0, fmt.Errorf("dynamic table size %d exceeds max %d", newSize, d.maxSize)
	}

	d.maxSize = newSize
	d.evict()

	return nil, consumed, nil
}

func (d *HPACKDecoder) decodeString(data []byte) (string, int, error) {
	if len(data) == 0 {
		return "", 0, fmt.Errorf("empty string data")
	}

	huffman := data[0]&0x80 != 0
	length, consumed, err := decodeInteger(data, 7)
	if err != nil {
		return "", 0, err
	}

	pos := consumed
	if len(data) < pos+length {
		return "", 0, fmt.Errorf("incomplete string data")
	}

	stringData := data[pos : pos+length]
	pos += length

	var result string
	if huffman {
		decoded, err := huffmanDecode(stringData)
		if err != nil {
			return "", 0, err
		}
		result = string(decoded)
	} else {
		result = string(stringData)
	}

	return result, pos, nil
}

func (d *HPACKDecoder) getField(index int) (HeaderField, error) {
	if index <= 0 {
		return HeaderField{}, fmt.Errorf("invalid index: %d", index)
	}

	if index <= len(staticTable) {
		entry := staticTable[index-1]
		return HeaderField{Name: entry.name, Value: entry.value}, nil
	}

	dynamicIndex := index - len(staticTable) - 1
	if dynamicIndex < 0 || dynamicIndex >= len(d.dynamicTable) {
		return HeaderField{}, fmt.Errorf("invalid dynamic table index: %d", index)
	}

	return d.dynamicTable[dynamicIndex], nil
}

func (d *HPACKDecoder) addToDynamicTable(field HeaderField) {
	size := len(field.Name) + len(field.Value) + 32
	if size > d.maxSize {
		return
	}

	d.dynamicTable = append([]HeaderField{field}, d.dynamicTable...)
	d.currentSize += size

	d.evict()
}

func (d *HPACKDecoder) evict() {
	for d.currentSize > d.maxSize && len(d.dynamicTable) > 0 {
		last := d.dynamicTable[len(d.dynamicTable)-1]
		d.currentSize -= len(last.Name) + len(last.Value) + 32
		d.dynamicTable = d.dynamicTable[:len(d.dynamicTable)-1]
	}
}

func decodeInteger(data []byte, prefix uint8) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty data for integer decoding")
	}

	mask := (1 << prefix) - 1
	i := int(data[0] & byte(mask))

	if i < mask {
		return i, 1, nil
	}

	pos := 1
	m := 0
	for pos < len(data) {
		b := data[pos]
		pos++
		i += int(b&0x7f) << m
		if b&0x80 == 0 {
			return i, pos, nil
		}
		m += 7

		if m > 21 {
			return 0, 0, fmt.Errorf("integer too large")
		}
	}

	return 0, 0, fmt.Errorf("incomplete integer")
}

type huffmanNode struct {
	value  byte
	hasVal bool
	left   *huffmanNode
	right  *huffmanNode
}

func buildHuffmanTree() *huffmanNode {
	root := &huffmanNode{}

	for b, codeInfo := range huffmanCodeMap {
		node := root
		code := codeInfo.code
		bits := codeInfo.bits

		for i := int(bits) - 1; i >= 0; i-- {
			bit := (code >> uint(i)) & 1

			if bit == 0 {
				if node.left == nil {
					node.left = &huffmanNode{}
				}
				node = node.left
			} else {
				if node.right == nil {
					node.right = &huffmanNode{}
				}
				node = node.right
			}
		}

		node.value = b
		node.hasVal = true
	}

	return root
}

var huffmanTree = buildHuffmanTree()

func huffmanDecode(data []byte) ([]byte, error) {
	var result bytes.Buffer
	node := huffmanTree
	padCount := 0

	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bit := (b >> uint(i)) & 1

			if bit == 0 {
				node = node.left
			} else {
				node = node.right
			}

			if node == nil {
				return nil, fmt.Errorf("invalid huffman code")
			}

			if node.hasVal {
				result.WriteByte(node.value)
				node = huffmanTree
			}
		}
		padCount++
	}

	if node != huffmanTree {
		if node.hasVal {
			result.WriteByte(node.value)
		}
	}

	return result.Bytes(), nil
}

func huffmanEncode(data []byte) []byte {
	var result bytes.Buffer
	var bitBuffer uint32
	bitCount := uint8(0)

	for _, b := range data {
		codeInfo := huffmanCodeMap[b]
		code := codeInfo.code
		bits := codeInfo.bits

		for i := int(bits) - 1; i >= 0; i-- {
			bit := (code >> uint(i)) & 1
			bitBuffer = (bitBuffer << 1) | uint32(bit)
			bitCount++

			if bitCount == 8 {
				result.WriteByte(byte(bitBuffer))
				bitBuffer = 0
				bitCount = 0
			}
		}
	}

	if bitCount > 0 {
		for bitCount < 8 {
			bitBuffer = (bitBuffer << 1) | 1
			bitCount++
		}
		result.WriteByte(byte(bitBuffer))
	}

	return result.Bytes()
}
