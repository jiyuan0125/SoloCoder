package cryptolib

import (
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/base64"

	"filippo.io/edwards25519"
)

const X25519KeySize = 32

func PrivateKeyToX25519(privateKeyStr string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", ErrInvalidBase64
	}
	if len(privBytes) != Ed25519PrivateKeySize {
		return "", ErrInvalidKeyLength
	}
	priv := ed25519.PrivateKey(privBytes)
	seed := priv.Seed()
	h := sha512.New()
	h.Write(seed)
	hashBytes := h.Sum(nil)
	x25519Priv := make([]byte, X25519KeySize)
	copy(x25519Priv, hashBytes[:X25519KeySize])
	x25519Priv[0] &= 248
	x25519Priv[31] &= 127
	x25519Priv[31] |= 64
	return base64.StdEncoding.EncodeToString(x25519Priv), nil
}

func PublicKeyToX25519(publicKeyStr string) (string, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return "", ErrInvalidBase64
	}
	if len(pubBytes) != Ed25519PublicKeySize {
		return "", ErrInvalidKeyLength
	}
	p, err := new(edwards25519.Point).SetBytes(pubBytes)
	if err != nil {
		return "", ErrKeyConversionFailed
	}
	x25519Pub := p.BytesMontgomery()
	return base64.StdEncoding.EncodeToString(x25519Pub), nil
}
