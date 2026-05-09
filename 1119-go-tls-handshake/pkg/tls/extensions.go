package tls

const (
	ExtTypeSNI               uint16 = 0x0000
	ExtTypeSupportedGroups   uint16 = 0x000A
	ExtTypeSignatureAlgs     uint16 = 0x000D
	ExtTypeALPN              uint16 = 0x0010
)

var Extensions = map[uint16]string{
	ExtTypeSNI:             "server_name",
	ExtTypeSupportedGroups: "supported_groups",
	ExtTypeSignatureAlgs:   "signature_algorithms",
	ExtTypeALPN:            "application_layer_protocol_negotiation",
}

var ExtensionNames = map[string]uint16{
	"server_name":             ExtTypeSNI,
	"supported_groups":        ExtTypeSupportedGroups,
	"signature_algorithms":    ExtTypeSignatureAlgs,
	"application_layer_protocol_negotiation": ExtTypeALPN,
}

var ExpectedExtensionOrder = []uint16{
	ExtTypeSNI,
	ExtTypeSupportedGroups,
	ExtTypeSignatureAlgs,
	ExtTypeALPN,
}

type Extension struct {
	Type    uint16
	Data    interface{}
}

type SNIData struct {
	HostName string
}

type SupportedGroupsData struct {
	Groups []string
}

type SignatureAlgsData struct {
	Algorithms []string
}

type ALPNData struct {
	Protocols []string
}

func ValidateExtensionOrder(exts []Extension) error {
	if len(exts) == 0 {
		return nil
	}

	if exts[0].Type == ExtTypeSNI {
		expectedIdx := 0
		for _, ext := range exts {
			for j := expectedIdx; j < len(ExpectedExtensionOrder); j++ {
				if ext.Type == ExpectedExtensionOrder[j] {
					expectedIdx = j + 1
					break
				}
			}
		}
	} else {
		return NewTLSProtocolError("SNI extension must be the first extension when present")
	}

	return nil
}
