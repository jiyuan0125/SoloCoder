package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func GenerateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func Sign(header Header, claims Claims, signer Signer) (string, error) {
	headerStr, err := header.Encode()
	if err != nil {
		return "", err
	}
	claimsStr, err := claims.Encode()
	if err != nil {
		return "", err
	}
	signingInput := headerStr + "." + claimsStr
	signature, err := signer.Sign([]byte(signingInput))
	if err != nil {
		return "", err
	}
	signatureStr := Base64URLEncode(signature)
	return signingInput + "." + signatureStr, nil
}

func Parse(tokenString string) (Header, Claims, string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return Header{}, Claims{}, "", &InvalidTokenError{Reason: "invalid token format"}
	}
	var header Header
	if err := header.Decode(parts[0]); err != nil {
		return Header{}, Claims{}, "", &InvalidTokenError{Reason: "invalid header: " + err.Error()}
	}
	var claims Claims
	if err := claims.Decode(parts[1]); err != nil {
		return Header{}, Claims{}, "", &InvalidTokenError{Reason: "invalid claims: " + err.Error()}
	}
	_, err := Base64URLDecode(parts[2])
	if err != nil {
		return Header{}, Claims{}, "", &InvalidTokenError{Reason: "invalid signature encoding: " + err.Error()}
	}
	signingInput := parts[0] + "." + parts[1]
	return header, claims, signingInput, nil
}

func Verify(tokenString string, signer Signer, expectedIssuer, expectedAudience string) (Header, Claims, error) {
	header, claims, signingInput, err := Parse(tokenString)
	if err != nil {
		return Header{}, Claims{}, err
	}
	signature, err := Base64URLDecode(strings.Split(tokenString, ".")[2])
	if err != nil {
		return Header{}, Claims{}, &InvalidTokenError{Reason: "invalid signature encoding"}
	}
	if header.Alg != signer.Alg() {
		return Header{}, Claims{}, &InvalidSignatureError{}
	}
	if !signer.Verify([]byte(signingInput), signature) {
		return Header{}, Claims{}, &InvalidSignatureError{}
	}
	if claims.IsExpired() {
		return header, claims, &ExpiredTokenError{ExpiresAt: claims.ExpiresAt()}
	}
	if expectedIssuer != "" && claims.Iss != expectedIssuer {
		return header, claims, &InvalidIssuerError{Got: claims.Iss, Expected: expectedIssuer}
	}
	if expectedAudience != "" && claims.Aud != expectedAudience {
		return header, claims, &InvalidAudienceError{Got: claims.Aud, Expected: expectedAudience}
	}
	return header, claims, nil
}

func CreateToken(userID, role string, tokenType string, duration time.Duration, issuer, audience string, kid string, signer Signer) (string, Claims, error) {
	jti, err := GenerateJTI()
	if err != nil {
		return "", Claims{}, err
	}
	now := time.Now()
	claims := Claims{
		Sub:    userID,
		UserID: userID,
		Role:   role,
		Jti:    jti,
		Iss:    issuer,
		Aud:    audience,
		Iat:    now.Unix(),
		Exp:    now.Add(duration).Unix(),
		Type:   tokenType,
	}
	header := Header{
		Alg: signer.Alg(),
		Typ: "JWT",
		Kid: kid,
	}
	token, err := Sign(header, claims, signer)
	return token, claims, err
}

type InvalidTokenError struct {
	Reason string
}

func (e *InvalidTokenError) Error() string {
	return "invalid token: " + e.Reason
}

type InvalidSignatureError struct{}

func (e *InvalidSignatureError) Error() string {
	return "invalid signature"
}

type ExpiredTokenError struct {
	ExpiresAt time.Time
}

func (e *ExpiredTokenError) Error() string {
	return "token expired at: " + e.ExpiresAt.Format(time.RFC3339)
}

type InvalidIssuerError struct {
	Got      string
	Expected string
}

func (e *InvalidIssuerError) Error() string {
	return "invalid issuer: got " + e.Got + ", expected " + e.Expected
}

type InvalidAudienceError struct {
	Got      string
	Expected string
}

func (e *InvalidAudienceError) Error() string {
	return "invalid audience: got " + e.Got + ", expected " + e.Expected
}

var ErrInvalidToken = errors.New("invalid token")
var ErrInvalidSignature = errors.New("invalid signature")
var ErrTokenExpired = errors.New("token expired")
