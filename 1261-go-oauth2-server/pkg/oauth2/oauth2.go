package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) Authorize(clientID, redirectURI, responseType, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	if responseType != "code" {
		return "", errors.New("unsupported_response_type")
	}

	client, ok := s.store.GetClient(clientID)
	if !ok {
		return "", errors.New("invalid_client")
	}

	if !isRedirectURIMatch(redirectURI, client.RedirectURIs) {
		return "", errors.New("invalid_redirect_uri")
	}

	if codeChallenge == "" || codeChallengeMethod != "S256" {
		return "", errors.New("invalid_request: code_challenge and code_challenge_method=S256 required")
	}

	code := generateRandomString(32)
	authCode := &AuthorizationCode{
		Code:               code,
		ClientID:           clientID,
		RedirectURI:        redirectURI,
		Scope:              scope,
		CodeChallenge:      codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:          time.Now().Add(CodeExpiry),
		Used:               false,
	}

	s.store.StoreAuthCode(authCode)
	return code, nil
}

func (s *Service) Exchange(clientID, clientSecret, code, redirectURI, codeVerifier string) (*Token, error) {
	client, ok := s.store.GetClient(clientID)
	if !ok || client.Secret != clientSecret {
		return nil, errors.New("invalid_client")
	}

	authCode, ok := s.store.GetAuthCode(code)
	if !ok {
		return nil, errors.New("invalid_grant")
	}

	if authCode.Used || time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("invalid_grant: code expired or used")
	}

	if authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		return nil, errors.New("invalid_grant")
	}

	if !isValidCodeVerifier(codeVerifier) {
		return nil, errors.New("invalid_grant: invalid code_verifier")
	}

	if !VerifyCodeChallenge(codeVerifier, authCode.CodeChallenge) {
		return nil, errors.New("invalid_grant: code_verifier mismatch")
	}

	authCode.Used = true
	s.store.StoreAuthCode(authCode)

	token := s.generateToken(authCode.ClientID, authCode.Scope)
	s.store.StoreToken(token)

	return token, nil
}

func (s *Service) Refresh(refreshToken string) (*Token, error) {
	oldToken, ok := s.store.GetTokenByRefresh(refreshToken)
	if !ok {
		return nil, errors.New("invalid_grant")
	}

	if time.Now().After(oldToken.RefreshExpiry) {
		return nil, errors.New("invalid_grant: refresh_token expired")
	}

	s.store.DeleteToken(oldToken.AccessToken)

	newToken := s.generateToken(oldToken.ClientID, oldToken.Scope)
	s.store.StoreToken(newToken)

	return newToken, nil
}

func (s *Service) generateToken(clientID, scope string) *Token {
	now := time.Now()
	return &Token{
		AccessToken:   generateRandomString(32),
		RefreshToken:  generateRandomString(32),
		TokenType:     "Bearer",
		Scope:         scope,
		AccessExpiry:  now.Add(AccessExpiry),
		RefreshExpiry: now.Add(RefreshExpiry),
		ClientID:      clientID,
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func isRedirectURIMatch(uri string, registered []string) bool {
	for _, r := range registered {
		if r == uri {
			return true
		}
	}
	return false
}

func ParseBasicAuth(header string) (string, string, error) {
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", errors.New("missing Basic auth")
	}
	encoded := strings.TrimPrefix(header, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid Basic auth format")
	}
	return parts[0], parts[1], nil
}
