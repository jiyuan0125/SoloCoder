package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"marketplace/common"
	"time"
)

type UserService struct {
	store *Store
}

func NewUserService(store *Store) *UserService {
	return &UserService{store: store}
}

func (u *UserService) Register(username string, role common.UserRole) (*common.User, error) {
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if role == "" {
		role = common.UserRoleUser
	}
	user := &common.User{
		ID:        generateID(),
		Username:  username,
		Role:      role,
		CreatedAt: time.Now(),
	}
	u.store.SaveUser(user)
	return user, nil
}

func (u *UserService) Get(id string) (*common.User, error) {
	user, ok := u.store.GetUser(id)
	if !ok {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

func (u *UserService) IsAdmin(id string) bool {
	user, ok := u.store.GetUser(id)
	if !ok {
		return false
	}
	return user.Role == common.UserRoleAdmin
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
