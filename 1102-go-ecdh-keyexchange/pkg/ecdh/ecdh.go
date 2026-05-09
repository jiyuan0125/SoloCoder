package ecdh

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

type Curve string

const (
	CurveP256 Curve = "P-256"
	CurveP384 Curve = "P-384"
	CurveP521 Curve = "P-521"
)

var (
	ErrUnsupportedCurve = errors.New("unsupported curve")
	ErrInvalidPublicKey = errors.New("invalid public key")
	ErrCurveMismatch    = errors.New("curve mismatch between claimed and actual public key")
	ErrNilPrivateKey    = errors.New("nil private key")
)

type KeyPair struct {
	Curve     Curve
	Private   *ecdh.PrivateKey
	PublicKey []byte
}

type SessionManager struct {
	sessions map[string]*KeyPair
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*KeyPair),
	}
}

func (sm *SessionManager) Add(sessionID string, kp *KeyPair) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[sessionID] = kp
}

func (sm *SessionManager) Get(sessionID string) (*KeyPair, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	kp, ok := sm.sessions[sessionID]
	return kp, ok
}

func (sm *SessionManager) Remove(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

func curveToGoCurve(c Curve) ecdh.Curve {
	switch c {
	case CurveP256:
		return ecdh.P256()
	case CurveP384:
		return ecdh.P384()
	case CurveP521:
		return ecdh.P521()
	default:
		return nil
	}
}

func goCurveNameToCurve(c ecdh.Curve) Curve {
	switch c {
	case ecdh.P256():
		return CurveP256
	case ecdh.P384():
		return CurveP384
	case ecdh.P521():
		return CurveP521
	default:
		return ""
	}
}

func ValidateCurve(c Curve) error {
	if curveToGoCurve(c) == nil {
		return ErrUnsupportedCurve
	}
	return nil
}

func GenerateKeyPair(curve Curve) (*KeyPair, error) {
	c := curveToGoCurve(curve)
	if c == nil {
		return nil, ErrUnsupportedCurve
	}

	priv, err := c.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	pubBytes := priv.PublicKey().Bytes()

	return &KeyPair{
		Curve:     curve,
		Private:   priv,
		PublicKey: pubBytes,
	}, nil
}

func ParsePublicKey(curve Curve, pubBytes []byte) (*ecdh.PublicKey, error) {
	c := curveToGoCurve(curve)
	if c == nil {
		return nil, ErrUnsupportedCurve
	}

	if len(pubBytes) == 0 {
		return nil, ErrInvalidPublicKey
	}

	pub, err := c.NewPublicKey(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPublicKey, err)
	}

	actualCurve := goCurveNameToCurve(pub.Curve())
	if actualCurve != curve {
		return nil, ErrCurveMismatch
	}

	return pub, nil
}

func ComputeSharedSecret(kp *KeyPair, peerPub *ecdh.PublicKey) (string, error) {
	if kp == nil || kp.Private == nil {
		return "", ErrNilPrivateKey
	}

	raw, err := kp.Private.ECDH(peerPub)
	if err != nil {
		return "", fmt.Errorf("failed to compute shared secret: %w", err)
	}

	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:]), nil
}

func ComputeSharedSecretFromBytes(kp *KeyPair, curve Curve, peerPubBytes []byte) (string, error) {
	if kp.Curve != curve {
		return "", ErrCurveMismatch
	}

	peerPub, err := ParsePublicKey(curve, peerPubBytes)
	if err != nil {
		return "", err
	}

	return ComputeSharedSecret(kp, peerPub)
}
