package tls

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type TLSVersion uint16

const (
	TLSVersion1_2 TLSVersion = 0x0303
)

type HandshakeType uint8

const (
	HandshakeTypeClientHello        HandshakeType = 1
	HandshakeTypeServerHello        HandshakeType = 2
	HandshakeTypeCertificate        HandshakeType = 11
	HandshakeTypeServerKeyExchange  HandshakeType = 12
	HandshakeTypeServerHelloDone    HandshakeType = 14
	HandshakeTypeClientKeyExchange  HandshakeType = 16
	HandshakeTypeFinished           HandshakeType = 20
)

type ContentType uint8

const (
	ContentTypeHandshake    ContentType = 22
	ContentTypeChangeCipher ContentType = 20
)

type Random struct {
	GmtUnixTime uint32
	RandomBytes [28]byte
}

func NewRandom() Random {
	var r Random
	r.GmtUnixTime = uint32(time.Now().Unix())
	rand.Read(r.RandomBytes[:])
	return r
}

func (r Random) Bytes() []byte {
	data := make([]byte, 32)
	binary.BigEndian.PutUint32(data[:4], r.GmtUnixTime)
	copy(data[4:], r.RandomBytes[:])
	return data
}

func (r Random) String() string {
	return fmt.Sprintf("Random(UnixTime=%d, Bytes=%s...)", r.GmtUnixTime, hex.EncodeToString(r.RandomBytes[:8]))
}

func ParseRandom(data []byte) (Random, error) {
	if len(data) != 32 {
		return Random{}, errors.New("random must be 32 bytes")
	}
	var r Random
	r.GmtUnixTime = binary.BigEndian.Uint32(data[:4])
	copy(r.RandomBytes[:], data[4:])
	return r, nil
}

type SessionID []byte

func (s SessionID) String() string {
	if len(s) == 0 {
		return "SessionID(empty)"
	}
	return fmt.Sprintf("SessionID(%s)", hex.EncodeToString(s))
}

type ClientHelloMessage struct {
	Version       TLSVersion
	Random        Random
	SessionID     SessionID
	CipherSuites  []CipherSuiteID
	Compression   []byte
	Extensions    []Extension
	RawData       []byte
}

func NewClientHello(version TLSVersion, sessionID SessionID, cipherSuites []CipherSuiteID, extensions []Extension) *ClientHelloMessage {
	return &ClientHelloMessage{
		Version:      version,
		Random:       NewRandom(),
		SessionID:    sessionID,
		CipherSuites: cipherSuites,
		Compression:  []byte{0},
		Extensions:   extensions,
	}
}

func (m *ClientHelloMessage) Type() HandshakeType {
	return HandshakeTypeClientHello
}

func (m *ClientHelloMessage) Serialize() ([]byte, error) {
	if err := ValidateExtensionOrder(m.Extensions); err != nil {
		return nil, err
	}
	cipherData := SerializeCipherSuites(m.CipherSuites)
	extData := SerializeExtensions(m.Extensions)
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint16(m.Version))
	buf.Write(m.Random.Bytes())
	buf.WriteByte(byte(len(m.SessionID)))
	buf.Write(m.SessionID)
	buf.Write(cipherData)
	buf.WriteByte(byte(len(m.Compression)))
	buf.Write(m.Compression)
	if extData != nil {
		buf.Write(extData)
	}
	return buf.Bytes(), nil
}

func (m *ClientHelloMessage) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "ClientHello {\n")
	fmt.Fprintf(&b, "  Version: 0x%04X (TLS 1.2)\n", m.Version)
	fmt.Fprintf(&b, "  Random: %s\n", m.Random)
	fmt.Fprintf(&b, "  %s\n", m.SessionID)
	fmt.Fprintf(&b, "  CipherSuites (%d):\n", len(m.CipherSuites))
	for _, id := range m.CipherSuites {
		if info, ok := id.GetInfo(); ok {
			fmt.Fprintf(&b, "    - 0x%04X %s\n", id, info.Name)
		} else {
			fmt.Fprintf(&b, "    - 0x%04X (unknown)\n", id)
		}
	}
	fmt.Fprintf(&b, "  Compression: %v\n", m.Compression)
	fmt.Fprintf(&b, "  Extensions (%d):\n", len(m.Extensions))
	for _, ext := range m.Extensions {
		fmt.Fprintf(&b, "    - %s\n", ext)
	}
	fmt.Fprintf(&b, "}")
	return b.String()
}

func ParseClientHello(data []byte) (*ClientHelloMessage, error) {
	if len(data) < 2+32+1+2+2+1+1 {
		return nil, errors.New("ClientHello too short")
	}
	offset := 0
	version := TLSVersion(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	random, err := ParseRandom(data[offset : offset+32])
	if err != nil {
		return nil, err
	}
	offset += 32
	sessionLen := int(data[offset])
	offset++
	if offset+sessionLen > len(data) {
		return nil, errors.New("invalid session ID length")
	}
	sessionID := make(SessionID, sessionLen)
	copy(sessionID, data[offset:offset+sessionLen])
	offset += sessionLen
	if offset+2 > len(data) {
		return nil, errors.New("invalid cipher suites")
	}
	cipherLen := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	if offset+int(cipherLen) > len(data) {
		return nil, errors.New("invalid cipher suites length")
	}
	cipherSuites, err := ParseCipherSuites(data[offset-2 : offset+int(cipherLen)])
	if err != nil {
		return nil, err
	}
	offset += int(cipherLen)
	if offset+1 > len(data) {
		return nil, errors.New("invalid compression")
	}
	compressionLen := int(data[offset])
	offset++
	if offset+compressionLen > len(data) {
		return nil, errors.New("invalid compression length")
	}
	compression := make([]byte, compressionLen)
	copy(compression, data[offset:offset+compressionLen])
	offset += compressionLen
	var extensions []Extension
	if offset < len(data) {
		if offset+2 > len(data) {
			return nil, errors.New("invalid extensions header")
		}
		extTotalLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
		if offset+int(extTotalLen) != len(data) {
			return nil, errors.New("invalid extensions total length")
		}
		extData := data[offset : offset+int(extTotalLen)]
		extensions, err = ParseExtensions(extData)
		if err != nil {
			return nil, err
		}
		if err := ValidateExtensionOrder(extensions); err != nil {
			return nil, err
		}
	}
	return &ClientHelloMessage{
		Version:      version,
		Random:       random,
		SessionID:    sessionID,
		CipherSuites: cipherSuites,
		Compression:  compression,
		Extensions:   extensions,
		RawData:      data,
	}, nil
}

type ServerHelloMessage struct {
	Version       TLSVersion
	Random        Random
	SessionID     SessionID
	CipherSuite   CipherSuiteID
	Compression   uint8
	Extensions    []Extension
	RawData       []byte
}

func NewServerHello(version TLSVersion, sessionID SessionID, cipherSuite CipherSuiteID, extensions []Extension) *ServerHelloMessage {
	return &ServerHelloMessage{
		Version:     version,
		Random:      NewRandom(),
		SessionID:   sessionID,
		CipherSuite: cipherSuite,
		Compression: 0,
		Extensions:  extensions,
	}
}

func (m *ServerHelloMessage) Type() HandshakeType {
	return HandshakeTypeServerHello
}

func (m *ServerHelloMessage) Serialize() ([]byte, error) {
	extData := SerializeExtensions(m.Extensions)
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint16(m.Version))
	buf.Write(m.Random.Bytes())
	buf.WriteByte(byte(len(m.SessionID)))
	buf.Write(m.SessionID)
	binary.Write(&buf, binary.BigEndian, uint16(m.CipherSuite))
	buf.WriteByte(m.Compression)
	if extData != nil {
		buf.Write(extData)
	}
	return buf.Bytes(), nil
}

func (m *ServerHelloMessage) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "ServerHello {\n")
	fmt.Fprintf(&b, "  Version: 0x%04X (TLS 1.2)\n", m.Version)
	fmt.Fprintf(&b, "  Random: %s\n", m.Random)
	fmt.Fprintf(&b, "  %s\n", m.SessionID)
	if info, ok := m.CipherSuite.GetInfo(); ok {
		fmt.Fprintf(&b, "  CipherSuite: 0x%04X %s\n", m.CipherSuite, info.Name)
	} else {
		fmt.Fprintf(&b, "  CipherSuite: 0x%04X (unknown)\n", m.CipherSuite)
	}
	fmt.Fprintf(&b, "  Compression: 0x%02X\n", m.Compression)
	fmt.Fprintf(&b, "  Extensions (%d):\n", len(m.Extensions))
	for _, ext := range m.Extensions {
		fmt.Fprintf(&b, "    - %s\n", ext)
	}
	fmt.Fprintf(&b, "}")
	return b.String()
}

func ParseServerHello(data []byte) (*ServerHelloMessage, error) {
	if len(data) < 2+32+1+2+1 {
		return nil, errors.New("ServerHello too short")
	}
	offset := 0
	version := TLSVersion(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	random, err := ParseRandom(data[offset : offset+32])
	if err != nil {
		return nil, err
	}
	offset += 32
	sessionLen := int(data[offset])
	offset++
	if offset+sessionLen > len(data) {
		return nil, errors.New("invalid session ID length")
	}
	sessionID := make(SessionID, sessionLen)
	copy(sessionID, data[offset:offset+sessionLen])
	offset += sessionLen
	cipherSuite := CipherSuiteID(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	compression := data[offset]
	offset++
	var extensions []Extension
	if offset < len(data) {
		if offset+2 > len(data) {
			return nil, errors.New("invalid extensions header")
		}
		extTotalLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
		if offset+int(extTotalLen) != len(data) {
			return nil, errors.New("invalid extensions total length")
		}
		extData := data[offset : offset+int(extTotalLen)]
		extensions, err = ParseExtensions(extData)
		if err != nil {
			return nil, err
		}
	}
	return &ServerHelloMessage{
		Version:     version,
		Random:      random,
		SessionID:   sessionID,
		CipherSuite: cipherSuite,
		Compression: compression,
		Extensions:  extensions,
		RawData:     data,
	}, nil
}

type CertificateMessage struct {
	Certificates [][]byte
	RawData      []byte
}

func NewCertificateMessage(certs [][]byte) *CertificateMessage {
	return &CertificateMessage{Certificates: certs}
}

func (m *CertificateMessage) Type() HandshakeType {
	return HandshakeTypeCertificate
}

func (m *CertificateMessage) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	var certsData bytes.Buffer
	for _, cert := range m.Certificates {
		certLen := make([]byte, 3)
		certLen[0] = byte(len(cert) >> 16)
		certLen[1] = byte(len(cert) >> 8)
		certLen[2] = byte(len(cert))
		certsData.Write(certLen)
		certsData.Write(cert)
	}
	totalLen := certsData.Len()
	buf.WriteByte(byte(totalLen >> 16))
	buf.WriteByte(byte(totalLen >> 8))
	buf.WriteByte(byte(totalLen))
	buf.Write(certsData.Bytes())
	return buf.Bytes(), nil
}

func (m *CertificateMessage) String() string {
	return fmt.Sprintf("Certificate { Certificates: %d }", len(m.Certificates))
}

func ParseCertificate(data []byte) (*CertificateMessage, error) {
	if len(data) < 3 {
		return nil, errors.New("Certificate message too short")
	}
	totalLen := int(data[0])<<16 | int(data[1])<<8 | int(data[2])
	if totalLen+3 != len(data) {
		return nil, errors.New("invalid certificate total length")
	}
	offset := 3
	var certs [][]byte
	for offset < len(data) {
		if offset+3 > len(data) {
			return nil, errors.New("invalid certificate length")
		}
		certLen := int(data[offset])<<16 | int(data[offset+1])<<8 | int(data[offset+2])
		offset += 3
		if offset+certLen > len(data) {
			return nil, errors.New("invalid certificate data")
		}
		certs = append(certs, data[offset:offset+certLen])
		offset += certLen
	}
	return &CertificateMessage{Certificates: certs, RawData: data}, nil
}

type ServerKeyExchangeMessage struct {
	RawData []byte
}

func NewServerKeyExchangeMessage() *ServerKeyExchangeMessage {
	return &ServerKeyExchangeMessage{}
}

func (m *ServerKeyExchangeMessage) Type() HandshakeType {
	return HandshakeTypeServerKeyExchange
}

func (m *ServerKeyExchangeMessage) Serialize() ([]byte, error) {
	return []byte{}, nil
}

func (m *ServerKeyExchangeMessage) String() string {
	return "ServerKeyExchange"
}

func ParseServerKeyExchange(data []byte) (*ServerKeyExchangeMessage, error) {
	return &ServerKeyExchangeMessage{RawData: data}, nil
}

type ServerHelloDoneMessage struct{}

func NewServerHelloDoneMessage() *ServerHelloDoneMessage {
	return &ServerHelloDoneMessage{}
}

func (m *ServerHelloDoneMessage) Type() HandshakeType {
	return HandshakeTypeServerHelloDone
}

func (m *ServerHelloDoneMessage) Serialize() ([]byte, error) {
	return []byte{}, nil
}

func (m *ServerHelloDoneMessage) String() string {
	return "ServerHelloDone"
}

func ParseServerHelloDone(data []byte) (*ServerHelloDoneMessage, error) {
	return &ServerHelloDoneMessage{}, nil
}

type ClientKeyExchangeMessage struct {
	ExchangeData []byte
	RawData      []byte
}

func NewClientKeyExchangeMessage(exchangeData []byte) *ClientKeyExchangeMessage {
	return &ClientKeyExchangeMessage{ExchangeData: exchangeData}
}

func (m *ClientKeyExchangeMessage) Type() HandshakeType {
	return HandshakeTypeClientKeyExchange
}

func (m *ClientKeyExchangeMessage) Serialize() ([]byte, error) {
	if len(m.ExchangeData) == 0 {
		return []byte{}, nil
	}
	if len(m.ExchangeData) > 0xFFFF {
		return nil, errors.New("exchange data too long")
	}
	data := make([]byte, 2+len(m.ExchangeData))
	binary.BigEndian.PutUint16(data[:2], uint16(len(m.ExchangeData)))
	copy(data[2:], m.ExchangeData)
	return data, nil
}

func (m *ClientKeyExchangeMessage) String() string {
	return fmt.Sprintf("ClientKeyExchange { Length: %d }", len(m.ExchangeData))
}

func ParseClientKeyExchange(data []byte) (*ClientKeyExchangeMessage, error) {
	if len(data) == 0 {
		return &ClientKeyExchangeMessage{RawData: data}, nil
	}
	if len(data) < 2 {
		return nil, errors.New("ClientKeyExchange too short")
	}
	length := binary.BigEndian.Uint16(data[:2])
	if int(length)+2 != len(data) {
		return nil, errors.New("invalid ClientKeyExchange length")
	}
	return &ClientKeyExchangeMessage{
		ExchangeData: data[2:],
		RawData:      data,
	}, nil
}

type FinishedMessage struct {
	VerifyData []byte
	RawData    []byte
}

func NewFinishedMessage(verifyData []byte) *FinishedMessage {
	return &FinishedMessage{VerifyData: verifyData}
}

func (m *FinishedMessage) Type() HandshakeType {
	return HandshakeTypeFinished
}

func (m *FinishedMessage) Serialize() ([]byte, error) {
	return m.VerifyData, nil
}

func (m *FinishedMessage) String() string {
	return fmt.Sprintf("Finished { VerifyData: %s... }", hex.EncodeToString(m.VerifyData[:min(8, len(m.VerifyData))]))
}

func ParseFinished(data []byte) (*FinishedMessage, error) {
	return &FinishedMessage{VerifyData: data, RawData: data}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ChangeCipherSpecMessage struct {
	Type uint8
}

func NewChangeCipherSpecMessage() *ChangeCipherSpecMessage {
	return &ChangeCipherSpecMessage{Type: 1}
}

func (m *ChangeCipherSpecMessage) Serialize() []byte {
	return []byte{m.Type}
}

func (m *ChangeCipherSpecMessage) String() string {
	return "ChangeCipherSpec"
}

func ParseChangeCipherSpec(data []byte) (*ChangeCipherSpecMessage, error) {
	if len(data) != 1 || data[0] != 1 {
		return nil, errors.New("invalid ChangeCipherSpec message")
	}
	return &ChangeCipherSpecMessage{Type: data[0]}, nil
}
