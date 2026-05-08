package core

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strings"

	"hash-checkpoint/common"
)

func CalculateChunkHash(data []byte, algorithm common.HashAlgorithm) (string, error) {
	var h hash.Hash
	switch strings.ToUpper(string(algorithm)) {
	case string(common.AlgorithmMD5):
		h = md5.New()
	case string(common.AlgorithmSHA256):
		h = sha256.New()
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func CalculateStreamHash(r io.Reader, algorithm common.HashAlgorithm) (string, error) {
	var h hash.Hash
	switch strings.ToUpper(string(algorithm)) {
	case string(common.AlgorithmMD5):
		h = md5.New()
	case string(common.AlgorithmSHA256):
		h = sha256.New()
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func CalculateFinalHash(chunkHashes []string, algorithm common.HashAlgorithm) (string, error) {
	concatenated := strings.Join(chunkHashes, "")
	return CalculateChunkHash([]byte(concatenated), algorithm)
}
