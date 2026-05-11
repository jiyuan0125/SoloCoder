package socks5

import (
	"fmt"
	"io"
	"net"
)

type AuthStore interface {
	Validate(username, password string) bool
	AddUser(username, password string) error
	RemoveUser(username string) error
	ListUsers() []string
}

type UserAuth struct {
	users map[string]string
}

func NewUserAuth() *UserAuth {
	return &UserAuth{
		users: make(map[string]string),
	}
}

func (a *UserAuth) Validate(username, password string) bool {
	expected, ok := a.users[username]
	return ok && expected == password
}

func (a *UserAuth) AddUser(username, password string) error {
	if len(username) == 0 || len(username) > 255 {
		return fmt.Errorf("username length must be 1-255")
	}
	if len(password) == 0 || len(password) > 255 {
		return fmt.Errorf("password length must be 1-255")
	}
	a.users[username] = password
	return nil
}

func (a *UserAuth) RemoveUser(username string) error {
	if _, ok := a.users[username]; !ok {
		return fmt.Errorf("user not found")
	}
	delete(a.users, username)
	return nil
}

func (a *UserAuth) ListUsers() []string {
	users := make([]string, 0, len(a.users))
	for u := range a.users {
		users = append(users, u)
	}
	return users
}

func (a *UserAuth) HasUsers() bool {
	return len(a.users) > 0
}

type NoAuth struct{}

func (a *NoAuth) Validate(username, password string) bool {
	return true
}

func (a *NoAuth) AddUser(username, password string) error {
	return nil
}

func (a *NoAuth) RemoveUser(username string) error {
	return nil
}

func (a *NoAuth) ListUsers() []string {
	return nil
}

func handleGreeting(conn net.Conn, authStore AuthStore, requireAuth bool) (bool, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return false, err
	}

	if buf[0] != Version {
		return false, fmt.Errorf("unsupported SOCKS version: %d", buf[0])
	}

	methodCount := int(buf[1])
	if methodCount == 0 {
		return false, fmt.Errorf("no authentication methods")
	}

	methods := make([]byte, methodCount)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return false, err
	}

	var selectedMethod byte
	if requireAuth {
		selectedMethod = MethodNoAcceptable
		for _, m := range methods {
			if m == MethodUserPassAuth {
				selectedMethod = MethodUserPassAuth
				break
			}
		}
	} else {
		for _, m := range methods {
			if m == MethodNoAuth {
				selectedMethod = MethodNoAuth
				break
			}
			if m == MethodUserPassAuth {
				selectedMethod = MethodUserPassAuth
				break
			}
		}
	}

	if selectedMethod == MethodNoAcceptable {
		_, err := conn.Write([]byte{Version, MethodNoAcceptable})
		return false, err
	}

	if _, err := conn.Write([]byte{Version, selectedMethod}); err != nil {
		return false, err
	}

	if selectedMethod == MethodUserPassAuth {
		return handleUserPassAuth(conn, authStore)
	}

	return true, nil
}

func handleUserPassAuth(conn net.Conn, authStore AuthStore) (bool, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return false, err
	}

	if buf[0] != AuthUserPassVersion {
		conn.Write([]byte{AuthUserPassVersion, AuthFailure})
		return false, fmt.Errorf("unsupported auth version: %d", buf[0])
	}

	usernameLen := int(buf[1])
	if usernameLen == 0 || usernameLen > 255 {
		conn.Write([]byte{AuthUserPassVersion, AuthFailure})
		return false, fmt.Errorf("invalid username length")
	}

	username := make([]byte, usernameLen)
	if _, err := io.ReadFull(conn, username); err != nil {
		return false, err
	}

	passLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, passLenBuf); err != nil {
		return false, err
	}

	passwordLen := int(passLenBuf[0])
	if passwordLen == 0 || passwordLen > 255 {
		conn.Write([]byte{AuthUserPassVersion, AuthFailure})
		return false, fmt.Errorf("invalid password length")
	}

	password := make([]byte, passwordLen)
	if _, err := io.ReadFull(conn, password); err != nil {
		return false, err
	}

	if authStore.Validate(string(username), string(password)) {
		_, err := conn.Write([]byte{AuthUserPassVersion, AuthSuccess})
		return true, err
	}

	_, err := conn.Write([]byte{AuthUserPassVersion, AuthFailure})
	return false, err
}
