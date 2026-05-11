package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"go-fuzz-corpus/api"
)

func ImportFromCrashDir(crashDir, targetName, rootPath string) (*api.ImportResponse, error) {
	corpusDir, err := FindCorpusDirForTarget(rootPath, targetName)
	if err != nil {
		return nil, err
	}

	stats, err := ScanCorpusDir(corpusDir)
	if err != nil {
		return nil, err
	}

	existingHashes := make(map[string]bool)
	for _, f := range stats.Files {
		existingHashes[f.Hash] = true
	}

	response := &api.ImportResponse{
		TargetName: targetName,
		Imported:   0,
		Skipped:    0,
		NewFiles:   []string{},
	}

	entries, err := os.ReadDir(crashDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(crashDir, entry.Name())

		data, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

		if existingHashes[hashStr] {
			response.Skipped++
			continue
		}

		newFileName := hashStr
		newFilePath := filepath.Join(corpusDir, newFileName)

		err = os.WriteFile(newFilePath, data, 0644)
		if err != nil {
			continue
		}

		existingHashes[hashStr] = true
		response.Imported++
		response.NewFiles = append(response.NewFiles, newFilePath)
	}

	return response, nil
}

func DeleteCorpusFile(filePath string) (*api.DeleteFileResponse, error) {
	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &api.DeleteFileResponse{
				Success: false,
				Message: "file not found",
			}, nil
		}
		return nil, err
	}

	return &api.DeleteFileResponse{
		Success: true,
		Message: "file deleted successfully",
	}, nil
}
