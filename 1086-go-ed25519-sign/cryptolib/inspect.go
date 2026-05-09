package cryptolib

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"filippo.io/edwards25519"
)

type KeyType string

const (
	KeyTypeEd25519Public  KeyType = "Ed25519Public"
	KeyTypeEd25519Private KeyType = "Ed25519Private"
	KeyTypeX25519Public   KeyType = "X25519Public"
	KeyTypeX25519Private  KeyType = "X25519Private"
	KeyTypeUnknown        KeyType = "Unknown"
)

type KeyInfo struct {
	Type        KeyType
	EncodedLen  int
	DecodedLen  int
	Fingerprint string
}

func InspectKey(keyStr string) (*KeyInfo, error) {
	encodedLen := len(keyStr)
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, ErrInvalidBase64
	}
	decodedLen := len(keyBytes)
	keyType := KeyTypeUnknown
	switch decodedLen {
	case Ed25519PrivateKeySize:
		keyType = KeyTypeEd25519Private
	case X25519KeySize:
		if isClampedX25519Private(keyBytes) {
			keyType = KeyTypeX25519Private
		} else if isValidEd25519Public(keyBytes) {
			keyType = KeyTypeEd25519Public
		} else {
			keyType = KeyTypeX25519Public
		}
	default:
		keyType = KeyTypeUnknown
	}
	hash := sha256.Sum256(keyBytes)
	fingerprint := hex.EncodeToString(hash[:16])
	return &KeyInfo{
		Type:        keyType,
		EncodedLen:  encodedLen,
		DecodedLen:  decodedLen,
		Fingerprint: fingerprint,
	}, nil
}

func isClampedX25519Private(keyBytes []byte) bool {
	if len(keyBytes) != X25519KeySize {
		return false
	}
	return keyBytes[0]&248 == keyBytes[0] &&
		keyBytes[31]&127 == keyBytes[31] &&
		keyBytes[31]&64 == 64
}

func isValidEd25519Public(keyBytes []byte) bool {
	if len(keyBytes) != Ed25519PublicKeySize {
		return false
	}
	_, err := new(edwards25519.Point).SetBytes(keyBytes)
	return err == nil
}

func (k *KeyInfo) String() string {
	return fmt.Sprintf("Type: %s, EncodedLen: %d, DecodedLen: %d, Fingerprint: %s",
		k.Type, k.EncodedLen, k.DecodedLen, k.Fingerprint)
}
