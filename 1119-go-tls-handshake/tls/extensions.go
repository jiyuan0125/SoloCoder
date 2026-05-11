package tls

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type ExtensionType uint16

const (
	ExtensionTypeServerName         ExtensionType = 0x0000
	ExtensionTypeSupportedGroups    ExtensionType = 0x000A
	ExtensionTypeSignatureAlgorithms ExtensionType = 0x000D
	ExtensionTypeALPN               ExtensionType = 0x0010
)

type Extension interface {
	Type() ExtensionType
	Name() string
	Serialize() []byte
	String() string
}

type SNIExtension struct {
	HostName string
}

func (e *SNIExtension) Type() ExtensionType {
	return ExtensionTypeServerName
}

func (e *SNIExtension) Name() string {
	return "server_name"
}

func (e *SNIExtension) Serialize() []byte {
	nameLen := len(e.HostName)
	data := make([]byte, 2+1+2+nameLen)
	binary.BigEndian.PutUint16(data[:2], uint16(1+2+nameLen))
	data[2] = 0
	binary.BigEndian.PutUint16(data[3:5], uint16(nameLen))
	copy(data[5:], e.HostName)
	return data
}

func (e *SNIExtension) String() string {
	return fmt.Sprintf("SNI(HostName=%q)", e.HostName)
}

type NamedGroup uint16

const (
	GroupX25519    NamedGroup = 0x001D
	GroupSecP256R1 NamedGroup = 0x0017
	GroupSecP384R1 NamedGroup = 0x0018
	GroupSecP521R1 NamedGroup = 0x0019
)

var groupNames = map[NamedGroup]string{
	GroupX25519:    "x25519",
	GroupSecP256R1: "secp256r1",
	GroupSecP384R1: "secp384r1",
	GroupSecP521R1: "secp521r1",
}

type SupportedGroupsExtension struct {
	Groups []NamedGroup
}

func (e *SupportedGroupsExtension) Type() ExtensionType {
	return ExtensionTypeSupportedGroups
}

func (e *SupportedGroupsExtension) Name() string {
	return "supported_groups"
}

func (e *SupportedGroupsExtension) Serialize() []byte {
	length := 2 + len(e.Groups)*2
	data := make([]byte, 2+length)
	binary.BigEndian.PutUint16(data[:2], uint16(length))
	binary.BigEndian.PutUint16(data[2:4], uint16(len(e.Groups)*2))
	for i, g := range e.Groups {
		binary.BigEndian.PutUint16(data[4+i*2:], uint16(g))
	}
	return data
}

func (e *SupportedGroupsExtension) String() string {
	names := make([]string, 0, len(e.Groups))
	for _, g := range e.Groups {
		if name, ok := groupNames[g]; ok {
			names = append(names, name)
		} else {
			names = append(names, fmt.Sprintf("0x%04X", g))
		}
	}
	return fmt.Sprintf("SupportedGroups(%v)", names)
}

type SignatureAlgorithm uint16

const (
	SignatureRSA_PKCS1_SHA256    SignatureAlgorithm = 0x0401
	SignatureRSA_PKCS1_SHA384    SignatureAlgorithm = 0x0501
	SignatureRSA_PKCS1_SHA512    SignatureAlgorithm = 0x0601
	SignatureECDSA_P256_SHA256   SignatureAlgorithm = 0x0403
	SignatureECDSA_P384_SHA384   SignatureAlgorithm = 0x0503
	SignatureECDSA_P521_SHA512   SignatureAlgorithm = 0x0603
	SignatureRSA_PSS_RSAE_SHA256 SignatureAlgorithm = 0x0804
	SignatureRSA_PSS_RSAE_SHA384 SignatureAlgorithm = 0x0805
	SignatureRSA_PSS_RSAE_SHA512 SignatureAlgorithm = 0x0806
)

var sigAlgoNames = map[SignatureAlgorithm]string{
	SignatureRSA_PKCS1_SHA256:    "rsa_pkcs1_sha256",
	SignatureRSA_PKCS1_SHA384:    "rsa_pkcs1_sha384",
	SignatureRSA_PKCS1_SHA512:    "rsa_pkcs1_sha512",
	SignatureECDSA_P256_SHA256:   "ecdsa_secp256r1_sha256",
	SignatureECDSA_P384_SHA384:   "ecdsa_secp384r1_sha384",
	SignatureECDSA_P521_SHA512:   "ecdsa_secp521r1_sha512",
	SignatureRSA_PSS_RSAE_SHA256: "rsa_pss_rsae_sha256",
	SignatureRSA_PSS_RSAE_SHA384: "rsa_pss_rsae_sha384",
	SignatureRSA_PSS_RSAE_SHA512: "rsa_pss_rsae_sha512",
}

type SignatureAlgorithmsExtension struct {
	Algorithms []SignatureAlgorithm
}

func (e *SignatureAlgorithmsExtension) Type() ExtensionType {
	return ExtensionTypeSignatureAlgorithms
}

func (e *SignatureAlgorithmsExtension) Name() string {
	return "signature_algorithms"
}

func (e *SignatureAlgorithmsExtension) Serialize() []byte {
	length := 2 + len(e.Algorithms)*2
	data := make([]byte, 2+length)
	binary.BigEndian.PutUint16(data[:2], uint16(length))
	binary.BigEndian.PutUint16(data[2:4], uint16(len(e.Algorithms)*2))
	for i, a := range e.Algorithms {
		binary.BigEndian.PutUint16(data[4+i*2:], uint16(a))
	}
	return data
}

func (e *SignatureAlgorithmsExtension) String() string {
	names := make([]string, 0, len(e.Algorithms))
	for _, a := range e.Algorithms {
		if name, ok := sigAlgoNames[a]; ok {
			names = append(names, name)
		} else {
			names = append(names, fmt.Sprintf("0x%04X", a))
		}
	}
	return fmt.Sprintf("SignatureAlgorithms(%v)", names)
}

type ALPNExtension struct {
	Protocols []string
}

func (e *ALPNExtension) Type() ExtensionType {
	return ExtensionTypeALPN
}

func (e *ALPNExtension) Name() string {
	return "application_layer_protocol_negotiation"
}

func (e *ALPNExtension) Serialize() []byte {
	var totalLen int
	for _, p := range e.Protocols {
		totalLen += 1 + len(p)
	}
	data := make([]byte, 2+totalLen)
	binary.BigEndian.PutUint16(data[:2], uint16(totalLen))
	offset := 2
	for _, p := range e.Protocols {
		data[offset] = byte(len(p))
		offset++
		copy(data[offset:], p)
		offset += len(p)
	}
	return data
}

func (e *ALPNExtension) String() string {
	return fmt.Sprintf("ALPN(%v)", e.Protocols)
}

type RawExtension struct {
	ExtType ExtensionType
	Data    []byte
}

func (e *RawExtension) Type() ExtensionType {
	return e.ExtType
}

func (e *RawExtension) Name() string {
	return fmt.Sprintf("extension_%04X", e.ExtType)
}

func (e *RawExtension) Serialize() []byte {
	data := make([]byte, 2+len(e.Data))
	binary.BigEndian.PutUint16(data[:2], uint16(len(e.Data)))
	copy(data[2:], e.Data)
	return data
}

func (e *RawExtension) String() string {
	return fmt.Sprintf("RawExtension(Type=0x%04X, Length=%d)", e.ExtType, len(e.Data))
}

func ParseExtension(extType ExtensionType, data []byte) (Extension, error) {
	switch extType {
	case ExtensionTypeServerName:
		if len(data) < 2 {
			return nil, errors.New("invalid SNI extension")
		}
		listLen := binary.BigEndian.Uint16(data[:2])
		if int(listLen)+2 != len(data) {
			return nil, errors.New("invalid SNI list length")
		}
		if listLen < 3 {
			return nil, errors.New("invalid SNI entry")
		}
		offset := 2
		nameType := data[offset]
		offset++
		if nameType != 0 {
			return nil, fmt.Errorf("unsupported SNI name type: %d", nameType)
		}
		nameLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
		if int(nameLen)+offset != len(data) {
			return nil, errors.New("invalid SNI name length")
		}
		return &SNIExtension{HostName: string(data[offset : offset+int(nameLen)])}, nil

	case ExtensionTypeSupportedGroups:
		if len(data) < 2 {
			return nil, errors.New("invalid supported groups extension")
		}
		listLen := binary.BigEndian.Uint16(data[:2])
		if int(listLen)+2 != len(data) || listLen%2 != 0 {
			return nil, errors.New("invalid supported groups length")
		}
		groups := make([]NamedGroup, 0, listLen/2)
		for i := 2; i < len(data); i += 2 {
			groups = append(groups, NamedGroup(binary.BigEndian.Uint16(data[i:i+2])))
		}
		return &SupportedGroupsExtension{Groups: groups}, nil

	case ExtensionTypeSignatureAlgorithms:
		if len(data) < 2 {
			return nil, errors.New("invalid signature algorithms extension")
		}
		listLen := binary.BigEndian.Uint16(data[:2])
		if int(listLen)+2 != len(data) || listLen%2 != 0 {
			return nil, errors.New("invalid signature algorithms length")
		}
		algos := make([]SignatureAlgorithm, 0, listLen/2)
		for i := 2; i < len(data); i += 2 {
			algos = append(algos, SignatureAlgorithm(binary.BigEndian.Uint16(data[i:i+2])))
		}
		return &SignatureAlgorithmsExtension{Algorithms: algos}, nil

	case ExtensionTypeALPN:
		if len(data) < 2 {
			return nil, errors.New("invalid ALPN extension")
		}
		listLen := binary.BigEndian.Uint16(data[:2])
		if int(listLen)+2 != len(data) {
			return nil, errors.New("invalid ALPN list length")
		}
		var protocols []string
		offset := 2
		for offset < len(data) {
			protoLen := int(data[offset])
			offset++
			if offset+protoLen > len(data) {
				return nil, errors.New("invalid ALPN protocol length")
			}
			protocols = append(protocols, string(data[offset:offset+protoLen]))
			offset += protoLen
		}
		return &ALPNExtension{Protocols: protocols}, nil

	default:
		return &RawExtension{ExtType: extType, Data: data}, nil
	}
}

func ParseExtensions(data []byte) ([]Extension, error) {
	var extensions []Extension
	offset := 0
	for offset < len(data) {
		if offset+4 > len(data) {
			return nil, errors.New("invalid extension data")
		}
		extType := ExtensionType(binary.BigEndian.Uint16(data[offset : offset+2]))
		extLen := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4
		if offset+int(extLen) > len(data) {
			return nil, errors.New("invalid extension length")
		}
		extData := data[offset : offset+int(extLen)]
		ext, err := ParseExtension(extType, extData)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, ext)
		offset += int(extLen)
	}
	return extensions, nil
}

func SerializeExtensions(extensions []Extension) []byte {
	if len(extensions) == 0 {
		return nil
	}
	var totalLen int
	for _, ext := range extensions {
		extData := ext.Serialize()
		totalLen += 4 + len(extData) - 2
	}
	data := make([]byte, 2+totalLen)
	binary.BigEndian.PutUint16(data[:2], uint16(totalLen))
	offset := 2
	for _, ext := range extensions {
		binary.BigEndian.PutUint16(data[offset:], uint16(ext.Type()))
		offset += 2
		extData := ext.Serialize()
		copy(data[offset:], extData)
		offset += len(extData)
	}
	return data
}

func ValidateExtensionOrder(extensions []Extension) error {
	foundSNI := false
	for i, ext := range extensions {
		if ext.Type() == ExtensionTypeServerName {
			if i != 0 {
				return errors.New("SNI extension must be the first extension")
			}
			foundSNI = true
			break
		}
	}
	if len(extensions) > 0 && !foundSNI {
		for i, ext := range extensions {
			if ext.Type() == ExtensionTypeServerName && i != 0 {
				return errors.New("SNI extension must be the first extension")
			}
		}
	}
	return nil
}

func FindExtension(extensions []Extension, extType ExtensionType) (Extension, bool) {
	for _, ext := range extensions {
		if ext.Type() == extType {
			return ext, true
		}
	}
	return nil, false
}
