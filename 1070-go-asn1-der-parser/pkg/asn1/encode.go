package asn1

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func Encode(node *Node) ([]byte, error) {
	var buf bytes.Buffer

	valueBytes, err := encodeValue(node)
	if err != nil {
		return nil, err
	}

	tagBytes := encodeTag(node.Tag)

	lengthBytes := encodeLength(len(valueBytes))

	buf.Write(tagBytes)
	buf.Write(lengthBytes)
	buf.Write(valueBytes)

	return buf.Bytes(), nil
}

func encodeTag(tag Tag) []byte {
	if tag.Number < 31 {
		firstByte := byte(tag.Class) << 6
		if tag.Constructed {
			firstByte |= 0x20
		}
		firstByte |= byte(tag.Number)
		return []byte{firstByte}
	}

	firstByte := byte(tag.Class) << 6
	if tag.Constructed {
		firstByte |= 0x20
	}
	firstByte |= 0x1F

	tagBytes := []byte{firstByte}

	num := tag.Number
	var buf []byte
	for {
		buf = append(buf, byte(num&0x7F))
		num >>= 7
		if num == 0 {
			break
		}
	}

	for i := len(buf) - 1; i >= 0; i-- {
		b := buf[i]
		if i > 0 {
			b |= 0x80
		}
		tagBytes = append(tagBytes, b)
	}

	return tagBytes
}

func encodeLength(length int) []byte {
	if length < 0 {
		return []byte{0x80}
	}

	if length <= 127 {
		return []byte{byte(length)}
	}

	var buf []byte
	for length > 0 {
		buf = append([]byte{byte(length & 0xFF)}, buf...)
		length >>= 8
	}

	firstByte := byte(0x80 | len(buf))
	return append([]byte{firstByte}, buf...)
}

func encodeValue(node *Node) ([]byte, error) {
	if node.Tag.Constructed {
		var buf bytes.Buffer

		children := node.Value.Children

		if isSet(node.Tag) {
			children = sortSetElements(children)
		}

		for _, child := range children {
			encoded, err := Encode(child)
			if err != nil {
				return nil, err
			}
			buf.Write(encoded)
		}

		return buf.Bytes(), nil
	}

	if isBitString(node.Tag) {
		if len(node.Value.Bytes) == 0 {
			return nil, errors.New("BIT STRING must have at least unused bits byte")
		}
	}

	return node.Value.Bytes, nil
}

func isSet(tag Tag) bool {
	return tag.Class == ClassUniversal && tag.Number == TagNumberSET
}

func sortSetElements(elements []*Node) []*Node {
	sorted := make([]*Node, len(elements))
	copy(sorted, elements)

	sort.Slice(sorted, func(i, j int) bool {
		a := sorted[i].Tag
		b := sorted[j].Tag

		if a.Class != b.Class {
			return a.Class < b.Class
		}

		return a.Number < b.Number
	})

	return sorted
}

func NewNode(tag Tag, valueBytes []byte) *Node {
	return &Node{
		Tag:    tag,
		Length: len(valueBytes),
		Value: Value{
			Bytes: valueBytes,
		},
	}
}

func NewConstructedNode(tag Tag, children []*Node) *Node {
	node := &Node{
		Tag:    tag,
		Value:  Value{Children: children},
	}

	valueBytes, _ := encodeValue(node)
	node.Length = len(valueBytes)

	return node
}

func IntegerTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberINTEGER,
	}
}

func BitStringTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberBITSTRING,
	}
}

func OctetStringTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberOCTETSTRING,
	}
}

func NullTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberNULL,
	}
}

func OIDTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberOID,
	}
}

func UTF8StringTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberUTF8String,
	}
}

func SequenceTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: true,
		Number:      TagNumberSEQUENCE,
	}
}

func SetTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: true,
		Number:      TagNumberSET,
	}
}

func PrintableStringTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberPrintableString,
	}
}

func UTCTimeTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberUTCTime,
	}
}

func GeneralizedTimeTag() Tag {
	return Tag{
		Class:       ClassUniversal,
		Constructed: false,
		Number:      TagNumberGeneralizedTime,
	}
}

func NewInteger(value []byte) *Node {
	return NewNode(IntegerTag(), value)
}

func NewBitString(unusedBits byte, data []byte) *Node {
	bytes := append([]byte{unusedBits}, data...)
	return NewNode(BitStringTag(), bytes)
}

func NewOctetString(data []byte) *Node {
	return NewNode(OctetStringTag(), data)
}

func NewNull() *Node {
	return NewNode(NullTag(), []byte{})
}

func NewOID(oid string) (*Node, error) {
	encoded, err := encodeOIDString(oid)
	if err != nil {
		return nil, err
	}
	return NewNode(OIDTag(), encoded), nil
}

func NewUTF8String(value string) *Node {
	return NewNode(UTF8StringTag(), []byte(value))
}

func NewSequence(children []*Node) *Node {
	return NewConstructedNode(SequenceTag(), children)
}

func NewSet(children []*Node) *Node {
	return NewConstructedNode(SetTag(), children)
}

func NewPrintableString(value string) *Node {
	return NewNode(PrintableStringTag(), []byte(value))
}

func encodeOIDString(oid string) ([]byte, error) {
	parts := strings.Split(oid, ".")
	if len(parts) < 2 {
		return nil, errors.New("OID must have at least two components")
	}

	comps := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid OID component: %s", p)
		}
		if n < 0 {
			return nil, fmt.Errorf("OID component cannot be negative: %s", p)
		}
		comps[i] = n
	}

	if comps[0] > 2 {
		return nil, errors.New("first OID component must be 0, 1, or 2")
	}

	if comps[0] < 2 && comps[1] > 39 {
		return nil, errors.New("second OID component must be <= 39 when first is 0 or 1")
	}

	firstVal := comps[0]*40 + comps[1]

	var buf bytes.Buffer

	buf.Write(encodeBase128(firstVal))

	for i := 2; i < len(comps); i++ {
		buf.Write(encodeBase128(comps[i]))
	}

	return buf.Bytes(), nil
}

func encodeBase128(val int) []byte {
	if val == 0 {
		return []byte{0x00}
	}

	var buf []byte
	for val > 0 {
		buf = append([]byte{byte(val & 0x7F)}, buf...)
		val >>= 7
	}

	for i := 0; i < len(buf)-1; i++ {
		buf[i] |= 0x80
	}

	return buf
}
