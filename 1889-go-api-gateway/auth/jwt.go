package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTValidator struct {
	secret []byte
}

func NewJWTValidator(secret []byte) *JWTValidator {
	return &JWTValidator{secret: secret}
}

func (v *JWTValidator) Validate(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return nil, errors.New("token expired")
			}
		}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
