package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltLength       = 32
	NonceLength      = 12
	KeyLength        = 32
	TagLength        = 16
	DefaultIterations = 100000
)

var (
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrCiphertextTooShort = errors.New("ciphertext is too short to be valid")
	Iterations = DefaultIterations
)

func SetIterations(n int) {
	if n > 0 {
		Iterations = n
	}
}

func DeriveKey(password string, salt []byte, iterations int, keyLength int) []byte {
	return pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)
}

func generateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func Encrypt(plaintext, password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	salt, err := generateRandomBytes(SaltLength)
	if err != nil {
		return "", err
	}

	key := DeriveKey(password, salt, Iterations, KeyLength)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce, err := generateRandomBytes(NonceLength)
	if err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	result := make([]byte, 0, SaltLength+NonceLength+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func Decrypt(ciphertextBase64, password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	minLength := SaltLength + NonceLength + TagLength
	if len(ciphertext) < minLength {
		return "", ErrCiphertextTooShort
	}

	salt := ciphertext[:SaltLength]
	nonce := ciphertext[SaltLength : SaltLength+NonceLength]
	encryptedData := ciphertext[SaltLength+NonceLength:]

	key := DeriveKey(password, salt, Iterations, KeyLength)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
