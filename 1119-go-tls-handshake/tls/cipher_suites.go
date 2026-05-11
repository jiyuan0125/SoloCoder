package tls

import (
	"encoding/binary"
	"errors"
)

type CipherSuiteID uint16

const (
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256  CipherSuiteID = 0xC02F
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384  CipherSuiteID = 0xC030
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 CipherSuiteID = 0xC02B
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 CipherSuiteID = 0xC02C
	TLS_RSA_WITH_AES_128_GCM_SHA256        CipherSuiteID = 0x009C
	TLS_RSA_WITH_AES_256_GCM_SHA384        CipherSuiteID = 0x009D
	TLS_RSA_WITH_AES_128_CBC_SHA           CipherSuiteID = 0x002F
	TLS_RSA_WITH_AES_256_CBC_SHA           CipherSuiteID = 0x0035
	TLS_RSA_WITH_AES_128_CBC_SHA256        CipherSuiteID = 0x003C
	TLS_RSA_WITH_AES_256_CBC_SHA256        CipherSuiteID = 0x003D
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA     CipherSuiteID = 0xC013
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA     CipherSuiteID = 0xC014
)

type CipherSuite struct {
	ID          CipherSuiteID
	Name        string
	KeyExchange string
	Encryption  string
	MAC         string
}

var CipherSuiteRegistry = map[CipherSuiteID]CipherSuite{
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256: {
		ID:          TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		Name:        "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		KeyExchange: "ECDHE_RSA",
		Encryption:  "AES_128_GCM",
		MAC:         "SHA256",
	},
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384: {
		ID:          TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		Name:        "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		KeyExchange: "ECDHE_RSA",
		Encryption:  "AES_256_GCM",
		MAC:         "SHA384",
	},
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: {
		ID:          TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		Name:        "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		KeyExchange: "ECDHE_ECDSA",
		Encryption:  "AES_128_GCM",
		MAC:         "SHA256",
	},
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: {
		ID:          TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		Name:        "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		KeyExchange: "ECDHE_ECDSA",
		Encryption:  "AES_256_GCM",
		MAC:         "SHA384",
	},
	TLS_RSA_WITH_AES_128_GCM_SHA256: {
		ID:          TLS_RSA_WITH_AES_128_GCM_SHA256,
		Name:        "TLS_RSA_WITH_AES_128_GCM_SHA256",
		KeyExchange: "RSA",
		Encryption:  "AES_128_GCM",
		MAC:         "SHA256",
	},
	TLS_RSA_WITH_AES_256_GCM_SHA384: {
		ID:          TLS_RSA_WITH_AES_256_GCM_SHA384,
		Name:        "TLS_RSA_WITH_AES_256_GCM_SHA384",
		KeyExchange: "RSA",
		Encryption:  "AES_256_GCM",
		MAC:         "SHA384",
	},
	TLS_RSA_WITH_AES_128_CBC_SHA: {
		ID:          TLS_RSA_WITH_AES_128_CBC_SHA,
		Name:        "TLS_RSA_WITH_AES_128_CBC_SHA",
		KeyExchange: "RSA",
		Encryption:  "AES_128_CBC",
		MAC:         "SHA",
	},
	TLS_RSA_WITH_AES_256_CBC_SHA: {
		ID:          TLS_RSA_WITH_AES_256_CBC_SHA,
		Name:        "TLS_RSA_WITH_AES_256_CBC_SHA",
		KeyExchange: "RSA",
		Encryption:  "AES_256_CBC",
		MAC:         "SHA",
	},
	TLS_RSA_WITH_AES_128_CBC_SHA256: {
		ID:          TLS_RSA_WITH_AES_128_CBC_SHA256,
		Name:        "TLS_RSA_WITH_AES_128_CBC_SHA256",
		KeyExchange: "RSA",
		Encryption:  "AES_128_CBC",
		MAC:         "SHA256",
	},
	TLS_RSA_WITH_AES_256_CBC_SHA256: {
		ID:          TLS_RSA_WITH_AES_256_CBC_SHA256,
		Name:        "TLS_RSA_WITH_AES_256_CBC_SHA256",
		KeyExchange: "RSA",
		Encryption:  "AES_256_CBC",
		MAC:         "SHA256",
	},
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA: {
		ID:          TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		Name:        "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		KeyExchange: "ECDHE_RSA",
		Encryption:  "AES_128_CBC",
		MAC:         "SHA",
	},
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA: {
		ID:          TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		Name:        "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		KeyExchange: "ECDHE_RSA",
		Encryption:  "AES_256_CBC",
		MAC:         "SHA",
	},
}

var DefaultClientCipherSuites = []CipherSuiteID{
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	TLS_RSA_WITH_AES_128_GCM_SHA256,
	TLS_RSA_WITH_AES_256_GCM_SHA384,
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	TLS_RSA_WITH_AES_128_CBC_SHA,
	TLS_RSA_WITH_AES_256_CBC_SHA,
}

var DefaultServerCipherSuites = []CipherSuiteID{
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	TLS_RSA_WITH_AES_128_GCM_SHA256,
	TLS_RSA_WITH_AES_256_GCM_SHA384,
	TLS_RSA_WITH_AES_128_CBC_SHA256,
	TLS_RSA_WITH_AES_256_CBC_SHA256,
}

func (id CipherSuiteID) GetInfo() (CipherSuite, bool) {
	cs, ok := CipherSuiteRegistry[id]
	return cs, ok
}

func NegotiateCipherSuite(clientSuites, serverSuites []CipherSuiteID) (CipherSuiteID, error) {
	for _, clientID := range clientSuites {
		for _, serverID := range serverSuites {
			if clientID == serverID {
				return clientID, nil
			}
		}
	}
	return 0, errors.New("no common cipher suite found")
}

func ParseCipherSuites(data []byte) ([]CipherSuiteID, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid cipher suites data")
	}
	length := binary.BigEndian.Uint16(data[:2])
	if length%2 != 0 || int(length)+2 != len(data) {
		return nil, errors.New("invalid cipher suites length")
	}
	suites := make([]CipherSuiteID, 0, length/2)
	for i := 2; i < len(data); i += 2 {
		suites = append(suites, CipherSuiteID(binary.BigEndian.Uint16(data[i:i+2])))
	}
	return suites, nil
}

func SerializeCipherSuites(suites []CipherSuiteID) []byte {
	length := len(suites) * 2
	data := make([]byte, 2+length)
	binary.BigEndian.PutUint16(data[:2], uint16(length))
	for i, id := range suites {
		binary.BigEndian.PutUint16(data[2+i*2:], uint16(id))
	}
	return data
}

func IsValidCipherSuite(id CipherSuiteID) bool {
	_, ok := CipherSuiteRegistry[id]
	return ok
}
