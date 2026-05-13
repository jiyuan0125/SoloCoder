package service

import (
	"errors"

	"jwt-middleware/config"
	"jwt-middleware/jwt"
	"jwt-middleware/keys"
	"jwt-middleware/store"
)

type TokenService struct {
	keystore *keys.KeyStore
	database *store.Store
}

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrTokenBlacklisted  = errors.New("token blacklisted")
	ErrRefreshTokenReused = errors.New("refresh token already used")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrInvalidAudience = errors.New("invalid audience")
	ErrKeyNotFound    = errors.New("key not found")
	ErrTokenNotExpired  = errors.New("token not expired")
)

func NewTokenService(ks *keys.KeyStore) *TokenService {
	db, _ := store.NewStore(config.DBPath)
	return &TokenService{
		keystore: ks,
		database: db,
	}
}

func (s *TokenService) IssueTokens(userID, role string) (*jwt.TokenPair, jwt.Claims, jwt.Claims, error) {
	keyEntry, err := s.keystore.GetCurrentKey()
	if err != nil {
		return nil, jwt.Claims{}, jwt.Claims{}, err
	}
	accessToken, accessClaims, err := jwt.CreateToken(
		userID, role, "access",
		config.AccessTokenDuration,
		config.Issuer, config.Audience,
		keyEntry.ID,
		keyEntry.Signer,
	)
	if err != nil {
		return nil, jwt.Claims{}, jwt.Claims{}, err
	}
	refreshToken, refreshClaims, err := jwt.CreateToken(
		userID, role, "refresh",
		config.RefreshTokenDuration,
		config.Issuer, config.Audience,
		keyEntry.ID,
		keyEntry.Signer,
	)
	if err != nil {
		return nil, jwt.Claims{}, jwt.Claims{}, err
	}
	err = s.database.StoreRefreshToken(refreshClaims.Jti, userID, role, refreshClaims.Exp)
	if err != nil {
		return nil, jwt.Claims{}, jwt.Claims{}, err
	}
	s.database.IncrementTokensIssued(userID)
	return &jwt.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, accessClaims, refreshClaims, nil
}

func (s *TokenService) ValidateToken(tokenString string) (jwt.Header, jwt.Claims, error) {
	header, claims, signingInput, err := jwt.Parse(tokenString)
	if err != nil {
		return jwt.Header{}, jwt.Claims{}, &jwt.InvalidTokenError{Reason: err.Error()}
	}
	if claims.IsExpired() {
		return header, claims, &jwt.ExpiredTokenError{ExpiresAt: claims.ExpiresAt()}
	}
	if claims.Iss != config.Issuer {
		return jwt.Header{}, jwt.Claims{}, &jwt.InvalidIssuerError{Got: claims.Iss, Expected: config.Issuer}
	}
	if claims.Aud != config.Audience {
		return jwt.Header{}, jwt.Claims{}, &jwt.InvalidAudienceError{Got: claims.Aud, Expected: config.Audience}
	}
	activeKeys := s.keystore.GetActiveKeys()
	var verified bool
	for _, keyEntry := range activeKeys {
		if keyEntry.Signer.Alg() == header.Alg {
			signature, decodeErr := jwt.Base64URLDecode(tokenString[len(signingInput)+1:])
			if decodeErr == nil {
				if keyEntry.Signer.Verify([]byte(signingInput), signature) {
					verified = true
					break
				}
			}
		}
	}
	if !verified {
		return jwt.Header{}, jwt.Claims{}, jwt.ErrInvalidSignature
	}
	blacklisted, err := s.database.IsBlacklisted(claims.Jti)
	if err != nil {
		return jwt.Header{}, jwt.Claims{}, err
	}
	if blacklisted {
		return jwt.Header{}, jwt.Claims{}, ErrTokenBlacklisted
	}
	return header, claims, nil
}

func (s *TokenService) RefreshTokens(refreshTokenString string) (*jwt.TokenPair, error) {
	header, claims, err := s.ValidateToken(refreshTokenString)
	_ = header
	if err != nil {
		return nil, err
	}
	if claims.Type != "refresh" {
		return nil, &jwt.InvalidTokenError{Reason: "not a refresh token"}
	}
	if claims.IsExpired() {
		return nil, &jwt.ExpiredTokenError{ExpiresAt: claims.ExpiresAt()}
	}
	valid, err := s.database.UseRefreshToken(claims.Jti, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !valid {
		s.database.RevokeAllTokens(claims.UserID)
		return nil, ErrRefreshTokenReused
	}
	pair, _, _, err := s.IssueTokens(claims.UserID, claims.Role)
	if err != nil {
		return nil, err
	}
	s.database.IncrementRefreshes()
	return pair, nil
}

func (s *TokenService) RevokeToken(tokenString string) error {
	_, claims, _, err := jwt.Parse(tokenString)
	if err != nil {
		return err
	}
	return s.database.AddToBlacklist(claims.Jti, claims.UserID, claims.Exp)
}

func (s *TokenService) RevokeAllUserTokens(userID string) error {
	return s.database.RevokeAllTokens(userID)
}

func (s *TokenService) GetStats() (map[string]int64, error) {
	return s.database.GetStats()
}

func (s *TokenService) RotateKey(transition bool) (string, error) {
	return s.keystore.RotateKey(transition)
}

func (s *TokenService) AddHMACKey(alg string, transition bool) (string, error) {
	return s.keystore.AddHMACKey(alg, transition)
}

func (s *TokenService) AddRSAKey(alg string, transition bool) (string, error) {
	kid, _, _, err := s.keystore.AddRSAKey(alg, transition)
	return kid, err
}
