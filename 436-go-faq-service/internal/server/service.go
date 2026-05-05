package server

import (
	"encoding/json"
	"fmt"
	"go-faq-service/pkg/protocol"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const HotViewThreshold = 1000
const OptimizeThresholdRatio = 0.3

type Service struct {
	storage *Storage
}

func NewService(storage *Storage) *Service {
	return &Service{storage: storage}
}

func generateID() string {
	return uuid.New().String()
}

func (s *Service) CreateFAQ(req *protocol.CreateFAQRequest) (*protocol.FAQ, error) {
	if req.CategoryID == "" {
		return nil, fmt.Errorf("category_id is required")
	}

	if _, exists := s.storage.GetCategory(req.CategoryID); !exists {
		return nil, fmt.Errorf("category not found")
	}

	if len(req.Content) == 0 {
		return nil, fmt.Errorf("content is required")
	}

	for lang, content := range req.Content {
		if content.Question == "" {
			return nil, fmt.Errorf("question is required for language %s", lang)
		}
		if content.Answer == "" {
			return nil, fmt.Errorf("answer is required for language %s", lang)
		}
	}

	maxWeight := s.storage.GetMaxSortWeight(req.CategoryID)
	sortWeight := maxWeight + 1

	now := time.Now()
	faq := &protocol.FAQ{
		ID:           generateID(),
		CategoryID:   req.CategoryID,
		Content:      req.Content,
		SortWeight:   sortWeight,
		IsEnabled:    true,
		IsPinned:     false,
		ViewCount:    0,
		IsHot:        false,
		NeedOptimize: false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.storage.CreateFAQ(faq)

	s.addHistory(faq.ID, "create", "", faqToJSON(faq), "system")

	return faq, nil
}

func (s *Service) GetFAQ(id string, lang protocol.Language) (*protocol.FAQDetailResponse, error) {
	faq, exists := s.storage.GetFAQ(id)
	if !exists {
		return nil, fmt.Errorf("FAQ not found")
	}

	if !faq.IsEnabled {
		return nil, fmt.Errorf("FAQ is not enabled")
	}

	faq.ViewCount++
	if faq.ViewCount >= HotViewThreshold && !faq.IsHot {
		faq.IsHot = true
	}
	s.storage.UpdateFAQ(faq)

	category, _ := s.storage.GetCategory(faq.CategoryID)
	categoryName := ""
	if category != nil {
		categoryName = category.Name
	}

	return &protocol.FAQDetailResponse{
		FAQ:          *faq,
		CategoryName: categoryName,
	}, nil
}

func (s *Service) UpdateFAQ(id string, req *protocol.UpdateFAQRequest) (*protocol.FAQ, error) {
	faq, exists := s.storage.GetFAQ(id)
	if !exists {
		return nil, fmt.Errorf("FAQ not found")
	}

	oldContent := faqToJSON(faq)

	if len(req.Content) > 0 {
		for lang, content := range req.Content {
			if content.Question == "" {
				return nil, fmt.Errorf("question is required for language %s", lang)
			}
			if content.Answer == "" {
				return nil, fmt.Errorf("answer is required for language %s", lang)
			}
		}
		faq.Content = req.Content
	}

	faq.UpdatedAt = time.Now()
	s.storage.UpdateFAQ(faq)

	s.addHistory(faq.ID, "update", oldContent, faqToJSON(faq), "system")

	return faq, nil
}

func (s *Service) DeleteFAQ(id string) error {
	faq, exists := s.storage.GetFAQ(id)
	if !exists {
		return fmt.Errorf("FAQ not found")
	}

	s.addHistory(faq.ID, "delete", faqToJSON(faq), "", "system")
	s.storage.DeleteFAQ(id)
	return nil
}

func (s *Service) SetFAQEnabled(id string, isEnabled bool) (*protocol.FAQ, error) {
	faq, exists := s.storage.GetFAQ(id)
	if !exists {
		return nil, fmt.Errorf("FAQ not found")
	}

	oldContent := faqToJSON(faq)
	faq.IsEnabled = isEnabled
	faq.UpdatedAt = time.Now()
	s.storage.UpdateFAQ(faq)

	action := "enable"
	if !isEnabled {
		action = "disable"
	}
	s.addHistory(faq.ID, action, oldContent, faqToJSON(faq), "system")

	return faq, nil
}

func (s *Service) SetFAQPinned(id string, isPinned bool) (*protocol.FAQ, error) {
	faq, exists := s.storage.GetFAQ(id)
	if !exists {
		return nil, fmt.Errorf("FAQ not found")
	}

	oldContent := faqToJSON(faq)
	faq.IsPinned = isPinned
	faq.UpdatedAt = time.Now()
	s.storage.UpdateFAQ(faq)

	action := "pin"
	if !isPinned {
		action = "unpin"
	}
	s.addHistory(faq.ID, action, oldContent, faqToJSON(faq), "system")

	return faq, nil
}

func (s *Service) SearchFAQ(req *protocol.SearchFAQRequest) (*protocol.SearchFAQResponse, error) {
	allFAQs := s.storage.ListAllFAQs()
	keyword := strings.ToLower(req.Keyword)

	var results []protocol.SearchResultItem

	for _, faq := range allFAQs {
		if !faq.IsEnabled {
			continue
		}

		content, exists := faq.Content[req.Language]
		if !exists {
			for _, c := range faq.Content {
				content = c
				break
			}
		}

		if content.Question == "" && content.Answer == "" {
			continue
		}

		matchScore := calculateMatchScore(content.Question, content.Answer, keyword)
		if matchScore <= 0 && req.Keyword != "" {
			continue
		}

		category, _ := s.storage.GetCategory(faq.CategoryID)
		categoryName := ""
		if category != nil {
			categoryName = category.Name
		}

		results = append(results, protocol.SearchResultItem{
			FAQ:          *faq,
			CategoryName: categoryName,
			MatchScore:   matchScore,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].FAQ.IsPinned != results[j].FAQ.IsPinned {
			return results[i].FAQ.IsPinned
		}
		if results[i].MatchScore != results[j].MatchScore {
			return results[i].MatchScore > results[j].MatchScore
		}
		if results[i].FAQ.IsHot != results[j].FAQ.IsHot {
			return results[i].FAQ.IsHot
		}
		return results[i].FAQ.SortWeight < results[j].FAQ.SortWeight
	})

	total := len(results)
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &protocol.SearchFAQResponse{
		Items:      results[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func calculateMatchScore(question, answer, keyword string) float64 {
	if keyword == "" {
		return 1.0
	}

	qLower := strings.ToLower(question)
	aLower := strings.ToLower(answer)

	qCount := strings.Count(qLower, keyword)
	aCount := strings.Count(aLower, keyword)

	if qCount == 0 && aCount == 0 {
		return 0
	}

	score := 0.0
	if qCount > 0 {
		score += 2.0 * float64(qCount)
	}
	if aCount > 0 {
		score += 1.0 * float64(aCount)
	}

	return score
}

func (s *Service) BatchImportFAQ(req *protocol.BatchImportFAQRequest) (*protocol.BatchImportResult, error) {
	result := &protocol.BatchImportResult{
		SuccessCount: 0,
		FailedCount:  0,
		FailedItems:  []protocol.BatchImportFailedItem{},
	}

	for i, item := range req.Items {
		createReq := &protocol.CreateFAQRequest{
			CategoryID: item.CategoryID,
			Content:    item.Content,
		}

		_, err := s.CreateFAQ(createReq)
		if err != nil {
			result.FailedCount++
			itemJSON, _ := json.Marshal(item)
			result.FailedItems = append(result.FailedItems, protocol.BatchImportFailedItem{
				Index: i,
				Item:  string(itemJSON),
				Error: err.Error(),
			})
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

func (s *Service) RecordClick(faqID string, req *protocol.RecordClickRequest) error {
	_, exists := s.storage.GetFAQ(faqID)
	if !exists {
		return fmt.Errorf("FAQ not found")
	}

	click := &protocol.ClickRecord{
		ID:        generateID(),
		FAQID:     faqID,
		UserID:    req.UserID,
		IsHelpful: req.IsHelpful,
		CreatedAt: time.Now(),
	}

	s.storage.AddClickRecord(click)
	s.evaluateFAQ(faqID)

	return nil
}

func (s *Service) evaluateFAQ(faqID string) {
	faq, exists := s.storage.GetFAQ(faqID)
	if !exists {
		return
	}

	totalClicks, helpfulClicks := s.storage.GetClickStats(faqID)

	if totalClicks > 0 && faq.ViewCount >= HotViewThreshold {
		helpfulRatio := float64(helpfulClicks) / float64(totalClicks)
		if helpfulRatio < OptimizeThresholdRatio {
			faq.NeedOptimize = true
			s.storage.UpdateFAQ(faq)
		}
	}
}

func (s *Service) CreateCategory(req *protocol.CreateCategoryRequest) (*protocol.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	if req.ParentID != "" {
		if _, exists := s.storage.GetCategory(req.ParentID); !exists {
			return nil, fmt.Errorf("parent category not found")
		}
	}

	if _, exists := s.storage.GetCategoryByName(req.Name, req.ParentID); exists {
		return nil, fmt.Errorf("category with same name already exists in this level")
	}

	category := &protocol.Category{
		ID:       generateID(),
		ParentID: req.ParentID,
		Name:     req.Name,
		Children: []protocol.Category{},
	}

	s.storage.CreateCategory(category)
	return category, nil
}

func (s *Service) GetCategoryTree() (*protocol.CategoryTreeResponse, error) {
	allCategories := s.storage.ListAllCategories()

	categoryMap := make(map[string]*protocol.Category)
	for _, cat := range allCategories {
		categoryMap[cat.ID] = cat
	}

	var rootCategories []protocol.Category
	for _, cat := range allCategories {
		if cat.ParentID == "" {
			rootCategories = append(rootCategories, *buildCategoryTree(cat.ID, categoryMap))
		}
	}

	return &protocol.CategoryTreeResponse{
		Categories: rootCategories,
	}, nil
}

func buildCategoryTree(categoryID string, categoryMap map[string]*protocol.Category) *protocol.Category {
	cat := categoryMap[categoryID]
	if cat == nil {
		return nil
	}

	children := []protocol.Category{}
	for _, c := range categoryMap {
		if c.ParentID == categoryID {
			child := buildCategoryTree(c.ID, categoryMap)
			if child != nil {
				children = append(children, *child)
			}
		}
	}

	result := *cat
	result.Children = children
	return &result
}

func (s *Service) UpdateCategory(id string, req *protocol.UpdateCategoryRequest) (*protocol.Category, error) {
	category, exists := s.storage.GetCategory(id)
	if !exists {
		return nil, fmt.Errorf("category not found")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	if existing, exists := s.storage.GetCategoryByName(req.Name, category.ParentID); exists && existing.ID != id {
		return nil, fmt.Errorf("category with same name already exists in this level")
	}

	category.Name = req.Name
	s.storage.UpdateCategory(category)

	return category, nil
}

func (s *Service) DeleteCategory(id string) error {
	_, exists := s.storage.GetCategory(id)
	if !exists {
		return fmt.Errorf("category not found")
	}

	childCats := s.storage.GetChildCategories(id)
	if len(childCats) > 0 {
		return fmt.Errorf("cannot delete category with children")
	}

	faqs := s.storage.GetFAQsByCategory(id)
	if len(faqs) > 0 {
		return fmt.Errorf("cannot delete category with FAQs")
	}

	s.storage.DeleteCategory(id)
	return nil
}

func (s *Service) GetFAQHistory(faqID string) (*protocol.HistoryListResponse, error) {
	_, exists := s.storage.GetFAQ(faqID)
	if !exists {
		return nil, fmt.Errorf("FAQ not found")
	}

	records := s.storage.GetHistory(faqID)
	recordPtrs := make([]protocol.HistoryRecord, len(records))
	for i, r := range records {
		recordPtrs[i] = *r
	}

	return &protocol.HistoryListResponse{
		Records: recordPtrs,
		Total:   len(records),
	}, nil
}

func (s *Service) GetStatistics() (*protocol.FAQStatistics, error) {
	allFAQs := s.storage.ListAllFAQs()
	clickStats := s.storage.GetAllClickStats()

	stats := &protocol.FAQStatistics{
		TotalFAQ:     len(allFAQs),
		EnabledFAQ:   0,
		HotFAQ:       0,
		NeedOptimize: 0,
		TotalViews:   0,
		TotalClicks:  0,
	}

	for _, faq := range allFAQs {
		if faq.IsEnabled {
			stats.EnabledFAQ++
		}
		if faq.IsHot {
			stats.HotFAQ++
		}
		if faq.NeedOptimize {
			stats.NeedOptimize++
		}
		stats.TotalViews += faq.ViewCount
	}

	for _, stat := range clickStats {
		stats.TotalClicks += stat[0]
	}

	return stats, nil
}

func (s *Service) addHistory(faqID, action, oldContent, newContent, changedBy string) {
	record := &protocol.HistoryRecord{
		ID:         generateID(),
		FAQID:      faqID,
		Action:     action,
		OldContent: oldContent,
		NewContent: newContent,
		ChangedBy:  changedBy,
		ChangedAt:  time.Now(),
	}
	s.storage.AddHistoryRecord(record)
}

func faqToJSON(faq *protocol.FAQ) string {
	data, _ := json.Marshal(faq)
	return string(data)
}
