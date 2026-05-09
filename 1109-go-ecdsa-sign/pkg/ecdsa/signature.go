package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/big"
	"sync"
)

const (
	DefaultMaxKCacheSize = 10000
)

type Signer struct {
	mu       sync.RWMutex
	key      *ecdsa.PrivateKey
	curve    elliptic.Curve
	curveName CurveName
	kCache   map[string]struct{}
	maxCache int
	signCount int64
}

func NewSigner(curveName CurveName, maxCacheSize int) (*Signer, error) {
	curve, err := GetCurve(curveName)
	if err != nil {
		return nil, err
	}

	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	if maxCacheSize <= 0 {
		maxCacheSize = DefaultMaxKCacheSize
	}

	return &Signer{
		key:       key,
		curve:     curve,
		curveName: curveName,
		kCache:    make(map[string]struct{}),
		maxCache:  maxCacheSize,
	}, nil
}

func (s *Signer) GenerateKey(curveName CurveName) error {
	curve, err := GetCurve(curveName)
	if err != nil {
		return err
	}

	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.key = key
	s.curve = curve
	s.curveName = curveName
	s.kCache = make(map[string]struct{})
	s.signCount = 0
	
	return nil
}

func (s *Signer) PublicKey() *ecdsa.PublicKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.key == nil {
		return nil
	}
	return &s.key.PublicKey
}

func (s *Signer) CurveName() CurveName {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.curveName
}

func (s *Signer) SignCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.signCount
}

func (s *Signer) KCacheSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.kCache)
}

func (s *Signer) Sign(message []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.key == nil {
		return nil, errors.New("no key available")
	}

	digest := hashMessage(s.curve, message)

	r, sVal, k, err := signWithKCheck(rand.Reader, s.key, digest, s.kCache, s.maxCache)
	if err != nil {
		return nil, err
	}

	kHex := hex.EncodeToString(k.Bytes())
	s.kCache[kHex] = struct{}{}
	s.signCount++

	if len(s.kCache) > s.maxCache {
		clearHalfCache(s.kCache)
	}

	return EncodeDERSignature(r, sVal)
}

func (s *Signer) Verify(message, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
	if pubKey == nil {
		return false, errors.New("nil public key")
	}
	return Verify(message, signature, pubKey)
}

func Verify(message, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
	if pubKey == nil {
		return false, errors.New("nil public key")
	}

	r, sVal, err := DecodeDERSignature(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	n := pubKey.Params().N
	halfN := new(big.Int).Rsh(n, 1)

	if sVal.Cmp(halfN) > 0 {
		sVal = new(big.Int).Sub(n, sVal)
	}

	if r.Sign() <= 0 || sVal.Sign() <= 0 {
		return false, errors.New("invalid signature values")
	}
	if r.Cmp(n) >= 0 || sVal.Cmp(n) >= 0 {
		return false, errors.New("signature values out of range")
	}

	digest := hashMessage(pubKey.Curve, message)
	return ecdsa.Verify(pubKey, digest, r, sVal), nil
}

func signWithKCheck(randReader io.Reader, key *ecdsa.PrivateKey, digest []byte, kCache map[string]struct{}, maxCache int) (*big.Int, *big.Int, *big.Int, error) {
	curve := key.Curve
	n := curve.Params().N
	halfN := new(big.Int).Rsh(n, 1)

	for attempts := 0; attempts < 100; attempts++ {
		k, err := generateK(randReader, n)
		if err != nil {
			return nil, nil, nil, err
		}

		kHex := hex.EncodeToString(k.Bytes())
		if _, exists := kCache[kHex]; exists {
			continue
		}

		r, sVal, err := signUsingK(key, digest, k)
		if err != nil {
			return nil, nil, nil, err
		}

		if sVal.Cmp(halfN) > 0 {
			sVal = new(big.Int).Sub(n, sVal)
		}

		return r, sVal, k, nil
	}

	return nil, nil, nil, errors.New("failed to find unique k value after multiple attempts")
}

func generateK(randReader io.Reader, n *big.Int) (*big.Int, error) {
	one := big.NewInt(1)
	for {
		k, err := rand.Int(randReader, new(big.Int).Sub(n, one))
		if err != nil {
			return nil, err
		}
		k.Add(k, one)
		if k.Sign() > 0 && k.Cmp(n) < 0 {
			return k, nil
		}
	}
}

func signUsingK(key *ecdsa.PrivateKey, hash []byte, k *big.Int) (*big.Int, *big.Int, error) {
	curve := key.Curve
	params := curve.Params()
	n := params.N

	kInv := new(big.Int).ModInverse(k, n)
	if kInv == nil {
		return nil, nil, errors.New("failed to compute k inverse")
	}

	r, _ := curve.ScalarBaseMult(k.Bytes())
	r.Mod(r, n)
	if r.Sign() == 0 {
		return nil, nil, errors.New("r is zero")
	}

	e := hashToInt(hash, n)
	s := new(big.Int).Mul(key.D, r)
	s.Add(s, e)
	s.Mul(s, kInv)
	s.Mod(s, n)
	if s.Sign() == 0 {
		return nil, nil, errors.New("s is zero")
	}

	return r, s, nil
}

func hashToInt(hash []byte, n *big.Int) *big.Int {
	orderBits := n.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	e := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		e.Rsh(e, uint(excess))
	}
	return e
}

func hashMessage(curve elliptic.Curve, message []byte) []byte {
	var h hash.Hash
	bitSize := curve.Params().BitSize
	switch {
	case bitSize <= 256:
		h = sha256.New()
	case bitSize <= 384:
		h = sha512.New384()
	default:
		h = sha512.New()
	}
	h.Write(message)
	return h.Sum(nil)
}

func clearHalfCache(cache map[string]struct{}) {
	half := len(cache) / 2
	count := 0
	for k := range cache {
		if count >= half {
			break
		}
		delete(cache, k)
		count++
	}
}
