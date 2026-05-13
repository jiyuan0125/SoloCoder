package jwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

type Signer interface {
	Sign(data []byte) ([]byte, error)
	Verify(data, signature []byte) bool
	Alg() string
}

type HMACSigner struct {
	key    []byte
	hash   crypto.Hash
	alg    string
}

type RSASigner struct {
	priv   *rsa.PrivateKey
	pub    *rsa.PublicKey
	hash   crypto.Hash
	alg    string
}

func NewHMACSigner(key []byte, alg string) (*HMACSigner, error) {
	var h crypto.Hash
	switch alg {
	case "HS256":
		h = crypto.SHA256
	case "HS384":
		h = crypto.SHA384
	case "HS512":
		h = crypto.SHA512
	default:
		return nil, &InvalidAlgorithmError{Alg: alg}
	}
	return &HMACSigner{key: key, hash: h, alg: alg}, nil
}

func (s *HMACSigner) Sign(data []byte) ([]byte, error) {
	var h hash.Hash
	switch s.hash {
	case crypto.SHA256:
		h = hmac.New(sha256.New, s.key)
	case crypto.SHA384:
		h = hmac.New(sha512.New384, s.key)
	case crypto.SHA512:
		h = hmac.New(sha512.New, s.key)
	}
	h.Write(data)
	return h.Sum(nil), nil
}

func (s *HMACSigner) Verify(data, signature []byte) bool {
	expected, err := s.Sign(data)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, signature)
}

func (s *HMACSigner) Alg() string {
	return s.alg
}

func NewRSASigner(priv *rsa.PrivateKey, pub *rsa.PublicKey, alg string) (*RSASigner, error) {
	var h crypto.Hash
	switch alg {
	case "RS256":
		h = crypto.SHA256
	case "RS384":
		h = crypto.SHA384
	case "RS512":
		h = crypto.SHA512
	default:
		return nil, &InvalidAlgorithmError{Alg: alg}
	}
	return &RSASigner{priv: priv, pub: pub, hash: h, alg: alg}, nil
}

func (s *RSASigner) Sign(data []byte) ([]byte, error) {
	if s.priv == nil {
		return nil, &NoPrivateKeyError{}
	}
	var h crypto.Hash
	switch s.hash {
	case crypto.SHA256:
		h = crypto.SHA256
	case crypto.SHA384:
		h = crypto.SHA384
	case crypto.SHA512:
		h = crypto.SHA512
	}
	hasher := h.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, s.priv, h, hashed)
}

func (s *RSASigner) Verify(data, signature []byte) bool {
	if s.pub == nil {
		return false
	}
	var h crypto.Hash
	switch s.hash {
	case crypto.SHA256:
		h = crypto.SHA256
	case crypto.SHA384:
		h = crypto.SHA384
	case crypto.SHA512:
		h = crypto.SHA512
	}
	hasher := h.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)
	err := rsa.VerifyPKCS1v15(s.pub, h, hashed, signature)
	return err == nil
}

func (s *RSASigner) Alg() string {
	return s.alg
}

type InvalidAlgorithmError struct {
	Alg string
}

func (e *InvalidAlgorithmError) Error() string {
	return "invalid algorithm: " + e.Alg
}

type NoPrivateKeyError struct{}

func (e *NoPrivateKeyError) Error() string {
	return "no private key available for signing"
}
