package cryptolib

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"io"
)

const (
	Ed25519PublicKeySize  = ed25519.PublicKeySize
	Ed25519PrivateKeySize = ed25519.PrivateKeySize
	Ed25519SignatureSize  = ed25519.SignatureSize
)

func GenerateKey() (publicKey, privateKey string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	publicKey = base64.StdEncoding.EncodeToString(pub)
	privateKey = base64.StdEncoding.EncodeToString(priv)
	return publicKey, privateKey, nil
}

func Sign(privateKeyStr, message string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", ErrInvalidBase64
	}
	if len(privBytes) != Ed25519PrivateKeySize {
		return "", ErrInvalidKeyLength
	}
	priv := ed25519.PrivateKey(privBytes)
	sig := ed25519.Sign(priv, []byte(message))
	return base64.StdEncoding.EncodeToString(sig), nil
}

func Verify(publicKeyStr, message, signatureStr string) error {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return ErrInvalidBase64
	}
	if len(pubBytes) != Ed25519PublicKeySize {
		return ErrInvalidKeyLength
	}
	sigBytes, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return ErrInvalidBase64
	}
	if len(sigBytes) != Ed25519SignatureSize {
		return ErrInvalidKeyLength
	}
	pub := ed25519.PublicKey(pubBytes)
	if !ed25519.Verify(pub, []byte(message), sigBytes) {
		return ErrInvalidSignature
	}
	return nil
}

func SignReader(privateKeyStr string, r io.Reader) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", ErrInvalidBase64
	}
	if len(privBytes) != Ed25519PrivateKeySize {
		return "", ErrInvalidKeyLength
	}
	priv := ed25519.PrivateKey(privBytes)
	content, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, content)
	return base64.StdEncoding.EncodeToString(sig), nil
}

func VerifyReader(publicKeyStr string, r io.Reader, signatureStr string) error {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return ErrInvalidBase64
	}
	if len(pubBytes) != Ed25519PublicKeySize {
		return ErrInvalidKeyLength
	}
	sigBytes, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return ErrInvalidBase64
	}
	if len(sigBytes) != Ed25519SignatureSize {
		return ErrInvalidKeyLength
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	pub := ed25519.PublicKey(pubBytes)
	if !ed25519.Verify(pub, content, sigBytes) {
		return ErrInvalidSignature
	}
	return nil
}
