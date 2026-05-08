package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"jwt-service/pkg/jwtcore"
)

type ServerConfig struct {
	Port         int
	JWTConfig    *jwtcore.Config
}

func LoadServerConfig() (*ServerConfig, error) {
	portStr := jwtcore.GetEnvOrDefault("JWT_SERVER_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	algorithm := jwtcore.Algorithm(jwtcore.GetEnvOrDefault("JWT_ALGORITHM", "HS256"))
	issuer := jwtcore.GetEnvOrDefault("JWT_ISSUER", "jwt-service")

	var jwtConfig *jwtcore.Config

	switch algorithm {
	case jwtcore.HS256:
		secret := os.Getenv("JWT_HS256_SECRET")
		if secret == "" {
			return nil, fmt.Errorf("JWT_HS256_SECRET is required for HS256 algorithm")
		}
		jwtConfig = jwtcore.NewHS256Config(secret, issuer)

	case jwtcore.RS256:
		privateKeyPath := os.Getenv("JWT_RSA_PRIVATE_KEY")
		publicKeyPath := os.Getenv("JWT_RSA_PUBLIC_KEY")

		if privateKeyPath == "" {
			return nil, fmt.Errorf("JWT_RSA_PRIVATE_KEY is required for RS256 algorithm")
		}
		if publicKeyPath == "" {
			return nil, fmt.Errorf("JWT_RSA_PUBLIC_KEY is required for RS256 algorithm")
		}

		privateKey, err := jwtcore.LoadRSAPrivateKeyFromFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}

		publicKey, err := jwtcore.LoadRSAPublicKeyFromFile(publicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}

		jwtConfig = jwtcore.NewRS256Config(privateKey, publicKey, issuer)

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	if accessDurationStr := os.Getenv("JWT_ACCESS_TOKEN_DURATION"); accessDurationStr != "" {
		accessDuration, err := time.ParseDuration(accessDurationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid access token duration: %w", err)
		}
		jwtConfig.SetAccessTokenDuration(accessDuration)
	}

	if refreshDurationStr := os.Getenv("JWT_REFRESH_TOKEN_DURATION"); refreshDurationStr != "" {
		refreshDuration, err := time.ParseDuration(refreshDurationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid refresh token duration: %w", err)
		}
		jwtConfig.SetRefreshTokenDuration(refreshDuration)
	}

	return &ServerConfig{
		Port:      port,
		JWTConfig: jwtConfig,
	}, nil
}
