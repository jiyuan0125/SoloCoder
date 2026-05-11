package corpus

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"go-fuzz-corpus/api"
)

type Mutator struct {
	maxRetries int
}

func NewMutator() *Mutator {
	return &Mutator{
		maxRetries: 10,
	}
}

func randomInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func (m *Mutator) BitFlip(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	byteIdx, err := randomInt(len(data))
	if err != nil {
		return nil, err
	}

	bitIdx, err := randomInt(8)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(data))
	copy(result, data)
	result[byteIdx] ^= 1 << uint(bitIdx)

	return result, nil
}

func (m *Mutator) ByteFlip(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	byteIdx, err := randomInt(len(data))
	if err != nil {
		return nil, err
	}

	randomByte, err := randomInt(256)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(data))
	copy(result, data)
	result[byteIdx] = byte(randomByte)

	return result, nil
}

func (m *Mutator) Insert(data []byte) ([]byte, error) {
	insertPos, err := randomInt(len(data) + 1)
	if err != nil {
		return nil, err
	}

	randomByte, err := randomInt(256)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(data)+1)
	copy(result[:insertPos], data[:insertPos])
	result[insertPos] = byte(randomByte)
	copy(result[insertPos+1:], data[insertPos:])

	return result, nil
}

func (m *Mutator) mutateOne(data []byte) ([]byte, error) {
	strategy, err := randomInt(3)
	if err != nil {
		return nil, err
	}

	switch strategy {
	case 0:
		return m.BitFlip(data)
	case 1:
		return m.ByteFlip(data)
	case 2:
		return m.Insert(data)
	default:
		return m.BitFlip(data)
	}
}

func (m *Mutator) selectSeedFile(files []api.CorpusFile) (*api.CorpusFile, error) {
	var manualSeeds []api.CorpusFile
	var autoSeeds []api.CorpusFile

	for _, f := range files {
		if f.IsManual {
			manualSeeds = append(manualSeeds, f)
		} else {
			autoSeeds = append(autoSeeds, f)
		}
	}

	var pool []api.CorpusFile
	if len(manualSeeds) > 0 {
		pool = manualSeeds
	} else {
		pool = autoSeeds
	}

	if len(pool) == 0 {
		return nil, nil
	}

	idx, err := randomInt(len(pool))
	if err != nil {
		return nil, err
	}

	return &pool[idx], nil
}

func (m *Mutator) MutateCorpus(targetName, rootPath string, count int) (*api.MutateResponse, error) {
	corpusDir, err := FindCorpusDirForTarget(rootPath, targetName)
	if err != nil {
		return nil, err
	}

	stats, err := ScanCorpusDir(corpusDir)
	if err != nil {
		return nil, err
	}

	if len(stats.Files) == 0 {
		return &api.MutateResponse{
			TargetName: targetName,
			Generated:  0,
			Failed:     count,
			NewFiles:   []string{},
		}, nil
	}

	response := &api.MutateResponse{
		TargetName: targetName,
		Generated:  0,
		Failed:     0,
		NewFiles:   []string{},
	}

	existingHashes := make(map[string]bool)
	for _, f := range stats.Files {
		existingHashes[f.Hash] = true
	}

	for i := 0; i < count; i++ {
		var generated bool
		for retry := 0; retry < m.maxRetries; retry++ {
			seedFile, err := m.selectSeedFile(stats.Files)
			if err != nil {
				continue
			}
			if seedFile == nil {
				break
			}

			data, err := os.ReadFile(seedFile.Path)
			if err != nil {
				continue
			}

			mutated, err := m.mutateOne(data)
			if err != nil {
				continue
			}

			hash := sha256.Sum256(mutated)
			hashStr := hex.EncodeToString(hash[:])

			if existingHashes[hashStr] {
				continue
			}

			newFileName := hashStr
			newFilePath := filepath.Join(corpusDir, newFileName)

			err = os.WriteFile(newFilePath, mutated, 0644)
			if err != nil {
				continue
			}

			existingHashes[hashStr] = true
			response.Generated++
			response.NewFiles = append(response.NewFiles, newFilePath)
			generated = true

			newStat, _ := os.Stat(newFilePath)
			stats.Files = append(stats.Files, api.CorpusFile{
				Path:      newFilePath,
				Hash:      hashStr,
				Size:      int64(len(mutated)),
				AddedTime: newStat.ModTime(),
				Source:    "mutated",
				IsManual:  false,
			})
			stats.FileCount++
			stats.TotalSize += int64(len(mutated))

			break
		}

		if !generated {
			response.Failed++
		}
	}

	return response, nil
}

func (m *Mutator) generateHashFileName(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func init() {
	_ = time.Now
}
