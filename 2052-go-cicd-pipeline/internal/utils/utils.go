package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"
)

func GenerateID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}

func Now() time.Time {
	return time.Now().UTC()
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func TimePtrNil() *time.Time {
	return nil
}
