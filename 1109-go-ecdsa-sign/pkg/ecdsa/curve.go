package ecdsa

import (
	"crypto/elliptic"
	"errors"
	"fmt"
)

type CurveName string

const (
	CurveP256 CurveName = "P-256"
	CurveP384 CurveName = "P-384"
	CurveP521 CurveName = "P-521"
)

func ParseCurveName(name string) (CurveName, error) {
	switch name {
	case string(CurveP256), "P256", "p256", "prime256v1", "secp256r1":
		return CurveP256, nil
	case string(CurveP384), "P384", "p384", "secp384r1":
		return CurveP384, nil
	case string(CurveP521), "P521", "p521", "secp521r1":
		return CurveP521, nil
	default:
		return "", fmt.Errorf("unsupported curve: %s", name)
	}
}

func GetCurve(name CurveName) (elliptic.Curve, error) {
	switch name {
	case CurveP256:
		return elliptic.P256(), nil
	case CurveP384:
		return elliptic.P384(), nil
	case CurveP521:
		return elliptic.P521(), nil
	default:
		return nil, errors.New("unsupported curve")
	}
}
