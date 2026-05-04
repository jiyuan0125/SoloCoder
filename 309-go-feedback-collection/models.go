package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	FeedbackTypeSuggestion  = "功能建议"
	FeedbackTypeBugReport   = "Bug报告"
	FeedbackTypeComplaint   = "体验投诉"

	StatusPending   = "待处理"
	StatusProcessing = "处理中"
	StatusClosed    = "已关闭"

	MaxDescriptionLength = 500
	MaxNoteLength        = 500
	MaxDailySubmissions  = 5
	MinRating            = 1
	MaxRating            = 5
	DefaultPageSize      = 20
	MaxPageSize          = 100
)

var ValidFeedbackTypes = map[string]bool{
	FeedbackTypeSuggestion: true,
	FeedbackTypeBugReport:  true,
	FeedbackTypeComplaint:  true,
}

var ValidStatuses = map[string]bool{
	StatusPending:   true,
	StatusProcessing: true,
	StatusClosed:    true,
}

type Feedback struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Type           string    `json:"type"`
	Rating         int       `json:"rating"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	InternalNote   string    `json:"internal_note,omitempty"`
	ProcessingNote string    `json:"processing_note,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Statistics struct {
	TotalCount     int                `json:"total_count"`
	ByType         map[string]TypeStat `json:"by_type"`
	AverageRating  float64            `json:"average_rating"`
}

type TypeStat struct {
	Count         int     `json:"count"`
	AverageRating float64 `json:"average_rating"`
}

type PaginatedResponse struct {
	Data       []Feedback `json:"data"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalCount int        `json:"total_count"`
	TotalPages int        `json:"total_pages"`
}

type Storage struct {
	Feedbacks []Feedback `json:"feedbacks"`
	mu        sync.RWMutex
	filePath  string
}

func NewStorage(filePath string) *Storage {
	s := &Storage{
		filePath: filePath,
	}
	s.load()
	return s
}

func (s *Storage) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.Feedbacks = []Feedback{}
			return
		}
		fmt.Printf("Warning: failed to load storage: %v\n", err)
		s.Feedbacks = []Feedback{}
		return
	}

	if err := json.Unmarshal(data, &s.Feedbacks); err != nil {
		fmt.Printf("Warning: failed to parse storage: %v\n", err)
		s.Feedbacks = []Feedback{}
	}
}

func (s *Storage) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.Feedbacks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) AddFeedback(feedback Feedback) error {
	s.mu.Lock()
	s.Feedbacks = append(s.Feedbacks, feedback)
	s.mu.Unlock()
	return s.save()
}

func (s *Storage) UpdateFeedback(feedback Feedback) error {
	s.mu.Lock()
	for i, f := range s.Feedbacks {
		if f.ID == feedback.ID {
			s.Feedbacks[i] = feedback
			break
		}
	}
	s.mu.Unlock()
	return s.save()
}

func (s *Storage) GetFeedbackByID(id string) (*Feedback, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.Feedbacks {
		if f.ID == id {
			return &f, true
		}
	}
	return nil, false
}

func (s *Storage) GetUserDailyCount(userID string, date time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	for _, f := range s.Feedbacks {
		if f.UserID == userID && f.CreatedAt.After(startOfDay) && f.CreatedAt.Before(endOfDay) {
			count++
		}
	}
	return count
}

func (s *Storage) GetAllFeedbacks() []Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Feedback, len(s.Feedbacks))
	copy(result, s.Feedbacks)
	return result
}

func (s *Storage) GetStatistics() Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := Statistics{
		TotalCount: len(s.Feedbacks),
		ByType: map[string]TypeStat{
			FeedbackTypeSuggestion: {},
			FeedbackTypeBugReport:  {},
			FeedbackTypeComplaint:  {},
		},
	}

	totalRating := 0
	typeRatings := map[string]int{
		FeedbackTypeSuggestion: 0,
		FeedbackTypeBugReport:  0,
		FeedbackTypeComplaint:  0,
	}

	for _, f := range s.Feedbacks {
		totalRating += f.Rating
		typeStat := stats.ByType[f.Type]
		typeStat.Count++
		stats.ByType[f.Type] = typeStat
		typeRatings[f.Type] += f.Rating
	}

	if stats.TotalCount > 0 {
		stats.AverageRating = float64(totalRating) / float64(stats.TotalCount)
	}

	for t, stat := range stats.ByType {
		if stat.Count > 0 {
			stat.AverageRating = float64(typeRatings[t]) / float64(stat.Count)
			stats.ByType[t] = stat
		}
	}

	return stats
}
