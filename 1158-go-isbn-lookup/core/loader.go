package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type LoadResult struct {
	TotalRecords   int
	LoadedRecords  int
	InvalidRecords []InvalidRecord
	Warnings       []string
}

type InvalidRecord struct {
	LineNumber int
	Line       string
	Error      string
}

func hasBOM(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF
}

func stripBOM(data []byte) []byte {
	if hasBOM(data) {
		return data[3:]
	}
	return data
}

func LoadFromReader(r io.Reader) (map[string]*Book, *LoadResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	data = stripBOM(data)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	result := &LoadResult{}
	books := make(map[string]*Book)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result.TotalRecords++
		book, err := parseLine(line)
		if err != nil {
			result.InvalidRecords = append(result.InvalidRecords, InvalidRecord{
				LineNumber: lineNum,
				Line:       line,
				Error:      err.Error(),
			})
			continue
		}
		if err := ValidateBook(book); err != nil {
			result.InvalidRecords = append(result.InvalidRecords, InvalidRecord{
				LineNumber: lineNum,
				Line:       line,
				Error:      err.Error(),
			})
			continue
		}
		existing, exists := books[book.ISBN]
		if exists {
			if !booksEqual(existing, book) {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Duplicate ISBN %s at line %d: overwriting previous record", book.ISBN, lineNum))
			}
		}
		books[book.ISBN] = book
		result.LoadedRecords++
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return books, result, nil
}

func parseLine(line string) (*Book, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 6 {
		return nil, ErrInvalidLineFormat
	}
	year, err := strconv.Atoi(strings.TrimSpace(parts[4]))
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}
	book := &Book{
		ISBN:         strings.TrimSpace(parts[0]),
		Title:        strings.TrimSpace(parts[1]),
		Authors:      ParseAuthors(parts[2]),
		Publisher:    strings.TrimSpace(parts[3]),
		Year:         year,
		CategoryCode: strings.TrimSpace(parts[5]),
	}
	return book, nil
}

func booksEqual(a, b *Book) bool {
	if a.ISBN != b.ISBN || a.Title != b.Title || a.Publisher != b.Publisher {
		return false
	}
	if a.Year != b.Year || a.CategoryCode != b.CategoryCode {
		return false
	}
	if len(a.Authors) != len(b.Authors) {
		return false
	}
	for i := range a.Authors {
		if a.Authors[i] != b.Authors[i] {
			return false
		}
	}
	return true
}
