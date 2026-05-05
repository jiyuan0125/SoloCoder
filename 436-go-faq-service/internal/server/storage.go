package server

import (
	"go-faq-service/pkg/protocol"
	"sync"
)

type Storage struct {
	faqs         map[string]*protocol.FAQ
	categories   map[string]*protocol.Category
	history      map[string][]*protocol.HistoryRecord
	clicks       map[string][]*protocol.ClickRecord
	mu           sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		faqs:       make(map[string]*protocol.FAQ),
		categories: make(map[string]*protocol.Category),
		history:    make(map[string][]*protocol.HistoryRecord),
		clicks:     make(map[string][]*protocol.ClickRecord),
	}
}

func (s *Storage) CreateFAQ(faq *protocol.FAQ) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faqs[faq.ID] = faq
}

func (s *Storage) GetFAQ(id string) (*protocol.FAQ, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	faq, exists := s.faqs[id]
	return faq, exists
}

func (s *Storage) UpdateFAQ(faq *protocol.FAQ) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faqs[faq.ID] = faq
}

func (s *Storage) DeleteFAQ(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.faqs, id)
}

func (s *Storage) ListAllFAQs() []*protocol.FAQ {
	s.mu.RLock()
	defer s.mu.RUnlock()
	faqs := make([]*protocol.FAQ, 0, len(s.faqs))
	for _, faq := range s.faqs {
		faqs = append(faqs, faq)
	}
	return faqs
}

func (s *Storage) GetFAQsByCategory(categoryID string) []*protocol.FAQ {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var faqs []*protocol.FAQ
	for _, faq := range s.faqs {
		if faq.CategoryID == categoryID {
			faqs = append(faqs, faq)
		}
	}
	return faqs
}

func (s *Storage) GetMaxSortWeight(categoryID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	maxWeight := 0
	for _, faq := range s.faqs {
		if faq.CategoryID == categoryID && faq.SortWeight > maxWeight {
			maxWeight = faq.SortWeight
		}
	}
	return maxWeight
}

func (s *Storage) IsSortWeightUsed(categoryID string, weight int, excludeFAQID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, faq := range s.faqs {
		if faq.CategoryID == categoryID && faq.SortWeight == weight && faq.ID != excludeFAQID {
			return true
		}
	}
	return false
}

func (s *Storage) CreateCategory(category *protocol.Category) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.categories[category.ID] = category
}

func (s *Storage) GetCategory(id string) (*protocol.Category, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	category, exists := s.categories[id]
	return category, exists
}

func (s *Storage) GetCategoryByName(name string, parentID string) (*protocol.Category, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, cat := range s.categories {
		if cat.Name == name && cat.ParentID == parentID {
			return cat, true
		}
	}
	return nil, false
}

func (s *Storage) UpdateCategory(category *protocol.Category) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.categories[category.ID] = category
}

func (s *Storage) DeleteCategory(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.categories, id)
}

func (s *Storage) ListAllCategories() []*protocol.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	categories := make([]*protocol.Category, 0, len(s.categories))
	for _, cat := range s.categories {
		categories = append(categories, cat)
	}
	return categories
}

func (s *Storage) GetChildCategories(parentID string) []*protocol.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var children []*protocol.Category
	for _, cat := range s.categories {
		if cat.ParentID == parentID {
			children = append(children, cat)
		}
	}
	return children
}

func (s *Storage) AddHistoryRecord(record *protocol.HistoryRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history[record.FAQID] = append(s.history[record.FAQID], record)
}

func (s *Storage) GetHistory(faqID string) []*protocol.HistoryRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.history[faqID]
}

func (s *Storage) AddClickRecord(record *protocol.ClickRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clicks[record.FAQID] = append(s.clicks[record.FAQID], record)
}

func (s *Storage) GetClickStats(faqID string) (total, helpful int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clicks := s.clicks[faqID]
	total = len(clicks)
	helpful = 0
	for _, click := range clicks {
		if click.IsHelpful {
			helpful++
		}
	}
	return total, helpful
}

func (s *Storage) GetAllClickStats() map[string][2]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string][2]int)
	for faqID, clicks := range s.clicks {
		total := len(clicks)
		helpful := 0
		for _, click := range clicks {
			if click.IsHelpful {
				helpful++
			}
		}
		result[faqID] = [2]int{total, helpful}
	}
	return result
}
