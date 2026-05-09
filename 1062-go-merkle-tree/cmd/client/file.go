package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"merkle-tree/pkg/merkle"
)

func computeFileHashes(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hashes := []string{}
	buf := make([]byte, merkle.BlockSize)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		chunk := buf[:n]
		hash := sha256.Sum256(chunk)
		hashes = append(hashes, hex.EncodeToString(hash[:]))
	}

	return hashes, nil
}
