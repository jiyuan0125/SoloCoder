package stun

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
)

const (
	HeaderSize       = 20
	MagicCookie      = 0x2112A442
	TransactionIDLen = 12
	MaxMessageSize   = 65536
)

const (
	BindingRequest      uint16 = 0x0001
	BindingResponse     uint16 = 0x0101
	BindingErrorResponse uint16 = 0x0111
)

const (
	AttrMappedAddress     uint16 = 0x0001
	AttrUsername          uint16 = 0x0006
	AttrMessageIntegrity  uint16 = 0x0008
	AttrErrorCode         uint16 = 0x0009
	AttrUnknownAttributes uint16 = 0x000A
	AttrRealm             uint16 = 0x0014
	AttrNonce             uint16 = 0x0015
	AttrXORMappedAddress  uint16 = 0x0020
	AttrSoftware          uint16 = 0x8022
	AttrFingerprint       uint16 = 0x8028
)

const (
	FamilyIPv4 byte = 0x01
	FamilyIPv6 byte = 0x02
)

type Message struct {
	Type          uint16
	Length        uint16
	MagicCookie   uint32
	TransactionID []byte
	Attributes    []*Attribute
}

type Attribute struct {
	Type   uint16
	Length uint16
	Value  []byte
}

func NewMessage(msgType uint16) (*Message, error) {
	tid := make([]byte, TransactionIDLen)
	if _, err := rand.Read(tid); err != nil {
		return nil, err
	}
	return &Message{
		Type:          msgType,
		MagicCookie:   MagicCookie,
		TransactionID: tid,
		Attributes:    make([]*Attribute, 0),
	}, nil
}

func (m *Message) AddAttribute(attr *Attribute) {
	m.Attributes = append(m.Attributes, attr)
	paddedLen := (attr.Length + 3) & ^uint16(3)
	m.Length += 4 + paddedLen
}

func (m *Message) GetAttribute(attrType uint16) *Attribute {
	for _, attr := range m.Attributes {
		if attr.Type == attrType {
			return attr
		}
	}
	return nil
}

func (m *Message) Encode() ([]byte, error) {
	buf := make([]byte, HeaderSize+m.Length)

	binary.BigEndian.PutUint16(buf[0:2], m.Type)
	binary.BigEndian.PutUint16(buf[2:4], m.Length)
	binary.BigEndian.PutUint32(buf[4:8], m.MagicCookie)
	copy(buf[8:20], m.TransactionID)

	offset := HeaderSize
	for _, attr := range m.Attributes {
		binary.BigEndian.PutUint16(buf[offset:offset+2], attr.Type)
		binary.BigEndian.PutUint16(buf[offset+2:offset+4], attr.Length)
		copy(buf[offset+4:], attr.Value)
		paddedLen := (attr.Length + 3) & ^uint16(3)
		offset += 4 + int(paddedLen)
	}

	return buf, nil
}

func DecodeMessage(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("message too short")
	}

	m := &Message{
		Type:          binary.BigEndian.Uint16(data[0:2]),
		Length:        binary.BigEndian.Uint16(data[2:4]),
		MagicCookie:   binary.BigEndian.Uint32(data[4:8]),
		TransactionID: make([]byte, TransactionIDLen),
		Attributes:    make([]*Attribute, 0),
	}
	copy(m.TransactionID, data[8:20])

	if m.MagicCookie != MagicCookie {
		return nil, errors.New("invalid magic cookie")
	}

	expectedLen := int(HeaderSize + m.Length)
	if len(data) < expectedLen {
		return nil, errors.New("message truncated")
	}

	offset := HeaderSize
	for offset < expectedLen {
		if expectedLen-offset < 4 {
			return nil, errors.New("attribute header truncated")
		}

		attr := &Attribute{
			Type:   binary.BigEndian.Uint16(data[offset : offset+2]),
			Length: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		}

		valueStart := offset + 4
		valueEnd := valueStart + int(attr.Length)
		if valueEnd > expectedLen {
			return nil, errors.New("attribute value truncated")
		}

		attr.Value = make([]byte, attr.Length)
		copy(attr.Value, data[valueStart:valueEnd])
		m.Attributes = append(m.Attributes, attr)

		paddedLen := (attr.Length + 3) & ^uint16(3)
		offset = valueStart + int(paddedLen)
	}

	return m, nil
}
