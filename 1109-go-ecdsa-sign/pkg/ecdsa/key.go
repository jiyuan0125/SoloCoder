package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

func PublicKeyToPEM(pubKey *ecdsa.PublicKey) (string, error) {
	if pubKey == nil {
		return "", errors.New("nil public key")
	}
	
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}
	
	return string(pem.EncodeToMemory(pemBlock)), nil
}

func PublicKeyFromPEM(pemData string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("expected PUBLIC KEY block, got %s", block.Type)
	}
	
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	
	return ecdsaPub, nil
}

func PublicKeyToBase64(pubKey *ecdsa.PublicKey) (string, error) {
	if pubKey == nil {
		return "", errors.New("nil public key")
	}
	
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	
	return base64.StdEncoding.EncodeToString(derBytes), nil
}

func PublicKeyFromBase64(b64Data string) (*ecdsa.PublicKey, error) {
	derBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	
	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	
	return ecdsaPub, nil
}

func PublicKeyToRawXY(pubKey *ecdsa.PublicKey) (x, y *big.Int, curveName CurveName, err error) {
	if pubKey == nil {
		return nil, nil, "", errors.New("nil public key")
	}
	if pubKey.X == nil || pubKey.Y == nil {
		return nil, nil, "", errors.New("public key coordinates are nil")
	}
	
	var name CurveName
	switch pubKey.Curve {
	case elliptic.P256():
		name = CurveP256
	case elliptic.P384():
		name = CurveP384
	case elliptic.P521():
		name = CurveP521
	default:
		return nil, nil, "", errors.New("unsupported curve")
	}
	
	return pubKey.X, pubKey.Y, name, nil
}

func PublicKeyFromRawXY(x, y *big.Int, curveName CurveName) (*ecdsa.PublicKey, error) {
	if x == nil || y == nil {
		return nil, errors.New("nil coordinates")
	}
	
	curve, err := GetCurve(curveName)
	if err != nil {
		return nil, err
	}
	
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("point not on curve")
	}
	
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}
