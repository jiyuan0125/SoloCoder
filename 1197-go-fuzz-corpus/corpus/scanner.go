package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go-fuzz-corpus/api"
)

var hashFileNameRegex = regexp.MustCompile(`^[0-9a-fA-F]{16,}$`)

func IsManualSeed(filename string) bool {
	return !hashFileNameRegex.MatchString(filename)
}

func computeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func FindCorpusDirs(rootPath string) ([]string, error) {
	var corpusDirs []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) >= 3 {
			if parts[len(parts)-3] == "testdata" && parts[len(parts)-2] == "fuzz" {
				corpusDirs = append(corpusDirs, path)
			}
		}

		return nil
	})

	return corpusDirs, err
}

func GetTargetNameFromPath(corpusPath string) string {
	return filepath.Base(corpusPath)
}

func ScanCorpusDir(corpusPath string) (*api.TargetStats, error) {
	targetName := GetTargetNameFromPath(corpusPath)
	stats := &api.TargetStats{
		TargetName: targetName,
		FilePath:   corpusPath,
		FileCount:  0,
		TotalSize:  0,
		Files:      []api.CorpusFile{},
	}

	entries, err := os.ReadDir(corpusPath)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(corpusPath, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		hash, err := computeFileHash(fullPath)
		if err != nil {
			continue
		}

		corpusFile := api.CorpusFile{
			Path:      fullPath,
			Hash:      hash,
			Size:      fileInfo.Size(),
			AddedTime: fileInfo.ModTime(),
			Source:    "corpus",
			IsManual:  IsManualSeed(entry.Name()),
		}

		stats.Files = append(stats.Files, corpusFile)
		stats.FileCount++
		stats.TotalSize += fileInfo.Size()
	}

	return stats, nil
}

func ScanRoot(rootPath string) ([]api.TargetStats, error) {
	corpusDirs, err := FindCorpusDirs(rootPath)
	if err != nil {
		return nil, err
	}

	var allStats []api.TargetStats
	for _, dir := range corpusDirs {
		stats, err := ScanCorpusDir(dir)
		if err != nil {
			return nil, err
		}
		allStats = append(allStats, *stats)
	}

	return allStats, nil
}

func FindCorpusDirForTarget(rootPath, targetName string) (string, error) {
	corpusDirs, err := FindCorpusDirs(rootPath)
	if err != nil {
		return "", err
	}

	for _, dir := range corpusDirs {
		if GetTargetNameFromPath(dir) == targetName {
			return dir, nil
		}
	}

	return "", fmt.Errorf("target %s not found", targetName)
}
