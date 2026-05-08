package jwtcore

import (
	"crypto/rsa"
	"os"
	"time"
)

type Algorithm string

const (
	HS256 Algorithm = "HS256"
	RS256 Algorithm = "RS256"
)

type Config struct {
	Algorithm Algorithm

	HS256Secret string

	RSAPrivateKey *rsa.PrivateKey
	RSAPublicKey  *rsa.PublicKey

	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	Issuer string
}

func NewHS256Config(secret string, issuer string) *Config {
	return &Config{
		Algorithm:          HS256,
		HS256Secret:  secret,
		Issuer:           issuer,
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	}
}

func NewRS256Config(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer string) *Config {
	return &Config{
		Algorithm:         RS256,
		RSAPrivateKey:  privateKey,
		RSAPublicKey:   publicKey,
		Issuer:          issuer,
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	}
}

func (c *Config) SetAccessTokenDuration(d time.Duration) {
	c.AccessTokenDuration = d
}

func (c *Config) SetRefreshTokenDuration(d time.Duration) {
	c.RefreshTokenDuration = d
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
