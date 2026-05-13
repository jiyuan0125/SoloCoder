package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"warehouse-mgmt/internal/models"
)

type Storage struct {
	filePath string
	mu       sync.RWMutex
}

func NewStorage(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

func (s *Storage) Load() (*models.Database, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Database{}, nil
		}
		return nil, fmt.Errorf("read file: %w", err)
	}

	var db models.Database
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &db, nil
}

func (s *Storage) Save(db *models.Database) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
