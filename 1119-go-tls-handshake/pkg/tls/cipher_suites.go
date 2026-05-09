package tls

type CipherSuite struct {
	ID          uint16
	Name        string
	Description string
}

var CipherSuites = map[string]CipherSuite{
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": {
		ID:          0xC02F,
		Name:        "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		Description: "ECDHE RSA key exchange with AES-128-GCM encryption and SHA-256 hash",
	},
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": {
		ID:          0xC030,
		Name:        "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		Description: "ECDHE RSA key exchange with AES-256-GCM encryption and SHA-384 hash",
	},
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": {
		ID:          0xC02B,
		Name:        "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		Description: "ECDHE ECDSA key exchange with AES-128-GCM encryption and SHA-256 hash",
	},
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": {
		ID:          0xC02C,
		Name:        "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		Description: "ECDHE ECDSA key exchange with AES-256-GCM encryption and SHA-384 hash",
	},
	"TLS_RSA_WITH_AES_256_CBC_SHA": {
		ID:          0x0035,
		Name:        "TLS_RSA_WITH_AES_256_CBC_SHA",
		Description: "RSA key exchange with AES-256-CBC encryption and SHA-1 hash",
	},
	"TLS_RSA_WITH_AES_128_CBC_SHA": {
		ID:          0x002F,
		Name:        "TLS_RSA_WITH_AES_128_CBC_SHA",
		Description: "RSA key exchange with AES-128-CBC encryption and SHA-1 hash",
	},
	"TLS_RSA_WITH_AES_256_CBC_SHA256": {
		ID:          0x003D,
		Name:        "TLS_RSA_WITH_AES_256_CBC_SHA256",
		Description: "RSA key exchange with AES-256-CBC encryption and SHA-256 hash",
	},
	"TLS_RSA_WITH_AES_128_GCM_SHA256": {
		ID:          0x009C,
		Name:        "TLS_RSA_WITH_AES_128_GCM_SHA256",
		Description: "RSA key exchange with AES-128-GCM encryption and SHA-256 hash",
	},
	"TLS_RSA_WITH_AES_256_GCM_SHA384": {
		ID:          0x009D,
		Name:        "TLS_RSA_WITH_AES_256_GCM_SHA384",
		Description: "RSA key exchange with AES-256-GCM encryption and SHA-384 hash",
	},
	"TLS_DHE_RSA_WITH_AES_128_GCM_SHA256": {
		ID:          0x009E,
		Name:        "TLS_DHE_RSA_WITH_AES_128_GCM_SHA256",
		Description: "DHE RSA key exchange with AES-128-GCM encryption and SHA-256 hash",
	},
	"TLS_DHE_RSA_WITH_AES_256_GCM_SHA384": {
		ID:          0x009F,
		Name:        "TLS_DHE_RSA_WITH_AES_256_GCM_SHA384",
		Description: "DHE RSA key exchange with AES-256-GCM encryption and SHA-384 hash",
	},
}

var CipherSuitesByID = make(map[uint16]CipherSuite)

func init() {
	for _, cs := range CipherSuites {
		CipherSuitesByID[cs.ID] = cs
	}
}

func GetCipherSuiteByName(name string) (CipherSuite, bool) {
	cs, ok := CipherSuites[name]
	return cs, ok
}

func GetCipherSuiteByID(id uint16) (CipherSuite, bool) {
	cs, ok := CipherSuitesByID[id]
	return cs, ok
}
