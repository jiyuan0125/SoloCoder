package iplocation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidIPFormat = errors.New("invalid IP format")
	ErrInvalidIPRange  = errors.New("invalid IP range")
	ErrRangeOverlap    = errors.New("IP range overlaps with existing range")
	ErrSpecialAddress  = errors.New("special address (0.0.0.0 or 255.255.255.255)")
	ErrNotFound        = errors.New("location not found")
)

func IPToUint32(ipStr string) (uint32, error) {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return 0, ErrInvalidIPFormat
	}

	var result uint32
	for _, part := range parts {
		num, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return 0, ErrInvalidIPFormat
		}
		result = result<<8 | uint32(num)
	}

	return result, nil
}

func Uint32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(ip>>24)&0xFF,
		(ip>>16)&0xFF,
		(ip>>8)&0xFF,
		ip&0xFF,
	)
}

func IsPrivateIP(ipStr string) (bool, error) {
	ip, err := IPToUint32(ipStr)
	if err != nil {
		return false, err
	}

	if (ip >= 0x0A000000 && ip <= 0x0AFFFFFF) ||
		(ip >= 0xAC100000 && ip <= 0xAC1FFFFF) ||
		(ip >= 0xC0A80000 && ip <= 0xC0A8FFFF) ||
		(ip >= 0x7F000000 && ip <= 0x7FFFFFFF) {
		return true, nil
	}
	return false, nil
}

func GetIPType(ipStr string) (IPType, error) {
	ip, err := IPToUint32(ipStr)
	if err != nil {
		return IPTypeInvalid, err
	}

	if ip == 0x00000000 || ip == 0xFFFFFFFF {
		return IPTypeSpecial, nil
	}

	firstByte := (ip >> 24) & 0xFF
	switch {
	case firstByte == 0 || firstByte == 255:
		return IPTypeSpecial, nil
	case firstByte <= 126:
		return IPTypeA, nil
	case firstByte >= 128 && firstByte <= 191:
		return IPTypeB, nil
	case firstByte >= 192 && firstByte <= 223:
		return IPTypeC, nil
	case firstByte >= 224 && firstByte <= 239:
		return IPTypeD, nil
	case firstByte >= 240:
		return IPTypeE, nil
	}
	return IPTypeInvalid, nil
}
