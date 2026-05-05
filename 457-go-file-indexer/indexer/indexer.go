package indexer

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go-file-indexer/common"
)

type Indexer struct {
	indexFilePath    string
	watchedDirectory string
	index            *InvertedIndex
	lastIndexedTime  time.Time
}

type InvertedIndex struct {
	WordIndex   map[string]map[string][]WordLocation
	FileModTime map[string]time.Time
	IndexTime   time.Time
}

type WordLocation struct {
	LineNum     int
	StartOffset int
	EndOffset   int
}

type PhraseMatch struct {
	FileName string
	LineNum  int
	LineText string
	Score    int
}

func NewIndexer(indexFilePath string) *Indexer {
	return &Indexer{
		indexFilePath: indexFilePath,
		index: &InvertedIndex{
			WordIndex:   make(map[string]map[string][]WordLocation),
			FileModTime: make(map[string]time.Time),
		},
	}
}

func (idx *Indexer) BuildIndex(directory string) error {
	idx.watchedDirectory = directory

	absDir, err := filepath.Abs(directory)
	if err != nil {
		return err
	}

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: Cannot access file %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if IsBinaryFile(path) {
			log.Printf("Skipping binary file: %s", path)
			return nil
		}

		if isExcluded(path) {
			return nil
		}

		err = idx.indexSingleFile(path, info.ModTime())
		if err != nil {
			log.Printf("Warning: Error indexing file %s: %v", path, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	idx.lastIndexedTime = time.Now()
	idx.index.IndexTime = idx.lastIndexedTime

	return nil
}

func (idx *Indexer) removeFileFromIndex(filePath string) {
	for word, fileMap := range idx.index.WordIndex {
		if _, ok := fileMap[filePath]; ok {
			delete(fileMap, filePath)
			if len(fileMap) == 0 {
				delete(idx.index.WordIndex, word)
			}
		}
	}
	delete(idx.index.FileModTime, filePath)
}

func (idx *Indexer) indexSingleFile(filePath string, modTime time.Time) error {
	idx.removeFileFromIndex(filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		tokenizeLine(line, lineNum+1, filePath, idx.index)
	}

	idx.index.FileModTime[filePath] = modTime

	return nil
}

func tokenizeLine(line string, lineNum int, filePath string, index *InvertedIndex) {
	re := regexp.MustCompile(`\w+`)
	matches := re.FindAllStringIndex(line, -1)

	for _, match := range matches {
		startOffset := match[0]
		endOffset := match[1]
		word := strings.ToLower(line[startOffset:endOffset])

		location := WordLocation{
			LineNum:     lineNum,
			StartOffset: startOffset,
			EndOffset:   endOffset,
		}

		if index.WordIndex[word] == nil {
			index.WordIndex[word] = make(map[string][]WordLocation)
		}
		index.WordIndex[word][filePath] = append(index.WordIndex[word][filePath], location)
	}
}

func (idx *Indexer) Search(query string) ([]common.SearchHit, error) {
	hits, err := idx.searchInternal(query)
	if err != nil {
		return nil, err
	}

	return rankHits(hits), nil
}

func (idx *Indexer) searchInternal(query string) ([]common.SearchHit, error) {
	var hits []common.SearchHit

	if strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"") {
		phrase := strings.Trim(query, "\"")
		return idx.searchPhrase(phrase)
	}

	query = strings.ToLower(query)
	words := extractWords(query)

	if len(words) == 0 {
		return hits, nil
	}

	candidateFiles := idx.findFilesContainingAllWords(words)

	for fileName := range candidateFiles {
		fileHits := idx.searchFileForWords(fileName, words)
		hits = append(hits, fileHits...)
	}

	return hits, nil
}

func (idx *Indexer) searchPhrase(phrase string) ([]common.SearchHit, error) {
	var hits []common.SearchHit

	words := extractWords(phrase)
	phraseLower := strings.ToLower(phrase)

	if len(words) == 0 {
		return hits, nil
	}

	candidateFiles := idx.findFilesContainingAllWords(words)
	if len(candidateFiles) == 0 {
		return hits, nil
	}

	for fileName := range candidateFiles {
		candidateLines := idx.findLinesContainingAllWords(fileName, words)
		for lineNum := range candidateLines {
			lineContent, err := getLineFromFile(fileName, lineNum)
			if err != nil {
				log.Printf("Warning: Could not read line %d from %s: %v", lineNum, fileName, err)
				continue
			}

			if strings.Contains(strings.ToLower(lineContent), phraseLower) {
				hit := common.SearchHit{
					FileName:    fileName,
					LineNumber:  lineNum,
					LineContent: lineContent,
					Score:       len(words) * 10,
				}
				hits = append(hits, hit)
			}
		}
	}

	return deduplicateHits(hits), nil
}

func (idx *Indexer) findLinesContainingAllWords(fileName string, words []string) map[int]bool {
	if len(words) == 0 {
		return nil
	}

	result := make(map[int]bool)
	firstWord := words[0]

	fileMap, ok := idx.index.WordIndex[firstWord]
	if !ok {
		return result
	}

	locations, ok := fileMap[fileName]
	if !ok {
		return result
	}

	for _, loc := range locations {
		result[loc.LineNum] = true
	}

	for i := 1; i < len(words); i++ {
		word := words[i]
		fileMap, ok := idx.index.WordIndex[word]
		if !ok {
			return make(map[int]bool)
		}

		locations, ok := fileMap[fileName]
		if !ok {
			return make(map[int]bool)
		}

		wordLines := make(map[int]bool)
		for _, loc := range locations {
			wordLines[loc.LineNum] = true
		}

		newResult := make(map[int]bool)
		for lineNum := range result {
			if wordLines[lineNum] {
				newResult[lineNum] = true
			}
		}
		result = newResult

		if len(result) == 0 {
			break
		}
	}

	return result
}

func deduplicateHits(hits []common.SearchHit) []common.SearchHit {
	if len(hits) <= 1 {
		return hits
	}

	type key struct {
		fileName string
		lineNum  int
	}

	seen := make(map[key]bool)
	var result []common.SearchHit

	for _, hit := range hits {
		k := key{hit.FileName, hit.LineNumber}
		if !seen[k] {
			seen[k] = true
			result = append(result, hit)
		}
	}

	return result
}

func (idx *Indexer) findFilesContainingAllWords(words []string) map[string]bool {
	if len(words) == 0 {
		return nil
	}

	result := make(map[string]bool)
	firstWord := words[0]

	fileMap, ok := idx.index.WordIndex[firstWord]
	if !ok {
		return result
	}

	for fileName := range fileMap {
		result[fileName] = true
	}

	for i := 1; i < len(words); i++ {
		word := words[i]
		fileMap, ok := idx.index.WordIndex[word]
		if !ok {
			return make(map[string]bool)
		}

		newResult := make(map[string]bool)
		for fileName := range result {
			if _, exists := fileMap[fileName]; exists {
				newResult[fileName] = true
			}
		}
		result = newResult

		if len(result) == 0 {
			break
		}
	}

	return result
}

func (idx *Indexer) searchFileForWords(fileName string, words []string) []common.SearchHit {
	var hits []common.SearchHit

	lineMatchCount := make(map[int]int)

	for _, word := range words {
		fileMap, ok := idx.index.WordIndex[word]
		if !ok {
			continue
		}

		locations, ok := fileMap[fileName]
		if !ok {
			continue
		}

		for _, loc := range locations {
			lineMatchCount[loc.LineNum]++
		}
	}

	for lineNum, matchCount := range lineMatchCount {
		if matchCount > 0 {
			lineContent, err := getLineFromFile(fileName, lineNum)
			if err != nil {
				log.Printf("Warning: Could not read line %d from %s: %v", lineNum, fileName, err)
				continue
			}

			hit := common.SearchHit{
				FileName:    fileName,
				LineNumber:  lineNum,
				LineContent: lineContent,
				Score:       matchCount * 10,
			}

			hits = append(hits, hit)
		}
	}

	return hits
}

func rankHits(hits []common.SearchHit) []common.SearchHit {
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[i].Score < hits[j].Score {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
	return hits
}

func (idx *Indexer) IncrementalUpdate() error {
	if idx.watchedDirectory == "" {
		return nil
	}

	absDir, err := filepath.Abs(idx.watchedDirectory)
	if err != nil {
		return err
	}

	currentFiles := make(map[string]bool)

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: Cannot access file %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if IsBinaryFile(path) {
			return nil
		}

		if isExcluded(path) {
			return nil
		}

		currentFiles[path] = true

		modTime, exists := idx.index.FileModTime[path]
		if !exists || info.ModTime().After(modTime) {
			err = idx.indexSingleFile(path, info.ModTime())
			if err != nil {
				log.Printf("Warning: Error re-indexing file %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	idx.removeDeletedFiles(currentFiles)

	idx.lastIndexedTime = time.Now()
	idx.index.IndexTime = idx.lastIndexedTime

	return nil
}

func (idx *Indexer) removeDeletedFiles(currentFiles map[string]bool) {
	for filePath := range idx.index.FileModTime {
		if _, exists := currentFiles[filePath]; !exists {
			for word, fileMap := range idx.index.WordIndex {
				if _, ok := fileMap[filePath]; ok {
					delete(fileMap, filePath)
					if len(fileMap) == 0 {
						delete(idx.index.WordIndex, word)
					}
				}
			}
			delete(idx.index.FileModTime, filePath)
		}
	}
}

func (idx *Indexer) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["indexed_files"] = len(idx.index.FileModTime)
	stats["indexed_words"] = len(idx.index.WordIndex)
	stats["last_indexed_time"] = idx.lastIndexedTime.Format(time.RFC3339)

	return stats
}
