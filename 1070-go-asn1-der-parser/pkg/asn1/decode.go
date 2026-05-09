package asn1

import (
	"errors"
	"fmt"
	"strings"
)

func Decode(data []byte) (*Node, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	pos := 0

	tag, tagLen, err := decodeTag(data)
	if err != nil {
		return nil, 0, err
	}
	pos += tagLen

	length, lengthLen, err := decodeLength(data[pos:])
	if err != nil {
		return nil, 0, err
	}
	pos += lengthLen

	node := &Node{
		Tag:    tag,
		Length: length,
	}

	if length < 0 {
		return nil, 0, errors.New("indefinite length not supported in DER")
	}

	if pos+length > len(data) {
		return nil, 0, fmt.Errorf("data too short: need %d bytes, have %d", pos+length, len(data))
	}

	valueBytes := data[pos : pos+length]

	if tag.Constructed {
		children, err := decodeChildren(valueBytes)
		if err != nil {
			return nil, 0, err
		}
		node.Value.Children = children
	} else {
		if isBitString(tag) {
			if length == 0 {
				return nil, 0, errors.New("BIT STRING with zero length is invalid")
			}
		}
		node.Value.Bytes = valueBytes
	}

	return node, pos + length, nil
}

func decodeTag(data []byte) (Tag, int, error) {
	if len(data) == 0 {
		return Tag{}, 0, errors.New("no tag byte")
	}

	firstByte := data[0]
	tag := Tag{
		Class:       TagClass((firstByte & 0xC0) >> 6),
		Constructed: (firstByte & 0x20) != 0,
	}

	tagNumber := int(firstByte & 0x1F)
	pos := 1

	if tagNumber == 0x1F {
		if len(data) < 2 {
			return Tag{}, 0, errors.New("incomplete multi-byte tag")
		}

		tagNumber = 0
		for {
			if pos >= len(data) {
				return Tag{}, 0, errors.New("incomplete multi-byte tag")
			}
			b := data[pos]
			pos++
			tagNumber = (tagNumber << 7) | int(b&0x7F)
			if (b & 0x80) == 0 {
				break
			}
		}
	}

	tag.Number = tagNumber
	return tag, pos, nil
}

func decodeLength(data []byte) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, errors.New("no length byte")
	}

	firstByte := data[0]
	if firstByte == 0x80 {
		return 0, 0, errors.New("indefinite length not supported")
	}

	if (firstByte & 0x80) == 0 {
		return int(firstByte), 1, nil
	}

	numBytes := int(firstByte & 0x7F)
	if numBytes == 0 || numBytes > 4 {
		return 0, 0, fmt.Errorf("invalid length encoding: %d bytes", numBytes)
	}

	if len(data) < 1+numBytes {
		return 0, 0, errors.New("incomplete length")
	}

	length := 0
	for i := 0; i < numBytes; i++ {
		length = (length << 8) | int(data[1+i])
	}

	if length <= 127 {
		return 0, 0, errors.New("DER requires short form for length <= 127")
	}

	return length, 1 + numBytes, nil
}

func decodeChildren(data []byte) ([]*Node, error) {
	children := []*Node{}
	pos := 0

	for pos < len(data) {
		node, n, err := Decode(data[pos:])
		if err != nil {
			return nil, err
		}
		children = append(children, node)
		pos += n
	}

	return children, nil
}

func isBitString(tag Tag) bool {
	return tag.Class == ClassUniversal && tag.Number == TagNumberBITSTRING && !tag.Constructed
}

func decodeOID(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty OID")
	}

	components := []int{}

	first := data[0]
	firstComp := int(first) / 40
	secondComp := int(first) % 40

	if firstComp >= 2 {
		if firstComp == 2 {
			var val int
			var valBytes []byte
			for i := 0; i < len(data); i++ {
				b := data[i]
				valBytes = append(valBytes, b)
				if (b & 0x80) == 0 {
					break
				}
			}
			if len(valBytes) == 1 && valBytes[0] >= 80 {
				firstComp = 2
				val = 0
				for _, b := range valBytes {
					val = (val << 7) | int(b&0x7F)
				}
				firstComp = val / 40
				secondComp = val % 40
				if firstComp > 2 {
					return "", errors.New("invalid first OID component")
				}
			}
		}
	}

	if firstComp > 2 {
		return "", errors.New("invalid first OID component")
	}

	components = append(components, firstComp, secondComp)

	pos := 1

	for pos < len(data) {
		val := 0
		for {
			if pos >= len(data) {
				return "", errors.New("incomplete OID component")
			}
			b := data[pos]
			pos++
			val = (val << 7) | int(b&0x7F)
			if (b & 0x80) == 0 {
				break
			}
		}
		components = append(components, val)
	}

	strs := make([]string, len(components))
	for i, c := range components {
		strs[i] = fmt.Sprintf("%d", c)
	}

	return strings.Join(strs, "."), nil
}
