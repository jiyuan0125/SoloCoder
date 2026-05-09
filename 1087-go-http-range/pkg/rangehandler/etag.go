package rangehandler

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateETag(content []byte) string {
	hash := sha256.Sum256(content)
	hexHash := hex.EncodeToString(hash[:])
	return hexHash[:32]
}
