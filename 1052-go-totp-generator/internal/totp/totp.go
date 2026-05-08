package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	DefaultDigits = 6
	DefaultPeriod = 30
	DefaultWindow = 1
)

func Generate(secret string) (string, error) {
	return GenerateAt(secret, time.Now().Unix(), DefaultPeriod, DefaultDigits)
}

func GenerateAt(secret string, timestamp int64, period, digits int) (string, error) {
	key, err := DecodeSecret(secret)
	if err != nil {
		return "", err
	}

	counter := timestamp / int64(period)

	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(counter))

	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	code := (int(hash[offset])&0x7f)<<24 |
		(int(hash[offset+1])&0xff)<<16 |
		(int(hash[offset+2])&0xff)<<8 |
		(int(hash[offset+3]) & 0xff)

	modulo := 1
	for i := 0; i < digits; i++ {
		modulo *= 10
	}
	otp := code % modulo

	return fmt.Sprintf("%0*d", digits, otp), nil
}

func Validate(secret, totp string) bool {
	return ValidateWithWindow(secret, totp, DefaultWindow)
}

func ValidateWithWindow(secret, totp string, window int) bool {
	timestamp := time.Now().Unix()
	return ValidateAtWithWindow(secret, totp, timestamp, DefaultPeriod, DefaultDigits, window)
}

func ValidateAtWithWindow(secret, totp string, timestamp int64, period, digits, window int) bool {
	if len(totp) != digits {
		return false
	}

	for i := -window; i <= window; i++ {
		ts := timestamp + int64(i*period)
		generated, err := GenerateAt(secret, ts, period, digits)
		if err != nil {
			continue
		}
		if generated == totp {
			return true
		}
	}
	return false
}
