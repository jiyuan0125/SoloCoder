package stun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"net"
)

type MappedAddress struct {
	Family byte
	IP     net.IP
	Port   uint16
}

func NewAttribute(attrType uint16, value []byte) *Attribute {
	return &Attribute{
		Type:   attrType,
		Length: uint16(len(value)),
		Value:  value,
	}
}

func NewXORMappedAddress(magicCookie uint32, transactionID []byte, family byte, ip net.IP, port uint16) (*Attribute, error) {
	xorIP, xorPort := xorAddress(magicCookie, transactionID, family, ip, port)
	return encodeMappedAddress(family, xorIP, xorPort)
}

func DecodeXORMappedAddress(attr *Attribute, magicCookie uint32, transactionID []byte) (*MappedAddress, error) {
	addr, err := decodeMappedAddress(attr)
	if err != nil {
		return nil, err
	}
	realIP, realPort := xorAddress(magicCookie, transactionID, addr.Family, addr.IP, addr.Port)
	return &MappedAddress{
		Family: addr.Family,
		IP:     realIP,
		Port:   realPort,
	}, nil
}

func xorAddress(magicCookie uint32, transactionID []byte, family byte, ip net.IP, port uint16) (net.IP, uint16) {
	magicBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(magicBytes, magicCookie)

	xorPort := port ^ uint16(magicCookie>>16)

	var xorIP net.IP
	if family == FamilyIPv4 {
		xorIP = make(net.IP, 4)
		for i := 0; i < 4; i++ {
			xorIP[i] = ip[i] ^ magicBytes[i]
		}
	} else {
		xorIP = make(net.IP, 16)
		xorMask := make([]byte, 16)
		copy(xorMask[0:4], magicBytes)
		copy(xorMask[4:16], transactionID)
		for i := 0; i < 16; i++ {
			xorIP[i] = ip[i] ^ xorMask[i]
		}
	}

	return xorIP, xorPort
}

func encodeMappedAddress(family byte, ip net.IP, port uint16) (*Attribute, error) {
	var value []byte
	if family == FamilyIPv4 {
		if len(ip) != 4 {
			return nil, errors.New("invalid IPv4 address length")
		}
		value = make([]byte, 8)
		value[1] = family
		binary.BigEndian.PutUint16(value[2:4], port)
		copy(value[4:8], ip)
	} else {
		if len(ip) != 16 {
			return nil, errors.New("invalid IPv6 address length")
		}
		value = make([]byte, 20)
		value[1] = family
		binary.BigEndian.PutUint16(value[2:4], port)
		copy(value[4:20], ip)
	}
	return NewAttribute(AttrXORMappedAddress, value), nil
}

func decodeMappedAddress(attr *Attribute) (*MappedAddress, error) {
	if len(attr.Value) < 8 {
		return nil, errors.New("mapped address attribute too short")
	}

	family := attr.Value[1]
	port := binary.BigEndian.Uint16(attr.Value[2:4])

	var ip net.IP
	if family == FamilyIPv4 {
		if len(attr.Value) < 8 {
			return nil, errors.New("invalid IPv4 mapped address")
		}
		ip = make(net.IP, 4)
		copy(ip, attr.Value[4:8])
	} else if family == FamilyIPv6 {
		if len(attr.Value) < 20 {
			return nil, errors.New("invalid IPv6 mapped address")
		}
		ip = make(net.IP, 16)
		copy(ip, attr.Value[4:20])
	} else {
		return nil, errors.New("unknown address family")
	}

	return &MappedAddress{
		Family: family,
		IP:     ip,
		Port:   port,
	}, nil
}

func NewMessageIntegrity(msg *Message, password []byte) (*Attribute, error) {
	tempMsg := cloneMessageForIntegrity(msg)
	tempData, err := tempMsg.Encode()
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha1.New, password)
	h.Write(tempData)
	mac := h.Sum(nil)
	return NewAttribute(AttrMessageIntegrity, mac), nil
}

func VerifyMessageIntegrity(msg *Message, password []byte, originalData []byte) bool {
	attr := msg.GetAttribute(AttrMessageIntegrity)
	if attr == nil {
		return false
	}

	truncated, err := truncateForIntegrity(originalData)
	if err != nil {
		return false
	}

	h := hmac.New(sha1.New, password)
	h.Write(truncated)
	expected := h.Sum(nil)
	return hmac.Equal(attr.Value, expected)
}

func cloneMessageForIntegrity(msg *Message) *Message {
	clone := &Message{
		Type:          msg.Type,
		MagicCookie:   msg.MagicCookie,
		TransactionID: make([]byte, len(msg.TransactionID)),
		Attributes:    make([]*Attribute, 0),
	}
	copy(clone.TransactionID, msg.TransactionID)

	var length uint16
	for _, attr := range msg.Attributes {
		if attr.Type == AttrFingerprint {
			continue
		}
		if attr.Type == AttrMessageIntegrity {
			placeholder := NewAttribute(AttrMessageIntegrity, make([]byte, 20))
			clone.Attributes = append(clone.Attributes, placeholder)
			length += 24
			continue
		}
		clone.Attributes = append(clone.Attributes, attr)
		paddedLen := (attr.Length + 3) & ^uint16(3)
		length += 4 + paddedLen
	}
	clone.Length = length
	return clone
}

func truncateForIntegrity(data []byte) ([]byte, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("message too short")
	}

	msgLen := binary.BigEndian.Uint16(data[2:4])
	totalLen := int(HeaderSize + msgLen)

	if len(data) < totalLen {
		return nil, errors.New("message truncated")
	}

	offset := HeaderSize
	for offset < totalLen {
		if totalLen-offset < 4 {
			return nil, errors.New("attribute header truncated")
		}
		attrType := binary.BigEndian.Uint16(data[offset : offset+2])
		attrLen := binary.BigEndian.Uint16(data[offset+2 : offset+4])

		if attrType == AttrMessageIntegrity {
			return data[:offset+24], nil
		}

		paddedLen := (attrLen + 3) & ^uint16(3)
		offset += 4 + int(paddedLen)
	}

	return data[:totalLen], nil
}

const FingerprintXOR uint32 = 0x5354554e

func NewFingerprint(msg *Message) (*Attribute, error) {
	tempMsg := cloneMessageForFingerprint(msg)
	tempData, err := tempMsg.Encode()
	if err != nil {
		return nil, err
	}

	crc := crc32.ChecksumIEEE(tempData) ^ FingerprintXOR
	value := make([]byte, 4)
	binary.BigEndian.PutUint32(value, crc)
	return NewAttribute(AttrFingerprint, value), nil
}

func VerifyFingerprint(msg *Message, originalData []byte) bool {
	attr := msg.GetAttribute(AttrFingerprint)
	if attr == nil || len(attr.Value) != 4 {
		return false
	}

	expected := binary.BigEndian.Uint32(attr.Value)

	offset := 0
	for i, a := range msg.Attributes {
		if a.Type == AttrFingerprint {
			offset = i
			break
		}
	}

	truncatedMsg := &Message{
		Type:          msg.Type,
		MagicCookie:   msg.MagicCookie,
		TransactionID: msg.TransactionID,
		Attributes:    msg.Attributes[:offset],
	}

	var length uint16
	for _, a := range truncatedMsg.Attributes {
		paddedLen := (a.Length + 3) & ^uint16(3)
		length += 4 + paddedLen
	}
	truncatedMsg.Length = length

	truncatedData, err := truncatedMsg.Encode()
	if err != nil {
		return false
	}

	crc := crc32.ChecksumIEEE(truncatedData) ^ FingerprintXOR
	return crc == expected
}

func cloneMessageForFingerprint(msg *Message) *Message {
	clone := &Message{
		Type:          msg.Type,
		MagicCookie:   msg.MagicCookie,
		TransactionID: make([]byte, len(msg.TransactionID)),
		Attributes:    make([]*Attribute, 0),
	}
	copy(clone.TransactionID, msg.TransactionID)

	var length uint16
	for _, attr := range msg.Attributes {
		if attr.Type == AttrFingerprint {
			continue
		}
		clone.Attributes = append(clone.Attributes, attr)
		paddedLen := (attr.Length + 3) & ^uint16(3)
		length += 4 + paddedLen
	}
	clone.Length = length
	return clone
}
