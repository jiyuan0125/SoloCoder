package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type StorageService struct {
	BaseDir string
}

func NewStorageService(baseDir string) (*StorageService, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &StorageService{BaseDir: baseDir}, nil
}

func (s *StorageService) getUserDir(userID string) string {
	return filepath.Join(s.BaseDir, sanitizeUserID(userID))
}

func sanitizeUserID(userID string) string {
	return strings.ReplaceAll(userID, "/", "_")
}

func (s *StorageService) IsValidFilename(filename string) bool {
	if strings.Contains(filename, "..") {
		return false
	}
	if strings.Contains(filename, "/") {
		return false
	}
	if strings.Contains(filename, "\\") {
		return false
	}
	return true
}

func (s *StorageService) SaveFile(userID, filename string, reader io.Reader) (string, int64, error) {
	if !s.IsValidFilename(filename) {
		return "", 0, fmt.Errorf("invalid filename: contains path traversal characters")
	}

	userDir := s.getUserDir(userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", 0, err
	}

	storedID := generateID()
	storedPath := filepath.Join(userDir, storedID)

	file, err := os.Create(storedPath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	size, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(storedPath)
		return "", 0, err
	}

	return storedID, size, nil
}

func (s *StorageService) OpenFile(userID, storedID string) (*os.File, error) {
	userDir := s.getUserDir(userID)
	storedPath := filepath.Join(userDir, storedID)

	cleanPath := filepath.Clean(storedPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.BaseDir)) {
		return nil, fmt.Errorf("path traversal detected")
	}

	return os.Open(cleanPath)
}

func (s *StorageService) DeleteFile(userID, storedID string) error {
	userDir := s.getUserDir(userID)
	storedPath := filepath.Join(userDir, storedID)
	return os.Remove(storedPath)
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
