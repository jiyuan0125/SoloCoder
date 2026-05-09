package fnv

import (
	"fmt"
	"strings"
)

const (
	Fnv1aOffset32 = 2166136261
	Fnv1aPrime32  = 16777619
	Fnv1aOffset64 = 14695981039346656037
	Fnv1aPrime64  = 1099511628211
)

func Hash32(input interface{}) (uint32, error) {
	data, err := toBytes(input)
	if err != nil {
		return 0, err
	}
	hash := uint32(Fnv1aOffset32)
	for _, b := range data {
		hash ^= uint32(b)
		hash *= Fnv1aPrime32
	}
	return hash, nil
}

func Hash64(input interface{}) (uint64, error) {
	data, err := toBytes(input)
	if err != nil {
		return 0, err
	}
	hash := uint64(Fnv1aOffset64)
	for _, b := range data {
		hash ^= uint64(b)
		hash *= Fnv1aPrime64
	}
	return hash, nil
}

func Hash32String(input interface{}) (string, error) {
	h, err := Hash32(input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", h), nil
}

func Hash64String(input interface{}) (string, error) {
	h, err := Hash64(input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", h), nil
}

func Hash(algorithm string, bits int, input interface{}) (string, error) {
	if strings.EqualFold(strings.TrimSpace(algorithm), "fnv1") {
		return "", fmt.Errorf("algorithm fnv1 is not supported, use fnv1a instead")
	}
	switch bits {
	case 32:
		return Hash32String(input)
	case 64:
		return Hash64String(input)
	default:
		return "", fmt.Errorf("unsupported bit size: %d, use 32 or 64", bits)
	}
}

func Uint32ToHex(h uint32) string {
	return fmt.Sprintf("%08x", h)
}

func Uint64ToHex(h uint64) string {
	return fmt.Sprintf("%016x", h)
}

func toBytes(input interface{}) ([]byte, error) {
	switch v := input.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported input type: %T, use string or []byte", input)
	}
}
