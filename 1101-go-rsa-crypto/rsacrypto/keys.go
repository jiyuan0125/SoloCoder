package rsacrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

var (
	ErrUnsupportedKeySize  = errors.New("unsupported key size, must be 2048 or 4096")
	ErrInvalidPublicKey    = errors.New("invalid public key")
	ErrInvalidPrivateKey   = errors.New("invalid private key")
	ErrKeyGenerationFailed = errors.New("key generation failed")
)

const (
	pemTypePublicKey  = "PUBLIC KEY"
	pemTypePrivateKey = "RSA PRIVATE KEY"
)

func GenerateKeyPair(keySize int) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	if keySize != 2048 && keySize != 4096 {
		return nil, nil, ErrUnsupportedKeySize
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrKeyGenerationFailed, err)
	}
	return &privateKey.PublicKey, privateKey, nil
}

func ExportPublicKeyPEM(pub *rsa.PublicKey) (string, error) {
	if pub == nil {
		return "", ErrInvalidPublicKey
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	block := &pem.Block{
		Type:  pemTypePublicKey,
		Bytes: pubBytes,
	}
	return string(pem.EncodeToMemory(block)), nil
}

func ExportPrivateKeyPEM(priv *rsa.PrivateKey) (string, error) {
	if priv == nil {
		return "", ErrInvalidPrivateKey
	}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	block := &pem.Block{
		Type:  pemTypePrivateKey,
		Bytes: privBytes,
	}
	return string(pem.EncodeToMemory(block)), nil
}

func ImportPublicKeyPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, ErrInvalidPublicKey
	}
	if block.Type != pemTypePublicKey {
		return nil, fmt.Errorf("unexpected PEM type: %s, expected %s", block.Type, pemTypePublicKey)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidPublicKey
	}
	return rsaPub, nil
}

func ImportPrivateKeyPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, ErrInvalidPrivateKey
	}
	if block.Type != pemTypePrivateKey {
		return nil, fmt.Errorf("unexpected PEM type: %s, expected %s", block.Type, pemTypePrivateKey)
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return priv, nil
}

func MaxPlaintextSize(keySize int, padding string) int {
	switch padding {
	case "pkcs1v15":
		return keySize/8 - 11
	case "oaep":
		return keySize/8 - 2*32 - 2
	default:
		return keySize/8 - 11
	}
}
