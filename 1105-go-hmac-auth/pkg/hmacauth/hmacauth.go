package hmacauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	timestampTolerance = 5 * time.Minute
	keyLength          = 32
)

type KeyStore struct {
	mu          sync.RWMutex
	currentKey  []byte
	currentVer  int
	previousKey []byte
	previousVer int
}

func NewKeyStore() (*KeyStore, error) {
	ks := &KeyStore{}
	key, err := generateKey()
	if err != nil {
		return nil, err
	}
	ks.currentKey = key
	ks.currentVer = 1
	return ks, nil
}

func generateKey() ([]byte, error) {
	key := make([]byte, keyLength)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

func (ks *KeyStore) RotateKey() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	newKey, err := generateKey()
	if err != nil {
		return err
	}

	ks.previousKey = ks.currentKey
	ks.previousVer = ks.currentVer
	ks.currentKey = newKey
	ks.currentVer = ks.previousVer + 1

	return nil
}

func (ks *KeyStore) GetCurrentKey() ([]byte, int) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	key := make([]byte, len(ks.currentKey))
	copy(key, ks.currentKey)
	return key, ks.currentVer
}

func (ks *KeyStore) GetAllKeys() map[int][]byte {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keys := make(map[int][]byte)
	currentCopy := make([]byte, len(ks.currentKey))
	copy(currentCopy, ks.currentKey)
	keys[ks.currentVer] = currentCopy

	if ks.previousKey != nil {
		prevCopy := make([]byte, len(ks.previousKey))
		copy(prevCopy, ks.previousKey)
		keys[ks.previousVer] = prevCopy
	}

	return keys
}

func BuildMessageToSign(method, path string, timestamp int64, body []byte) string {
	var builder strings.Builder
	builder.WriteString(strings.ToUpper(method))
	builder.WriteString("\n")
	builder.WriteString(path)
	builder.WriteString("\n")
	builder.WriteString(strconv.FormatInt(timestamp, 10))
	builder.WriteString("\n")
	builder.WriteString(string(body))
	return builder.String()
}

func ComputeHMAC(key []byte, message string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func Sign(key []byte, method, path string, timestamp int64, body []byte) string {
	message := BuildMessageToSign(method, path, timestamp, body)
	return ComputeHMAC(key, message)
}

func Verify(key []byte, method, path string, timestamp int64, body []byte, signature string) bool {
	now := time.Now().Unix()
	if now-timestamp > int64(timestampTolerance.Seconds()) {
		return false
	}
	if timestamp-now > int64(timestampTolerance.Seconds()) {
		return false
	}

	expected := Sign(key, method, path, timestamp, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func VerifyWithKeyStore(ks *KeyStore, method, path string, timestamp int64, body []byte, signature string) bool {
	ks.mu.RLock()
	currentKey := ks.currentKey
	previousKey := ks.previousKey
	ks.mu.RUnlock()

	if Verify(currentKey, method, path, timestamp, body, signature) {
		return true
	}

	if previousKey != nil {
		return Verify(previousKey, method, path, timestamp, body, signature)
	}

	return false
}
