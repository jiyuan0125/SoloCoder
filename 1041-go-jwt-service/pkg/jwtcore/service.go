package jwtcore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenRevoked      = errors.New("token revoked")
	ErrInvalidAlgorithm  = errors.New("invalid algorithm")
	ErrInvalidTokenType  = errors.New("invalid token type")
	ErrSigningMethod     = errors.New("unexpected signing method")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessClaims *CustomClaims
	RefreshClaims *CustomClaims
}

type Service struct {
	config    *Config
	blacklist *Blacklist
}

func NewService(config *Config, blacklist *Blacklist) *Service {
	return &Service{
		config:    config,
		blacklist: blacklist,
	}
}

func (s *Service) IssueTokenPair(userID string, customClaims map[string]any) (*TokenPair, error) {
	now := time.Now()

	accessTokenID, err := generateTokenID()
	if err != nil {
		return nil, err
	}

	refreshTokenID, err := generateTokenID()
	if err != nil {
		return nil, err
	}

	accessClaims := &CustomClaims{
		TokenType: AccessToken,
		TokenID:   accessTokenID,
		UserID:    userID,
		Claims:    customClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        accessTokenID,
		},
	}

	refreshClaims := &CustomClaims{
		TokenType: RefreshToken,
		TokenID:   refreshTokenID,
		UserID:    userID,
		Claims:    customClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        refreshTokenID,
		},
	}

	accessToken, err := s.signToken(accessClaims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.signToken(refreshClaims)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessClaims:  accessClaims,
		RefreshClaims: refreshClaims,
	}, nil
}

func (s *Service) signToken(claims *CustomClaims) (string, error) {
	var token *jwt.Token

	switch s.config.Algorithm {
	case HS256:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(s.config.HS256Secret))
	case RS256:
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(s.config.RSAPrivateKey)
	default:
		return "", ErrInvalidAlgorithm
	}
}

func (s *Service) ValidateToken(tokenString string, expectedType TokenType) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		switch s.config.Algorithm {
		case HS256:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrSigningMethod
			}
			return []byte(s.config.HS256Secret), nil
		case RS256:
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, ErrSigningMethod
			}
			return s.config.RSAPublicKey, nil
		default:
			return nil, ErrInvalidAlgorithm
		}
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	if s.blacklist.IsRevoked(claims.TokenID) {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	return s.ValidateToken(tokenString, AccessToken)
}

func (s *Service) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	return s.ValidateToken(tokenString, RefreshToken)
}

func (s *Service) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	oldRefreshClaims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	if s.blacklist.IsRevoked(oldRefreshClaims.TokenID) {
		s.blacklist.RevokeUserTokens(oldRefreshClaims.UserID, time.Now().Add(s.config.RefreshTokenDuration))
		return nil, fmt.Errorf("%w: refresh token compromised, all user tokens revoked", ErrTokenRevoked)
	}

	s.blacklist.Revoke(oldRefreshClaims.TokenID, oldRefreshClaims.ExpiresAt.Time)

	newPair, err := s.IssueTokenPair(oldRefreshClaims.UserID, oldRefreshClaims.Claims)
	if err != nil {
		return nil, err
	}

	return newPair, nil
}

func (s *Service) RevokeToken(tokenID string, expiresAt time.Time) {
	s.blacklist.Revoke(tokenID, expiresAt)
}

func (s *Service) RevokeUserTokens(userID string, expiresAt time.Time) {
	s.blacklist.RevokeUserTokens(userID, expiresAt)
}

func generateTokenID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func LoadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key8, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, err
		}
		return key8.(*rsa.PrivateKey), nil
	}
	return key, nil
}

func LoadRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		key8, err2 := x509.ParsePKIXPublicKey(block.Bytes)
		if err2 != nil {
			return nil, err
		}
		return key8.(*rsa.PublicKey), nil
	}
	return key, nil
}
