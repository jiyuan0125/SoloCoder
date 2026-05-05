package indexer

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (idx *Indexer) SaveIndex() error {
	dir := filepath.Dir(idx.indexFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for index file: %v", err)
	}

	file, err := os.Create(idx.indexFilePath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %v", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(idx.index); err != nil {
		return fmt.Errorf("failed to encode index: %v", err)
	}

	return nil
}

func (idx *Indexer) LoadIndex() error {
	file, err := os.Open(idx.indexFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			idx.index = &InvertedIndex{
				WordIndex:   make(map[string]map[string][]WordLocation),
				FileModTime: make(map[string]time.Time),
			}
			return nil
		}
		return fmt.Errorf("failed to open index file: %v", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	idx.index = &InvertedIndex{}
	if err := decoder.Decode(idx.index); err != nil {
		idx.index = &InvertedIndex{
			WordIndex:   make(map[string]map[string][]WordLocation),
			FileModTime: make(map[string]time.Time),
		}
		return fmt.Errorf("failed to decode index: %v", err)
	}

	idx.lastIndexedTime = idx.index.IndexTime

	return nil
}

func getLineFromFile(fileName string, lineNum int) (string, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if lineNum < 1 || lineNum > len(lines) {
		return "", fmt.Errorf("line number out of range")
	}

	return lines[lineNum-1], nil
}

func extractWords(text string) []string {
	var words []string
	var currentWord []rune

	for _, r := range text {
		if isWordChar(r) {
			currentWord = append(currentWord, r)
		} else {
			if len(currentWord) > 0 {
				words = append(words, strings.ToLower(string(currentWord)))
				currentWord = []rune{}
			}
		}
	}

	if len(currentWord) > 0 {
		words = append(words, strings.ToLower(string(currentWord)))
	}

	return words
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' ||
		(r >= 128)
}
