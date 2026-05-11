package core

import (
	"strings"
	"time"
)

type Book struct {
	ISBN         string
	Title        string
	Authors      []string
	Publisher    string
	Year         int
	CategoryCode string
}

func ParseAuthors(raw string) []string {
	raw = strings.ReplaceAll(raw, ";", ",")
	parts := strings.Split(raw, ",")
	var authors []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			authors = append(authors, p)
		}
	}
	return authors
}

func NormalizeForMatch(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if !strings.ContainsRune(" .,!?;:()[]{}'\"-", r) {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func AuthorsContain(authors []string, keyword string) bool {
	keyword = NormalizeForMatch(keyword)
	if keyword == "" {
		return false
	}
	for _, a := range authors {
		if strings.Contains(NormalizeForMatch(a), keyword) {
			return true
		}
	}
	return false
}

func ValidateBook(b *Book) error {
	if b.Title == "" {
		return ErrEmptyTitle
	}
	if len(b.Authors) == 0 {
		return ErrEmptyAuthor
	}
	if !ValidateISBN(b.ISBN) {
		return ErrInvalidISBN
	}
	currentYear := time.Now().Year()
	if b.Year < 1450 || b.Year > currentYear+5 {
		return ErrInvalidYear
	}
	return nil
}

func CompareBookByTitle(a, b *Book) int {
	return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
}
