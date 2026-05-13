package archiver

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

const (
	aesKeySize = 32
	nonceSize  = 12
)

func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func encryptFile(inputPath, outputPath, password string) error {
	key := deriveKey(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, nonceSize)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	data, err := io.ReadAll(input)
	if err != nil {
		return err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)

	return os.WriteFile(outputPath, encrypted, 0644)
}

func decryptFile(inputPath, outputPath, password string) error {
	key := deriveKey(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	encryptedData, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	if len(encryptedData) < nonceSize {
		return fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: incorrect password or corrupted data")
	}

	return os.WriteFile(outputPath, plaintext, 0644)
}
