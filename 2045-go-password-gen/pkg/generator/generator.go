package generator

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

const (
	DefaultLength      = 16
	MinLength          = 4
	MaxBatch           = 10000
	DefaultSpecialChars = `!@#$%^&*()_+-=[]{}|;':",./<>?`
)

type PasswordType int

const (
	TypeNumeric PasswordType = iota
	TypeAlphanumeric
	TypeWithSpecial
	TypeReadable
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(length int, passwordType PasswordType, specialChars string) (string, error) {
	if length < MinLength {
		return "", errors.New("密码长度不能小于 4")
	}

	var charset string
	switch passwordType {
	case TypeNumeric:
		charset = "0123456789"
	case TypeAlphanumeric:
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	case TypeWithSpecial:
		if strings.TrimSpace(specialChars) == "" {
			return "", errors.New("字符集不能为空")
		}
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" + specialChars
	default:
		return "", errors.New("不支持的密码类型")
	}

	return g.generateFromCharset(length, charset)
}

func (g *Generator) generateFromCharset(length int, charset string) (string, error) {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[idx.Int64()]
	}

	return string(result), nil
}
