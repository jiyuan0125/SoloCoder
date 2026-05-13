package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"jwt-middleware/config"
	"jwt-middleware/jwt"
)

type KeyEntry struct {
	ID          string
	Signer      jwt.Signer
	CreatedAt   time.Time
	ExpiresAt   time.Time
	IsActive    bool
}

type KeyStore struct {
	mu       sync.RWMutex
	keys     map[string]*KeyEntry
	current  *KeyEntry
	previous *KeyEntry
	transitionEnd time.Time
}

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrNoActiveKey = errors.New("no active key available")
)

func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys:     make(map[string]*KeyEntry),
	}
}

func GenerateHMACKey() []byte {
	key := make([]byte, 64)
	rand.Read(key)
	return key
}

func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return priv, &priv.PublicKey, nil
}

func GenerateKeyID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("key-%x", b)
}

func (ks *KeyStore) AddHMACKey(alg string, transition bool) (string, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	key := GenerateHMACKey()
	signer, err := jwt.NewHMACSigner(key, alg)
	if err != nil {
		return "", err
	}
	kid := GenerateKeyID()
	entry := &KeyEntry{
		ID:        kid,
		Signer:    signer,
		CreatedAt: time.Now(),
		IsActive:  true,
	}
	if ks.current != nil {
		ks.previous = ks.current
		if transition {
			ks.transitionEnd = time.Now().Add(config.KeyTransitionPeriod)
		} else {
			ks.previous.IsActive = false
		}
	}
	ks.keys[kid] = entry
	ks.current = entry
	return kid, nil
}

func (ks *KeyStore) AddRSAKey(alg string, transition bool) (string, *rsa.PrivateKey, *rsa.PublicKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	priv, pub, err := GenerateRSAKeyPair()
	if err != nil {
		return "", nil, nil, err
	}
	signer, err := jwt.NewRSASigner(priv, pub, alg)
	if err != nil {
		return "", nil, nil, err
	}
	kid := GenerateKeyID()
	entry := &KeyEntry{
		ID:        kid,
		Signer:    signer,
		CreatedAt: time.Now(),
		IsActive:  true,
	}
	if ks.current != nil {
		ks.previous = ks.current
		if transition {
			ks.transitionEnd = time.Now().Add(config.KeyTransitionPeriod)
		} else {
			ks.previous.IsActive = false
		}
	}
	ks.keys[kid] = entry
	ks.current = entry
	return kid, priv, pub, nil
}

func (ks *KeyStore) GetCurrentKey() (*KeyEntry, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	if ks.current == nil {
		return nil, ErrNoActiveKey
	}
	return ks.current, nil
}

func (ks *KeyStore) GetKey(kid string) (*KeyEntry, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	entry, exists := ks.keys[kid]
	if !exists {
		return nil, ErrKeyNotFound
	}
	return entry, nil
}

func (ks *KeyStore) GetActiveKeys() []*KeyEntry {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	var active []*KeyEntry
	now := time.Now()
	if ks.current != nil {
		active = append(active, ks.current)
	}
	if ks.previous != nil {
		if !now.After(ks.transitionEnd) {
			active = append(active, ks.previous)
		}
	}
	return active
}

func (ks *KeyStore) IsInTransition() bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	if ks.previous == nil {
		return false
	}
	return !time.Now().After(ks.transitionEnd)
}

func (ks *KeyStore) RotateKey(transition bool) (string, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if ks.current == nil {
		return "", ErrNoActiveKey
	}
	var alg string
	switch s := ks.current.Signer.(type) {
	case *jwt.HMACSigner:
		alg = s.Alg()
		key := GenerateHMACKey()
		signer, err := jwt.NewHMACSigner(key, alg)
		if err != nil {
			return "", err
		}
		kid := GenerateKeyID()
		entry := &KeyEntry{
			ID:        kid,
			Signer:    signer,
			CreatedAt: time.Now(),
			IsActive:  true,
		}
		ks.previous = ks.current
		if transition {
			ks.transitionEnd = time.Now().Add(config.KeyTransitionPeriod)
		} else {
			ks.previous.IsActive = false
		}
		ks.keys[kid] = entry
		ks.current = entry
		return kid, nil
	case *jwt.RSASigner:
		alg = s.Alg()
		priv, pub, err := GenerateRSAKeyPair()
		if err != nil {
			return "", err
		}
		signer, err := jwt.NewRSASigner(priv, pub, alg)
		if err != nil {
			return "", err
		}
		kid := GenerateKeyID()
		entry := &KeyEntry{
			ID:        kid,
			Signer:    signer,
			CreatedAt: time.Now(),
			IsActive:  true,
		}
		ks.previous = ks.current
		if transition {
			ks.transitionEnd = time.Now().Add(config.KeyTransitionPeriod)
		} else {
			ks.previous.IsActive = false
		}
		ks.keys[kid] = entry
		ks.current = entry
		return kid, nil
	default:
		return "", errors.New("unknown signer type")
	}
}

func EncodeRSAPrivateKey(priv *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}))
}

func EncodeRSAPublicKey(pub *rsa.PublicKey) string {
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	}))
}
