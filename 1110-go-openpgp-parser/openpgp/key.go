package openpgp

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"time"
)

type PublicKeyAlgorithm uint8

const (
	PubKeyAlgoRSA         PublicKeyAlgorithm = 1
	PubKeyAlgoRSAEncrypt  PublicKeyAlgorithm = 2
	PubKeyAlgoRSASign     PublicKeyAlgorithm = 3
	PubKeyAlgoElGamal     PublicKeyAlgorithm = 16
	PubKeyAlgoDSA         PublicKeyAlgorithm = 17
	PubKeyAlgoECDH        PublicKeyAlgorithm = 18
	PubKeyAlgoECDSA       PublicKeyAlgorithm = 19
	PubKeyAlgoElGamalEncrypt PublicKeyAlgorithm = 20
	PubKeyAlgoEdDSA       PublicKeyAlgorithm = 22
)

func (a PublicKeyAlgorithm) String() string {
	switch a {
	case PubKeyAlgoRSA:
		return "RSA (Encrypt or Sign)"
	case PubKeyAlgoRSAEncrypt:
		return "RSA Encrypt-Only"
	case PubKeyAlgoRSASign:
		return "RSA Sign-Only"
	case PubKeyAlgoElGamal:
		return "ElGamal (Encrypt-Only)"
	case PubKeyAlgoDSA:
		return "DSA (Digital Signature Algorithm)"
	case PubKeyAlgoECDH:
		return "ECDH (Elliptic Curve Diffie-Hellman)"
	case PubKeyAlgoECDSA:
		return "ECDSA (Elliptic Curve Digital Signature Algorithm)"
	case PubKeyAlgoEdDSA:
		return "EdDSA (Edwards-curve Digital Signature Algorithm)"
	default:
		return "Unknown"
	}
}

type PublicKeyInfo struct {
	Version      uint8
	CreationTime time.Time
	Algorithm    PublicKeyAlgorithm
	AlgorithmID  uint8
	KeyID        string
	Fingerprint  string
	FingerprintBytes []byte
}

type KeyInfo struct {
	PrimaryKey  *PublicKeyInfo
	Subkeys     []*PublicKeyInfo
	UserIDs     []string
	UserAttrs   [][]byte
}

func ParsePublicKeyPacket(packet *Packet) (*PublicKeyInfo, error) {
	if packet.Tag != TagPublicKey && packet.Tag != TagPublicSubkey {
		return nil, errors.New("packet is not a public key packet")
	}

	body := packet.Body
	if len(body) < 6 {
		return nil, errors.New("public key packet too short")
	}

	version := body[0]
	info := &PublicKeyInfo{
		Version: version,
	}

	if version == 3 {
		if len(body) < 9 {
			return nil, errors.New("v3 public key packet too short")
		}
		info.CreationTime = time.Unix(int64(uint32(body[1])<<24|uint32(body[2])<<16|uint32(body[3])<<8|uint32(body[4])), 0)
		info.AlgorithmID = body[8]
		info.Algorithm = PublicKeyAlgorithm(info.AlgorithmID)

		if len(body) < 9+2 {
			return nil, errors.New("v3 public key packet missing modulus length")
		}
		modulusBits := int(body[9])<<8 | int(body[10])
		modulusBytes := (modulusBits + 7) / 8
		nStart := 11
		if len(body) < nStart+modulusBytes+2 {
			return nil, errors.New("v3 public key packet truncated")
		}
		modulus := body[nStart : nStart+modulusBytes]
		keyID := modulus[len(modulus)-8:]
		info.KeyID = stringsToUpper(hex.EncodeToString(keyID))
		info.Fingerprint = info.KeyID
		info.FingerprintBytes = keyID
	} else if version == 4 {
		if len(body) < 6 {
			return nil, errors.New("v4 public key packet too short")
		}
		info.CreationTime = time.Unix(int64(uint32(body[1])<<24|uint32(body[2])<<16|uint32(body[3])<<8|uint32(body[4])), 0)
		info.AlgorithmID = body[5]
		info.Algorithm = PublicKeyAlgorithm(info.AlgorithmID)

		hash := sha1.New()
		header := []byte{
			0x99,
			byte(len(body) >> 8),
			byte(len(body)),
		}
		hash.Write(header)
		hash.Write(body)
		fingerprint := hash.Sum(nil)
		info.FingerprintBytes = fingerprint
		info.Fingerprint = stringsToUpper(hex.EncodeToString(fingerprint))
		info.KeyID = stringsToUpper(hex.EncodeToString(fingerprint[12:]))
	} else {
		return nil, errors.New("unsupported public key version")
	}

	return info, nil
}

func ParseKey(packets []*Packet) (*KeyInfo, error) {
	keyInfo := &KeyInfo{}

	for _, packet := range packets {
		switch packet.Tag {
		case TagPublicKey:
			info, err := ParsePublicKeyPacket(packet)
			if err != nil {
				return nil, err
			}
			keyInfo.PrimaryKey = info
		case TagPublicSubkey:
			info, err := ParsePublicKeyPacket(packet)
			if err != nil {
				return nil, err
			}
			keyInfo.Subkeys = append(keyInfo.Subkeys, info)
		case TagUserID:
			userID := string(packet.Body)
			keyInfo.UserIDs = append(keyInfo.UserIDs, userID)
		case TagUserAttribute:
			keyInfo.UserAttrs = append(keyInfo.UserAttrs, packet.Body)
		}
	}

	return keyInfo, nil
}

func stringsToUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 'a' + 'A'
		} else {
			result[i] = c
		}
	}
	return string(result)
}

type LiteralDataInfo struct {
	Format    byte
	Filename  string
	Timestamp time.Time
	Data      []byte
}

func ParseLiteralDataPacket(packet *Packet) (*LiteralDataInfo, error) {
	if packet.Tag != TagLiteralData {
		return nil, errors.New("packet is not a literal data packet")
	}

	body := packet.Body
	if len(body) < 2 {
		return nil, errors.New("literal data packet too short")
	}

	info := &LiteralDataInfo{
		Format: body[0],
	}

	filenameLen := int(body[1])
	if len(body) < 2+filenameLen+4 {
		return nil, errors.New("literal data packet truncated")
	}

	info.Filename = string(body[2 : 2+filenameLen])
	offset := 2 + filenameLen

	timestamp := uint32(body[offset])<<24 | uint32(body[offset+1])<<16 | uint32(body[offset+2])<<8 | uint32(body[offset+3])
	info.Timestamp = time.Unix(int64(timestamp), 0)
	offset += 4

	if offset <= len(body) {
		info.Data = body[offset:]
	}

	return info, nil
}

func ParseArmorAndPackets(data []byte) (*Armor, []*Packet, error) {
	armor, err := ParseArmor(data)
	if err != nil {
		return nil, nil, err
	}

	packets, err := ParsePackets(armor.Payload)
	if err != nil {
		return nil, nil, err
	}

	return armor, packets, nil
}
