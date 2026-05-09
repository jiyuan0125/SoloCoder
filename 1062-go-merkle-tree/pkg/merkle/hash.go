package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

const BlockSize = 64 * 1024

func Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func HashHex(data []byte) string {
	return hex.EncodeToString(Hash(data))
}

func EmptyHash() []byte {
	return Hash(nil)
}

func EmptyHashHex() string {
	return HashHex(nil)
}

func ConcatAndHash(left, right []byte) []byte {
	combined := make([]byte, 0, len(left)+len(right))
	combined = append(combined, left...)
	combined = append(combined, right...)
	return Hash(combined)
}
