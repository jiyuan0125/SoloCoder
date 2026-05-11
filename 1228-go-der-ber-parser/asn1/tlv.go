package asn1

import (
	"errors"
)

type TagClass uint8

const (
	TagClassUniversal       TagClass = 0
	TagClassApplication     TagClass = 1
	TagClassContextSpecific TagClass = 2
	TagClassPrivate         TagClass = 3
)

type TagNumber uint32

const (
	TagBoolean           TagNumber = 1
	TagInteger           TagNumber = 2
	TagBitString         TagNumber = 3
	TagOctetString       TagNumber = 4
	TagNull              TagNumber = 5
	TagObjectIdentifier  TagNumber = 6
	TagSequence          TagNumber = 16
	TagSet               TagNumber = 17
	TagPrintableString   TagNumber = 19
	TagT61String         TagNumber = 20
	TagIA5String         TagNumber = 22
	TagUTCTime           TagNumber = 23
	TagGeneralizedTime   TagNumber = 24
	TagUTF8String        TagNumber = 12
)

type TLV struct {
	Class      TagClass
	Number     TagNumber
	Constructed bool
	Value      []byte
	Children   []*TLV
}

func NewTLV(class TagClass, number TagNumber, constructed bool) *TLV {
	return &TLV{
		Class:      class,
		Number:     number,
		Constructed: constructed,
	}
}

func (t *TLV) AddChild(child *TLV) {
	if !t.Constructed {
		panic("cannot add child to primitive TLV")
	}
	t.Children = append(t.Children, child)
}

func (t *TLV) TagBytes() []byte {
	var bytes []byte
	firstByte := byte(t.Class) << 6
	if t.Constructed {
		firstByte |= 0x20
	}

	if t.Number < 31 {
		firstByte |= byte(t.Number)
		bytes = []byte{firstByte}
	} else {
		firstByte |= 0x1F
		bytes = []byte{firstByte}
		n := uint32(t.Number)
		var encoded []byte
		for n > 0 {
			encoded = append(encoded, byte(n&0x7F))
			n >>= 7
		}
		if len(encoded) == 0 {
			encoded = []byte{0}
		}
		for i := len(encoded) - 1; i >= 0; i-- {
			if i > 0 {
				bytes = append(bytes, encoded[i]|0x80)
			} else {
				bytes = append(bytes, encoded[i])
			}
		}
	}
	return bytes
}

func ParseTag(data []byte) (tlv *TLV, tagLen int, err error) {
	if len(data) < 1 {
		return nil, 0, errors.New("insufficient data for tag")
	}

	firstByte := data[0]
	class := TagClass((firstByte & 0xC0) >> 6)
	constructed := (firstByte & 0x20) != 0
	tagNumber := TagNumber(firstByte & 0x1F)

	tagLen = 1

	if tagNumber == 31 {
		var n uint32
		for {
			if tagLen >= len(data) {
				return nil, 0, errors.New("insufficient data for tag")
			}
			b := data[tagLen]
			tagLen++
			n = (n << 7) | uint32(b&0x7F)
			if (b & 0x80) == 0 {
				break
			}
			if tagLen > 5 {
				return nil, 0, errors.New("tag number too large")
			}
		}
		tagNumber = TagNumber(n)
	}

	return &TLV{
		Class:       class,
		Number:      tagNumber,
		Constructed: constructed,
	}, tagLen, nil
}

func IsEndOfContents(tlv *TLV) bool {
	return tlv.Class == TagClassUniversal &&
		tlv.Number == 0 &&
		!tlv.Constructed
}
