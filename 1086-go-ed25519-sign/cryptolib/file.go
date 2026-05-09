package cryptolib

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"os"
)

const bufferSize = 64 * 1024

func readAllStream(r io.Reader) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	tmp := make([]byte, bufferSize)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func SignFile(privateKeyStr, filePath string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", ErrInvalidBase64
	}
	if len(privBytes) != Ed25519PrivateKeySize {
		return "", ErrInvalidKeyLength
	}
	priv := ed25519.PrivateKey(privBytes)
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	content, err := readAllStream(f)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, content)
	return base64.StdEncoding.EncodeToString(sig), nil
}

func VerifyFile(publicKeyStr, filePath, signatureStr string) error {
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
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := readAllStream(f)
	if err != nil {
		return err
	}
	pub := ed25519.PublicKey(pubBytes)
	if !ed25519.Verify(pub, content, sigBytes) {
		return ErrInvalidSignature
	}
	return nil
}
