package relay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type OffsetStore struct {
	basePath string
	mu       sync.RWMutex
	offsets  map[string]int64
}

type offsetData struct {
	Offsets map[string]int64 `json:"offsets"`
}

func NewOffsetStore(basePath string) (*OffsetStore, error) {
	store := &OffsetStore{
		basePath: basePath,
		offsets:  make(map[string]int64),
	}

	if err := store.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return store, nil
}

func (s *OffsetStore) load() error {
	dataPath := filepath.Join(s.basePath, "offsets.json")
	file, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}

	var data offsetData
	if err := json.Unmarshal(file, &data); err != nil {
		return err
	}

	s.offsets = data.Offsets
	return nil
}

func (s *OffsetStore) save() error {
	dataPath := filepath.Join(s.basePath, "offsets.json")
	
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return err
	}

	data := offsetData{
		Offsets: s.offsets,
	}

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataPath, file, 0644)
}

func (s *OffsetStore) GetOffset(filePath string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.offsets[filePath]
}

func (s *OffsetStore) SetOffset(filePath string, offset int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offsets[filePath] = offset
	return s.save()
}
