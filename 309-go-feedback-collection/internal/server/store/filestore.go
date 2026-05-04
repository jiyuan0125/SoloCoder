package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"feedback-system/internal/server/model"
	"feedback-system/pkg/common"
)

const (
	feedbacksFile = "feedbacks.json"
	dailyCountsFile = "daily_counts.json"
)

type FileStore struct {
	dataDir     string
	mu          sync.RWMutex
	feedbacks   map[string]*model.Feedback
	dailyCounts map[string]*model.UserDailyCount
}

func NewFileStore(dataDir string) (*FileStore, error) {
	fs := &FileStore{
		dataDir:     dataDir,
		feedbacks:   make(map[string]*model.Feedback),
		dailyCounts: make(map[string]*model.UserDailyCount),
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := fs.load(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return fs, nil
}

func (fs *FileStore) load() error {
	feedbacksPath := filepath.Join(fs.dataDir, feedbacksFile)
	if data, err := os.ReadFile(feedbacksPath); err == nil {
		var feedbacks []*model.Feedback
		if err := json.Unmarshal(data, &feedbacks); err != nil {
			return fmt.Errorf("failed to unmarshal feedbacks: %w", err)
		}
		for _, f := range feedbacks {
			fs.feedbacks[f.ID] = f
		}
	}

	dailyCountsPath := filepath.Join(fs.dataDir, dailyCountsFile)
	if data, err := os.ReadFile(dailyCountsPath); err == nil {
		var dailyCounts []*model.UserDailyCount
		if err := json.Unmarshal(data, &dailyCounts); err != nil {
			return fmt.Errorf("failed to unmarshal daily counts: %w", err)
		}
		for _, dc := range dailyCounts {
			key := fmt.Sprintf("%s:%s", dc.UserID, dc.Date)
			fs.dailyCounts[key] = dc
		}
	}

	return nil
}

func (fs *FileStore) save() error {
	feedbacks := make([]*model.Feedback, 0, len(fs.feedbacks))
	for _, f := range fs.feedbacks {
		feedbacks = append(feedbacks, f)
	}

	data, err := json.MarshalIndent(feedbacks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedbacks: %w", err)
	}

	feedbacksPath := filepath.Join(fs.dataDir, feedbacksFile)
	if err := os.WriteFile(feedbacksPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write feedbacks: %w", err)
	}

	dailyCounts := make([]*model.UserDailyCount, 0, len(fs.dailyCounts))
	for _, dc := range fs.dailyCounts {
		dailyCounts = append(dailyCounts, dc)
	}

	data, err = json.MarshalIndent(dailyCounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal daily counts: %w", err)
	}

	dailyCountsPath := filepath.Join(fs.dataDir, dailyCountsFile)
	if err := os.WriteFile(dailyCountsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write daily counts: %w", err)
	}

	return nil
}

func (fs *FileStore) CreateFeedback(feedback *model.Feedback) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.feedbacks[feedback.ID] = feedback
	return fs.save()
}

func (fs *FileStore) GetFeedback(id string) (*model.Feedback, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	feedback, ok := fs.feedbacks[id]
	if !ok {
		return nil, ErrFeedbackNotFound
	}
	return feedback, nil
}

func (fs *FileStore) ListFeedbacks(page, pageSize int, filterType common.FeedbackType, filterStatus common.FeedbackStatus) ([]*model.Feedback, int, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var filtered []*model.Feedback
	for _, f := range fs.feedbacks {
		matchType := filterType == "" || f.Type == filterType
		matchStatus := filterStatus == "" || f.Status == filterStatus
		if matchType && matchStatus {
			filtered = append(filtered, f)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)

	start := (page - 1) * pageSize
	if start >= total {
		return []*model.Feedback{}, total, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (fs *FileStore) UpdateFeedback(feedback *model.Feedback) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, ok := fs.feedbacks[feedback.ID]; !ok {
		return ErrFeedbackNotFound
	}

	feedback.UpdatedAt = time.Now()
	fs.feedbacks[feedback.ID] = feedback
	return fs.save()
}

func (fs *FileStore) GetUserDailyCount(userID string, date string) (int, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, date)
	if dc, ok := fs.dailyCounts[key]; ok {
		return dc.Count, nil
	}
	return 0, nil
}

func (fs *FileStore) IncrementUserDailyCount(userID string, date string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, date)
	if dc, ok := fs.dailyCounts[key]; ok {
		dc.Count++
		dc.UpdatedAt = time.Now()
	} else {
		fs.dailyCounts[key] = &model.UserDailyCount{
			UserID:    userID,
			Date:      date,
			Count:     1,
			UpdatedAt: time.Now(),
		}
	}
	return fs.save()
}

func (fs *FileStore) GetStatistics() ([]common.TypeStats, int, float64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	typeStats := make(map[common.FeedbackType]struct {
		count      int
		totalRating int
	})

	var totalCount int
	var totalRatingSum int

	for _, f := range fs.feedbacks {
		totalCount++
		totalRatingSum += f.Rating

		stats := typeStats[f.Type]
		stats.count++
		stats.totalRating += f.Rating
		typeStats[f.Type] = stats
	}

	var result []common.TypeStats
	for _, t := range []common.FeedbackType{
		common.FeedbackTypeSuggestion,
		common.FeedbackTypeBug,
		common.FeedbackTypeComplaint,
	} {
		stats := typeStats[t]
		avgRating := 0.0
		if stats.count > 0 {
			avgRating = float64(stats.totalRating) / float64(stats.count)
		}
		result = append(result, common.TypeStats{
			Type:          t,
			TypeName:      t.DisplayName(),
			Count:         stats.count,
			AverageRating: avgRating,
		})
	}

	overallAvg := 0.0
	if totalCount > 0 {
		overallAvg = float64(totalRatingSum) / float64(totalCount)
	}

	return result, totalCount, overallAvg, nil
}
