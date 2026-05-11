package idtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"time"
)

type Validator struct {
	config     *Config
	nonceCache *NonceCache
}

func NewValidator(config *Config, nonceCache *NonceCache) *Validator {
	if nonceCache == nil {
		nonceCache = NewNonceCache()
	}
	return &Validator{
		config:     config,
		nonceCache: nonceCache,
	}
}

func (v *Validator) UpdateConfig(config *Config) {
	v.config = config
}

func (v *Validator) GetConfig() *Config {
	return v.config
}

func (v *Validator) AddNonce(nonce string) {
	v.nonceCache.Add(nonce)
}

func (v *Validator) Validate(token string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  false,
		Claims: nil,
		Errors: []error{},
	}

	if v.config == nil {
		return result, ErrInvalidFormat
	}

	header, claims, signature, signingInput, err := Parse(token)
	if err != nil {
		return result, err
	}
	result.Claims = claims

	if !strings.EqualFold(header.Alg, "HS256") {
		result.Errors = append(result.Errors, ErrInvalidAlgorithm)
	}

	expectedSignature := v.sign(signingInput)
	if !hmac.Equal(signature, expectedSignature) {
		result.Errors = append(result.Errors, ErrInvalidSignature)
	}

	now := time.Now()
	iat := time.Unix(claims.Iat, 0)
	exp := time.Unix(claims.Exp, 0)

	if iat.After(now.Add(ClockSkew)) {
		result.Errors = append(result.Errors, ErrTokenUsedInFuture)
	}

	if !exp.After(now.Add(-ClockSkew)) {
		result.Errors = append(result.Errors, ErrTokenExpired)
	}

	if claims.Iss != v.config.Issuer {
		result.Errors = append(result.Errors, ErrInvalidIssuer)
	}

	audiences, err := claims.GetAudience()
	if err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		found := false
		for _, aud := range audiences {
			if aud == v.config.ClientID {
				found = true
				break
			}
		}
		if !found {
			result.Errors = append(result.Errors, ErrInvalidAudience)
		}
	}

	if claims.Nonce != "" {
		if !v.nonceCache.Validate(claims.Nonce) {
			result.Errors = append(result.Errors, ErrInvalidNonce)
		} else {
			v.nonceCache.Remove(claims.Nonce)
		}
	}

	if claims.Acr != "" {
		if claims.Acr != "urn:mfa" && claims.Acr != "urn:password" {
			result.Errors = append(result.Errors, ErrInvalidAcr)
		}
	}

	if len(result.Errors) == 0 {
		result.Valid = true
	}

	return result, nil
}

func (v *Validator) sign(input string) []byte {
	mac := hmac.New(sha256.New, v.config.Secret)
	mac.Write([]byte(input))
	return mac.Sum(nil)
}

func (v *Validator) Sign(input string) string {
	mac := hmac.New(sha256.New, v.config.Secret)
	mac.Write([]byte(input))
	signature := mac.Sum(nil)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(signature)
}
