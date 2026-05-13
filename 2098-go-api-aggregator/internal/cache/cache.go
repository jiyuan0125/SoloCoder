package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"api-aggregator/internal/db"
)

func GenerateKey(prefix string, data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(append([]byte(prefix+":"), jsonData...))
	return hex.EncodeToString(hash[:]), nil
}

func Get(key string) ([]byte, error) {
	entry, err := db.GetCacheEntry(key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return entry.Data, nil
}

func Set(key string, data []byte, ttlSeconds int) error {
	if ttlSeconds <= 0 {
		return nil
	}
	return db.SetCacheEntry(key, data, ttlSeconds)
}

func GetOrCompute(key string, ttlSeconds int, compute func() ([]byte, error)) ([]byte, error) {
	if cached, err := Get(key); err == nil && cached != nil {
		return cached, nil
	}

	data, err := compute()
	if err != nil {
		return nil, err
	}

	if err := Set(key, data, ttlSeconds); err != nil {
		fmt.Printf("warning: failed to set cache: %v\n", err)
	}

	return data, nil
}

func Cleanup() error {
	return db.CleanExpiredCache()
}
