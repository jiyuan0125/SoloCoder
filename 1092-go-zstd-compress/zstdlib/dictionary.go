package zstdlib

import (
	"crypto/sha256"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type DictionaryManager struct {
	storeDir     string
	dictionaries map[string]DictionaryInfo
	mu           sync.RWMutex
}

type DictionaryInfo struct {
	Name        string
	SampleCount int
	DictSize    int
	CreatedAt   int64
	Data        []byte
	Hash        [32]byte
}

func NewDictionaryManager(storeDir string) (*DictionaryManager, error) {
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, err
	}

	return &DictionaryManager{
		storeDir:     storeDir,
		dictionaries: make(map[string]DictionaryInfo),
	}, nil
}

func (dm *DictionaryManager) TrainDictionary(dictName string, samplePaths []string) (*DictionaryInfo, error) {
	if len(samplePaths) == 0 {
		return nil, errors.New("no sample files provided")
	}

	var samples [][]byte
	for _, path := range samplePaths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			samples = append(samples, data)
		}
	}

	if len(samples) == 0 {
		return nil, errors.New("all sample files are empty")
	}

	var combined []byte
	for _, sample := range samples {
		combined = append(combined, sample...)
	}

	hash := sha256.Sum256(combined)

	dictData := make([]byte, 0, len(combined))
	if len(combined) > 128*1024 {
		dictData = append(dictData, combined[:64*1024]...)
		dictData = append(dictData, combined[len(combined)-64*1024:]...)
	} else {
		dictData = append(dictData, combined...)
	}

	info := DictionaryInfo{
		Name:        dictName,
		SampleCount: len(samplePaths),
		DictSize:    len(dictData),
		CreatedAt:   0,
		Data:        dictData,
		Hash:        hash,
	}

	dm.mu.Lock()
	dm.dictionaries[dictName] = info
	dm.mu.Unlock()

	dictPath := filepath.Join(dm.storeDir, dictName+".dict")
	if err := ioutil.WriteFile(dictPath, dictData, 0644); err != nil {
		return nil, err
	}

	infoPath := filepath.Join(dm.storeDir, dictName+".info")
	infoContent := []byte(dictName)
	if err := ioutil.WriteFile(infoPath, infoContent, 0644); err != nil {
		return nil, err
	}

	return &info, nil
}

func (dm *DictionaryManager) GetDictionary(dictName string) ([]byte, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	info, ok := dm.dictionaries[dictName]
	if !ok {
		return nil, false
	}
	return info.Data, true
}

func (dm *DictionaryManager) ListDictionaries() []DictionaryInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	result := make([]DictionaryInfo, 0, len(dm.dictionaries))
	for _, info := range dm.dictionaries {
		result = append(result, DictionaryInfo{
			Name:        info.Name,
			SampleCount: info.SampleCount,
			DictSize:    info.DictSize,
			CreatedAt:   info.CreatedAt,
		})
	}
	return result
}

func (dm *DictionaryManager) LoadExistingDictionaries() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	files, err := filepath.Glob(filepath.Join(dm.storeDir, "*.dict"))
	if err != nil {
		return err
	}

	for _, filePath := range files {
		name := filepath.Base(filePath)
		name = name[:len(name)-len(".dict")]

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		dm.dictionaries[name] = DictionaryInfo{
			Name:        name,
			SampleCount: 0,
			DictSize:    len(data),
			CreatedAt:   0,
			Data:        data,
		}
	}

	return nil
}
