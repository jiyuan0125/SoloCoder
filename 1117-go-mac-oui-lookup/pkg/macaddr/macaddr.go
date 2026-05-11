package macaddr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	multicastBit = 0x01
	localAdminBit = 0x02
)

type MAC [6]byte

type Format int

const (
	FormatColon Format = iota
	FormatDash
	FormatDot
)

func (f Format) String() string {
	switch f {
	case FormatColon:
		return "colon"
	case FormatDash:
		return "dash"
	case FormatDot:
		return "dot"
	default:
		return "colon"
	}
}

func ParseFormat(s string) (MAC, error) {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	if strings.Contains(s, ":") {
		return parseColonDash(s, ":")
	} else if strings.Contains(s, "-") {
		return parseColonDash(s, "-")
	} else if strings.Contains(s, ".") {
		return parseDot(s)
	} else {
		return parsePlain(s)
	}
}

func parseColonDash(s string, sep string) (MAC, error) {
	parts := strings.Split(s, sep)
	if len(parts) != 6 {
		return MAC{}, errors.New("invalid MAC address format")
	}

	var mac MAC
	for i, part := range parts {
		if len(part) != 2 {
			return MAC{}, errors.New("invalid MAC address format")
		}
		b, err := hex.DecodeString(part)
		if err != nil {
			return MAC{}, fmt.Errorf("invalid hex: %v", err)
		}
		mac[i] = b[0]
	}
	return mac, nil
}

func parseDot(s string) (MAC, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return MAC{}, errors.New("invalid MAC address format")
	}

	var mac MAC
	idx := 0
	for _, part := range parts {
		if len(part) != 4 {
			return MAC{}, errors.New("invalid MAC address format")
		}
		b, err := hex.DecodeString(part)
		if err != nil {
			return MAC{}, fmt.Errorf("invalid hex: %v", err)
		}
		mac[idx] = b[0]
		mac[idx+1] = b[1]
		idx += 2
	}
	return mac, nil
}

func parsePlain(s string) (MAC, error) {
	if len(s) != 12 {
		return MAC{}, errors.New("invalid MAC address format")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return MAC{}, fmt.Errorf("invalid hex: %v", err)
	}
	var mac MAC
	copy(mac[:], b)
	return mac, nil
}

func (m MAC) String() string {
	return m.Format(FormatColon)
}

func (m MAC) Format(format Format) string {
	switch format {
	case FormatDash:
		return fmt.Sprintf(
			"%02x-%02x-%02x-%02x-%02x-%02x",
			m[0], m[1], m[2], m[3], m[4], m[5],
		)
	case FormatDot:
		return fmt.Sprintf(
			"%02x%02x.%02x%02x.%02x%02x",
			m[0], m[1], m[2], m[3], m[4], m[5],
		)
	default:
		return fmt.Sprintf(
			"%02x:%02x:%02x:%02x:%02x:%02x",
			m[0], m[1], m[2], m[3], m[4], m[5],
		)
	}
}

func (m MAC) IsUnicast() bool {
	return m[0]&multicastBit == 0
}

func (m MAC) IsMulticast() bool {
	return m[0]&multicastBit != 0
}

func (m MAC) IsGloballyUnique() bool {
	return m[0]&localAdminBit == 0
}

func (m MAC) IsLocallyAdministered() bool {
	return m[0]&localAdminBit != 0
}

func (m MAC) IsZero() bool {
	return m == MAC{}
}

func (m MAC) IsBroadcast() bool {
	return m == MAC{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
}

func (m MAC) OUI() [3]byte {
	return [3]byte{m[0], m[1], m[2]}
}

func (m MAC) Bytes() []byte {
	return m[:]
}
