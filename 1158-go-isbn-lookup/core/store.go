package core

import (
	"sort"
	"strings"
	"sync"
)

type Query struct {
	ISBN      string
	Title     string
	Authors   []string
	Publisher string
}

type PageRequest struct {
	Page     int
	PageSize int
}

type PageResult struct {
	Total       int
	TotalPages  int
	CurrentPage int
	PageSize    int
	Books       []*Book
}

func DefaultPageRequest() PageRequest {
	return PageRequest{Page: 1, PageSize: 20}
}

type BookStore struct {
	mu    sync.RWMutex
	books map[string]*Book
}

func NewBookStore() *BookStore {
	return &BookStore{books: make(map[string]*Book)}
}

func (s *BookStore) LoadAll(books map[string]*Book) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.books = books
}

func (s *BookStore) Add(book *Book) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateBook(book); err != nil {
		return false
	}
	s.books[book.ISBN] = book
	return true
}

func (s *BookStore) ReplaceAll(books map[string]*Book) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.books = make(map[string]*Book, len(books))
	for k, v := range books {
		s.books[k] = v
	}
}

func (s *BookStore) GetByISBN(isbn string) *Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.books[isbn]
}

func (s *BookStore) Query(q *Query, pr *PageRequest) *PageResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var matches []*Book
	for _, b := range s.books {
		if matchQuery(b, q) {
			matches = append(matches, b)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return CompareBookByTitle(matches[i], matches[j]) < 0
	})
	total := len(matches)
	if pr == nil {
		return &PageResult{
			Total:       total,
			TotalPages:  1,
			CurrentPage: 1,
			PageSize:    total,
			Books:       matches,
		}
	}
	page := pr.Page
	if page < 1 {
		page = 1
	}
	pageSize := pr.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total-1)/pageSize + 1
	}
	if page > totalPages && total > 0 {
		page = totalPages
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	return &PageResult{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		Books:       matches[start:end],
	}
}

func matchQuery(b *Book, q *Query) bool {
	if q.ISBN != "" {
		if b.ISBN != q.ISBN {
			return false
		}
	}
	if q.Title != "" {
		title := strings.ToLower(b.Title)
		queryTitle := strings.ToLower(q.Title)
		if !strings.Contains(title, queryTitle) {
			return false
		}
	}
	if len(q.Authors) > 0 {
		matchedAny := false
		for _, a := range q.Authors {
			if AuthorsContain(b.Authors, a) {
				matchedAny = true
				break
			}
		}
		if !matchedAny {
			return false
		}
	}
	if q.Publisher != "" {
		pub := strings.ToLower(b.Publisher)
		queryPub := strings.ToLower(q.Publisher)
		if !strings.Contains(pub, queryPub) {
			return false
		}
	}
	return true
}
