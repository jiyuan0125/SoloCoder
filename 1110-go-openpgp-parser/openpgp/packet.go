package openpgp

import (
	"bytes"
	"errors"
	"io"
)

type PacketTag uint8

const (
	TagReserved                    PacketTag = 0
	TagPublicEncryptedSessionKey   PacketTag = 1
	TagSignature                   PacketTag = 2
	TagSymmetricKeyEncryptedSessionKey PacketTag = 3
	TagOnePassSignature            PacketTag = 4
	TagSecretKey                   PacketTag = 5
	TagPublicKey                   PacketTag = 6
	TagSecretSubkey                PacketTag = 7
	TagCompressedData              PacketTag = 8
	TagSymmetricallyEncrypted      PacketTag = 9
	TagMarker                      PacketTag = 10
	TagLiteralData                 PacketTag = 11
	TagTrust                       PacketTag = 12
	TagUserID                      PacketTag = 13
	TagPublicSubkey                PacketTag = 14
	TagUserAttribute               PacketTag = 17
	TagSymmetricEncryptedIntegrityProtected PacketTag = 18
	TagModificationDetectionCode   PacketTag = 19
	TagPrivateExperimental1        PacketTag = 60
	TagPrivateExperimental2        PacketTag = 61
	TagPrivateExperimental3        PacketTag = 62
	TagPrivateExperimental4        PacketTag = 63
)

func (t PacketTag) String() string {
	switch t {
	case TagReserved:
		return "Reserved"
	case TagPublicEncryptedSessionKey:
		return "Public-Key Encrypted Session Key"
	case TagSignature:
		return "Signature"
	case TagSymmetricKeyEncryptedSessionKey:
		return "Symmetric-Key Encrypted Session Key"
	case TagOnePassSignature:
		return "One-Pass Signature"
	case TagSecretKey:
		return "Secret Key"
	case TagPublicKey:
		return "Public Key"
	case TagSecretSubkey:
		return "Secret Subkey"
	case TagCompressedData:
		return "Compressed Data"
	case TagSymmetricallyEncrypted:
		return "Symmetrically Encrypted Data"
	case TagMarker:
		return "Marker"
	case TagLiteralData:
		return "Literal Data"
	case TagTrust:
		return "Trust"
	case TagUserID:
		return "User ID"
	case TagPublicSubkey:
		return "Public Subkey"
	case TagUserAttribute:
		return "User Attribute"
	case TagSymmetricEncryptedIntegrityProtected:
		return "Symmetrically Encrypted Integrity Protected Data"
	case TagModificationDetectionCode:
		return "Modification Detection Code"
	default:
		if t >= 60 && t <= 63 {
			return "Private/Experimental"
		}
		return "Unknown"
	}
}

type Packet struct {
	Tag          PacketTag
	TagValue     uint8
	IsNewFormat  bool
	Length       uint64
	IsIndefinite bool
	Body         []byte
	Subpackets   []Subpacket
}

type Subpacket struct {
	Type   uint8
	Length uint32
	Data   []byte
}

func ParsePackets(data []byte) ([]*Packet, error) {
	return parsePacketsReader(bytes.NewReader(data))
}

func parsePacketsReader(r io.Reader) ([]*Packet, error) {
	var packets []*Packet

	for {
		packet, err := parsePacket(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}

	return packets, nil
}

func parsePacket(r io.Reader) (*Packet, error) {
	var tagByte [1]byte
	_, err := io.ReadFull(r, tagByte[:])
	if err != nil {
		return nil, err
	}

	if tagByte[0]&0x80 == 0 {
		return nil, errors.New("invalid packet: bit 7 must be set")
	}

	packet := &Packet{}

	if tagByte[0]&0x40 != 0 {
		packet.IsNewFormat = true
		packet.TagValue = tagByte[0] & 0x3F
		packet.Tag = PacketTag(packet.TagValue)

		length, isIndefinite, err := readNewLength(r)
		if err != nil {
			return nil, err
		}
		packet.Length = length
		packet.IsIndefinite = isIndefinite
	} else {
		packet.IsNewFormat = false
		lengthType := tagByte[0] & 0x03
		packet.TagValue = (tagByte[0] >> 2) & 0x0F
		packet.Tag = PacketTag(packet.TagValue)

		length, err := readOldLength(r, lengthType)
		if err != nil {
			return nil, err
		}
		packet.Length = length
		packet.IsIndefinite = (lengthType == 3)
	}

	if packet.IsIndefinite {
		body, err := readPartialBody(r)
		if err != nil {
			return nil, err
		}
		packet.Body = body
		packet.Length = uint64(len(body))
	} else {
		body := make([]byte, packet.Length)
		_, err = io.ReadFull(r, body)
		if err != nil {
			return nil, err
		}
		packet.Body = body
	}

	if packet.Tag == TagSignature {
		subpackets, err := parseSignatureSubpackets(packet.Body)
		if err != nil {
			return nil, err
		}
		packet.Subpackets = subpackets
	}

	return packet, nil
}

func readNewLength(r io.Reader) (uint64, bool, error) {
	var firstByte [1]byte
	_, err := io.ReadFull(r, firstByte[:])
	if err != nil {
		return 0, false, err
	}

	b := firstByte[0]
	switch {
	case b < 192:
		return uint64(b), false, nil
	case b < 224:
		var secondByte [1]byte
		_, err := io.ReadFull(r, secondByte[:])
		if err != nil {
			return 0, false, err
		}
		return (uint64(b-192) << 8) + uint64(secondByte[0]) + 192, false, nil
	case b == 255:
		var lengthBytes [4]byte
		_, err := io.ReadFull(r, lengthBytes[:])
		if err != nil {
			return 0, false, err
		}
		return uint64(lengthBytes[0])<<24 |
			uint64(lengthBytes[1])<<16 |
			uint64(lengthBytes[2])<<8 |
			uint64(lengthBytes[3]), false, nil
	default:
		return uint64(1 << (b & 0x1F)), true, nil
	}
}

func readOldLength(r io.Reader, lengthType uint8) (uint64, error) {
	switch lengthType {
	case 0:
		var lengthByte [1]byte
		_, err := io.ReadFull(r, lengthByte[:])
		if err != nil {
			return 0, err
		}
		return uint64(lengthByte[0]), nil
	case 1:
		var lengthBytes [2]byte
		_, err := io.ReadFull(r, lengthBytes[:])
		if err != nil {
			return 0, err
		}
		return uint64(lengthBytes[0])<<8 | uint64(lengthBytes[1]), nil
	case 2:
		var lengthBytes [4]byte
		_, err := io.ReadFull(r, lengthBytes[:])
		if err != nil {
			return 0, err
		}
		return uint64(lengthBytes[0])<<24 |
			uint64(lengthBytes[1])<<16 |
			uint64(lengthBytes[2])<<8 |
			uint64(lengthBytes[3]), nil
	case 3:
		return 0, nil
	default:
		return 0, errors.New("invalid old length type")
	}
}

func readPartialBody(r io.Reader) ([]byte, error) {
	var result []byte
	for {
		var firstByte [1]byte
		_, err := io.ReadFull(r, firstByte[:])
		if err != nil {
			return nil, err
		}

		b := firstByte[0]
		var partialLen uint64

		switch {
		case b < 192:
			partialLen = uint64(b)
		case b < 224:
			var secondByte [1]byte
			_, err := io.ReadFull(r, secondByte[:])
			if err != nil {
				return nil, err
			}
			partialLen = (uint64(b-192) << 8) + uint64(secondByte[0]) + 192
		case b == 255:
			var lengthBytes [4]byte
			_, err := io.ReadFull(r, lengthBytes[:])
			if err != nil {
				return nil, err
			}
			partialLen = uint64(lengthBytes[0])<<24 |
				uint64(lengthBytes[1])<<16 |
				uint64(lengthBytes[2])<<8 |
				uint64(lengthBytes[3])
			body := make([]byte, partialLen)
			_, err = io.ReadFull(r, body)
			if err != nil {
				return nil, err
			}
			return append(result, body...), nil
		default:
			partialLen = uint64(1 << (b & 0x1F))
		}

		body := make([]byte, partialLen)
		_, err = io.ReadFull(r, body)
		if err != nil {
			return nil, err
		}
		result = append(result, body...)
	}
}

func parseSignatureSubpackets(body []byte) ([]Subpacket, error) {
	if len(body) < 4 {
		return nil, nil
	}

	var subpackets []Subpacket

	version := body[0]
	if version != 3 && version != 4 {
		return nil, errors.New("unsupported signature version")
	}

	if version == 4 {
		offset := 4
		hashedSubpacketsLen := uint16(body[offset])<<8 | uint16(body[offset+1])
		offset += 2

		if hashedSubpacketsLen > 0 && int(offset)+int(hashedSubpacketsLen) <= len(body) {
			parsed, err := parseSubpacketList(body[offset : offset+int(hashedSubpacketsLen)])
			if err != nil {
				return nil, err
			}
			subpackets = append(subpackets, parsed...)
			offset += int(hashedSubpacketsLen)
		}

		if offset+2 <= len(body) {
			unhashedSubpacketsLen := uint16(body[offset])<<8 | uint16(body[offset+1])
			offset += 2

			if unhashedSubpacketsLen > 0 && int(offset)+int(unhashedSubpacketsLen) <= len(body) {
				parsed, err := parseSubpacketList(body[offset : offset+int(unhashedSubpacketsLen)])
				if err != nil {
					return nil, err
				}
				subpackets = append(subpackets, parsed...)
			}
		}
	}

	return subpackets, nil
}

func parseSubpacketList(data []byte) ([]Subpacket, error) {
	var subpackets []Subpacket
	offset := 0

	for offset < len(data) {
		subLength, bytesRead, err := parseSubpacketLength(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += bytesRead

		if offset >= len(data) {
			break
		}

		subType := data[offset]
		offset++

		if offset+int(subLength-1) > len(data) {
			return nil, errors.New("subpacket data truncated")
		}

		subData := data[offset : offset+int(subLength-1)]
		offset += int(subLength - 1)

		subpackets = append(subpackets, Subpacket{
			Type:   subType,
			Length: subLength,
			Data:   subData,
		})
	}

	return subpackets, nil
}

func parseSubpacketLength(data []byte) (uint32, int, error) {
	if len(data) == 0 {
		return 0, 0, io.EOF
	}

	firstByte := data[0]
	switch {
	case firstByte < 192:
		return uint32(firstByte), 1, nil
	case firstByte < 255:
		if len(data) < 2 {
			return 0, 0, errors.New("subpacket length truncated")
		}
		return uint32(firstByte-192)*256 + uint32(data[1]) + 192, 2, nil
	case firstByte == 255:
		if len(data) < 5 {
			return 0, 0, errors.New("subpacket length truncated")
		}
		return uint32(data[1])<<24 |
			uint32(data[2])<<16 |
			uint32(data[3])<<8 |
			uint32(data[4]), 5, nil
	}
	return 0, 0, errors.New("invalid subpacket length")
}
