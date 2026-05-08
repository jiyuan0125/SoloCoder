package jwtcore

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type CustomClaims struct {
	TokenType TokenType         `json:"token_type"`
	TokenID   string            `json:"jti"`
	UserID    string            `json:"user_id,omitempty"`
	Claims    map[string]any    `json:"claims,omitempty"`
	jwt.RegisteredClaims
}
