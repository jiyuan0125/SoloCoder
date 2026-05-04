package service

import (
	"errors"

	"feedback-collection/dao"
	"feedback-collection/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrTokenInvalid       = errors.New("invalid or expired token")
)

func Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	admin, err := dao.GetAdminByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if admin == nil {
		return nil, ErrInvalidCredentials
	}

	if admin.Password != req.Password {
		return nil, ErrInvalidCredentials
	}

	token := "admin_token_" + admin.Username

	return &models.LoginResponse{
		Token: token,
	}, nil
}

func ValidateToken(token string) (bool, error) {
	if len(token) < 12 || token[:12] != "admin_token_" {
		return false, ErrTokenInvalid
	}

	username := token[12:]
	admin, err := dao.GetAdminByUsername(username)
	if err != nil {
		return false, err
	}

	if admin == nil {
		return false, ErrTokenInvalid
	}

	return true, nil
}
