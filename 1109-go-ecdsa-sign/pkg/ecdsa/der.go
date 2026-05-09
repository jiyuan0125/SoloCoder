package ecdsa

import (
	"encoding/asn1"
	"errors"
	"math/big"
)

type DERSignature struct {
	R *big.Int
	S *big.Int
}

func EncodeDERSignature(r, s *big.Int) ([]byte, error) {
	if r == nil || s == nil {
		return nil, errors.New("invalid signature components")
	}
	return asn1.Marshal(DERSignature{R: r, S: s})
}

func DecodeDERSignature(data []byte) (*big.Int, *big.Int, error) {
	if len(data) == 0 {
		return nil, nil, errors.New("empty signature data")
	}
	
	var sig DERSignature
	rest, err := asn1.Unmarshal(data, &sig)
	if err != nil {
		return nil, nil, err
	}
	if len(rest) > 0 {
		return nil, nil, errors.New("unexpected trailing data in signature")
	}
	if sig.R == nil || sig.S == nil {
		return nil, nil, errors.New("invalid signature components")
	}
	if sig.R.Sign() <= 0 || sig.S.Sign() <= 0 {
		return nil, nil, errors.New("signature components must be positive")
	}
	
	return sig.R, sig.S, nil
}
