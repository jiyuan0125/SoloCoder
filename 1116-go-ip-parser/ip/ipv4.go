package ip

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ParseIPv4(s string) (*IPv4Address, error) {
	var parts []uint32
	var err error

	parts, err = parseIPv4Decimal(s)
	if err == nil {
		return buildIPv4Address(parts), nil
	}

	parts, err = parseIPv4Hex(s)
	if err == nil {
		return buildIPv4Address(parts), nil
	}

	parts, err = parseIPv4Octal(s)
	if err == nil {
		return buildIPv4Address(parts), nil
	}

	return nil, fmt.Errorf("invalid IPv4 address: %s", s)
}

func parseIPv4Decimal(s string) ([]uint32, error) {
	partsStr := strings.Split(s, ".")
	if len(partsStr) < 1 || len(partsStr) > 4 {
		return nil, errors.New("invalid number of parts")
	}

	parts := make([]uint32, 0, 4)
	for _, partStr := range partsStr {
		if strings.Contains(partStr, "x") || strings.Contains(partStr, "X") {
			return nil, errors.New("hexadecimal format")
		}
		if len(partStr) > 1 && partStr[0] == '0' {
			return nil, errors.New("leading zeros in decimal")
		}

		part, err := strconv.ParseUint(partStr, 10, 32)
		if err != nil {
			return nil, err
		}

		if part > 255 {
			return nil, errors.New("decimal part out of range")
		}

		parts = append(parts, uint32(part))
	}

	return parts, nil
}

func parseIPv4Hex(s string) ([]uint32, error) {
	partsStr := strings.Split(s, ".")
	if len(partsStr) < 1 || len(partsStr) > 4 {
		return nil, errors.New("invalid number of parts")
	}

	hasHex := false
	parts := make([]uint32, 0, 4)

	for _, partStr := range partsStr {
		if strings.HasPrefix(partStr, "0x") || strings.HasPrefix(partStr, "0X") {
			hasHex = true
			partStr = partStr[2:]

			if len(partStr) > 2 {
				return nil, errors.New("hex part too long")
			}

			part, err := strconv.ParseUint(partStr, 16, 32)
			if err != nil {
				return nil, err
			}

			parts = append(parts, uint32(part))
		} else {
			part, err := strconv.ParseUint(partStr, 10, 32)
			if err != nil {
				return nil, err
			}

			if part > 255 {
				return nil, errors.New("part out of range")
			}

			parts = append(parts, uint32(part))
		}
	}

	if !hasHex {
		return nil, errors.New("no hex parts")
	}

	return parts, nil
}

func parseIPv4Octal(s string) ([]uint32, error) {
	partsStr := strings.Split(s, ".")
	if len(partsStr) < 1 || len(partsStr) > 4 {
		return nil, errors.New("invalid number of parts")
	}

	hasOctal := false
	parts := make([]uint32, 0, 4)

	for _, partStr := range partsStr {
		if len(partStr) > 1 && partStr[0] == '0' {
			hasOctal = true

			part, err := strconv.ParseUint(partStr, 8, 32)
			if err != nil {
				return nil, err
			}

			if part > 255 {
				return nil, errors.New("octal part out of range")
			}

			parts = append(parts, uint32(part))
		} else {
			part, err := strconv.ParseUint(partStr, 10, 32)
			if err != nil {
				return nil, err
			}

			if part > 255 {
				return nil, errors.New("part out of range")
			}

			parts = append(parts, uint32(part))
		}
	}

	if !hasOctal {
		return nil, errors.New("no octal parts")
	}

	return parts, nil
}

func buildIPv4Address(parts []uint32) *IPv4Address {
	addr := &IPv4Address{}
	
	switch len(parts) {
	case 1:
		addr.bytes[3] = byte(parts[0] & 0xff)
		addr.bytes[2] = byte((parts[0] >> 8) & 0xff)
		addr.bytes[1] = byte((parts[0] >> 16) & 0xff)
		addr.bytes[0] = byte((parts[0] >> 24) & 0xff)
	case 2:
		addr.bytes[0] = byte(parts[0] & 0xff)
		addr.bytes[3] = byte(parts[1] & 0xff)
		addr.bytes[2] = byte((parts[1] >> 8) & 0xff)
		addr.bytes[1] = byte((parts[1] >> 16) & 0xff)
	case 3:
		addr.bytes[0] = byte(parts[0] & 0xff)
		addr.bytes[1] = byte(parts[1] & 0xff)
		addr.bytes[3] = byte(parts[2] & 0xff)
		addr.bytes[2] = byte((parts[2] >> 8) & 0xff)
	case 4:
		for i, part := range parts {
			addr.bytes[i] = byte(part & 0xff)
		}
	}

	return addr
}

func (a *IPv4Address) Version() IPVersion {
	return IPv4
}

func (a *IPv4Address) ToBytes() []byte {
	result := make([]byte, 4)
	copy(result, a.bytes[:])
	return result
}

func (a *IPv4Address) ToInt64() uint32 {
	return uint32(a.bytes[0])<<24 | uint32(a.bytes[1])<<16 | uint32(a.bytes[2])<<8 | uint32(a.bytes[3])
}

func (a *IPv4Address) ToInt128() [2]uint64 {
	return [2]uint64{0, uint64(a.ToInt64())}
}

func (a *IPv4Address) IsLoopback() bool {
	return a.bytes[0] == 127
}

func (a *IPv4Address) IsPrivate() bool {
	if a.bytes[0] == 10 {
		return true
	}
	if a.bytes[0] == 172 && a.bytes[1] >= 16 && a.bytes[1] <= 31 {
		return true
	}
	if a.bytes[0] == 192 && a.bytes[1] == 168 {
		return true
	}
	return false
}

func (a *IPv4Address) IsMulticast() bool {
	return a.bytes[0] >= 224 && a.bytes[0] <= 239
}

func (a *IPv4Address) IsBroadcast() bool {
	return a.bytes[0] == 255 && a.bytes[1] == 255 && a.bytes[2] == 255 && a.bytes[3] == 255
}

func (a *IPv4Address) IsLinkLocal() bool {
	return a.bytes[0] == 169 && a.bytes[1] == 254
}

func (a *IPv4Address) IsAnycast() bool {
	return false
}

func (a *IPv4Address) IsUnspecified() bool {
	return a.bytes[0] == 0 && a.bytes[1] == 0 && a.bytes[2] == 0 && a.bytes[3] == 0
}

func (a *IPv4Address) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", a.bytes[0], a.bytes[1], a.bytes[2], a.bytes[3])
}

func (a *IPv4Address) Class() IPv4Class {
	firstByte := a.bytes[0]
	
	if firstByte&0x80 == 0 {
		return IPv4ClassA
	}
	if firstByte&0xC0 == 0x80 {
		return IPv4ClassB
	}
	if firstByte&0xE0 == 0xC0 {
		return IPv4ClassC
	}
	if firstByte&0xF0 == 0xE0 {
		return IPv4ClassD
	}
	if firstByte&0xF0 == 0xF0 {
		return IPv4ClassE
	}
	return IPv4ClassUnknown
}

func (a *IPv4Address) GetIPv4MappedIPv6() *IPv6Address {
	result := &IPv6Address{
		mappingType: IPv6MappingIPv4,
	}
	result.bytes[10] = 0xff
	result.bytes[11] = 0xff
	result.bytes[12] = a.bytes[0]
	result.bytes[13] = a.bytes[1]
	result.bytes[14] = a.bytes[2]
	result.bytes[15] = a.bytes[3]
	return result
}
