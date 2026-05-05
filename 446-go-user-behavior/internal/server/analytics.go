package server

import (
	"sort"
	"time"
	"userbehavior/internal/shared"
)

type AnalyticsEngine struct {
	store *Store
}

func NewAnalyticsEngine(store *Store) *AnalyticsEngine {
	return &AnalyticsEngine{store: store}
}

func (e *AnalyticsEngine) matchesStep(behavior *shared.Behavior, step shared.FunnelStep) bool {
	if behavior.Type != step.BehaviorType {
		return false
	}

	switch behavior.Type {
	case shared.BehaviorTypePageView:
		if behavior.PageView != nil && step.PageKey != "" {
			return behavior.PageView.PageKey == step.PageKey
		}
	case shared.BehaviorTypeButtonClick:
		if behavior.ButtonClick != nil && step.ButtonID != "" {
			return behavior.ButtonClick.ButtonID == step.ButtonID
		}
	case shared.BehaviorTypeFeatureUse:
		if behavior.FeatureUse != nil && step.FeatureName != "" {
			return behavior.FeatureUse.FeatureName == step.FeatureName
		}
	}

	return true
}

func (e *AnalyticsEngine) AnalyzeFunnel(steps []shared.FunnelStep) (*shared.FunnelResult, error) {
	if len(steps) == 0 {
		return nil, nil
	}

	if len(steps) > shared.MaxFunnelSteps {
		steps = steps[:shared.MaxFunnelSteps]
	}

	users := e.store.GetAllUsers()

	userStepCompletion := make(map[string]int)

	for _, userID := range users {
		behaviors := e.store.GetUserBehaviors(userID)

		sortedBehaviors := make([]*shared.Behavior, len(behaviors))
		copy(sortedBehaviors, behaviors)
		sort.Slice(sortedBehaviors, func(i, j int) bool {
			return sortedBehaviors[i].Timestamp.Before(sortedBehaviors[j].Timestamp)
		})

		currentStep := 0
		for _, behavior := range sortedBehaviors {
			if behavior.IsFiltered {
				continue
			}

			if currentStep < len(steps) && e.matchesStep(behavior, steps[currentStep]) {
				currentStep++
			}
		}

		userStepCompletion[userID] = currentStep
	}

	result := &shared.FunnelResult{
		Steps: make([]shared.FunnelStepResult, len(steps)),
	}

	prevCount := len(users)
	prevAvailable := true

	for i := 0; i < len(steps); i++ {
		stepResult := shared.FunnelStepResult{
			Name:        steps[i].Name,
			IsAvailable: prevAvailable,
		}

		if prevAvailable {
			count := 0
			for _, completedSteps := range userStepCompletion {
				if completedSteps > i {
					count++
				}
			}

			stepResult.UserCount = count

			if i == 0 {
				if prevCount > 0 {
					stepResult.ConversionRate = float64(count) / float64(prevCount) * 100
					stepResult.ChurnRate = 100 - stepResult.ConversionRate
				}
			} else {
				if prevCount > 0 {
					stepResult.ConversionRate = float64(count) / float64(prevCount) * 100
					stepResult.ChurnRate = 100 - stepResult.ConversionRate
				}
			}

			if count == 0 {
				prevAvailable = false
			}

			prevCount = count
		}

		result.Steps[i] = stepResult
	}

	return result, nil
}

func (e *AnalyticsEngine) AnalyzeRetention(startDate, endDate time.Time) (*shared.RetentionResult, error) {
	day1Start := startDate.Truncate(24 * time.Hour)
	day1End := day1Start.Add(24 * time.Hour)

	day1Users := make(map[string]bool)

	day1Behaviors := e.store.GetBehaviorsByDateRange(day1Start, day1End)
	for _, b := range day1Behaviors {
		if !b.IsFiltered {
			day1Users[b.UserID] = true
		}
	}

	totalDay1Users := len(day1Users)
	if totalDay1Users == 0 {
		return &shared.RetentionResult{
			Day1Retention:  0,
			Day7Retention:  0,
			Day30Retention: 0,
		}, nil
	}

	day1RetentionCount := 0
	day7RetentionCount := 0
	day30RetentionCount := 0

	for userID := range day1Users {
		behaviors := e.store.GetUserBehaviors(userID)

		hasDay1Return := false
		hasDay7Return := false
		hasDay30Return := false

		for _, b := range behaviors {
			if b.IsFiltered {
				continue
			}

			daysSinceDay1 := int(b.Timestamp.Sub(day1Start).Hours() / 24)

			if daysSinceDay1 == 1 {
				hasDay1Return = true
			}
			if daysSinceDay1 >= 1 && daysSinceDay1 <= 7 {
				hasDay7Return = true
			}
			if daysSinceDay1 >= 1 && daysSinceDay1 <= 30 {
				hasDay30Return = true
			}
		}

		if hasDay1Return {
			day1RetentionCount++
		}
		if hasDay7Return {
			day7RetentionCount++
		}
		if hasDay30Return {
			day30RetentionCount++
		}
	}

	return &shared.RetentionResult{
		Day1Retention:  float64(day1RetentionCount) / float64(totalDay1Users) * 100,
		Day7Retention:  float64(day7RetentionCount) / float64(totalDay1Users) * 100,
		Day30Retention: float64(day30RetentionCount) / float64(totalDay1Users) * 100,
	}, nil
}

func (e *AnalyticsEngine) AnalyzePaths(fromPage, toPage string) (*shared.PathResult, error) {
	type userPathInfo struct {
		pages     []string
		timestamps []time.Time
	}

	userPaths := make(map[string]*userPathInfo)

	users := e.store.GetAllUsers()
	for _, userID := range users {
		behaviors := e.store.GetUserBehaviors(userID)

		userPageViews := make([]*shared.Behavior, 0)
		for _, b := range behaviors {
			if !b.IsFiltered && b.Type == shared.BehaviorTypePageView && b.PageView != nil {
				userPageViews = append(userPageViews, b)
			}
		}

		sort.Slice(userPageViews, func(i, j int) bool {
			return userPageViews[i].Timestamp.Before(userPageViews[j].Timestamp)
		})

		for _, b := range userPageViews {
			if _, exists := userPaths[userID]; !exists {
				userPaths[userID] = &userPathInfo{
					pages:      make([]string, 0),
					timestamps: make([]time.Time, 0),
				}
			}

			pathInfo := userPaths[userID]
			pathInfo.pages = append(pathInfo.pages, b.PageView.PageKey)
			pathInfo.timestamps = append(pathInfo.timestamps, b.Timestamp)
		}
	}

	pathCounts := make(map[string]int)
	totalPaths := 0

	for userID, pathInfo := range userPaths {
		pages := pathInfo.pages
		if len(pages) < 2 {
			continue
		}

		fromIndices := make([]int, 0)
		toIndices := make([]int, 0)

		for i, page := range pages {
			if page == fromPage {
				fromIndices = append(fromIndices, i)
			}
			if page == toPage {
				toIndices = append(toIndices, i)
			}
		}

		for _, fromIdx := range fromIndices {
			for _, toIdx := range toIndices {
				if toIdx > fromIdx {
					pathKey := ""
					for i := fromIdx; i <= toIdx; i++ {
						if i > fromIdx {
							pathKey += "->"
						}
						pathKey += pages[i]
					}

					pathCounts[pathKey]++
					totalPaths++
					break
				}
			}
		}

		_ = userID
	}

	type pathWithCount struct {
		path  []string
		count int
	}

	paths := make([]pathWithCount, 0, len(pathCounts))
	for pathKey, count := range pathCounts {
		pages := splitPath(pathKey)
		paths = append(paths, pathWithCount{
			path:  pages,
			count: count,
		})
	}

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].count > paths[j].count
	})

	resultPaths := make([]shared.PathItem, 0)
	for i := 0; i < len(paths) && i < 10; i++ {
		percentage := 0.0
		if totalPaths > 0 {
			percentage = float64(paths[i].count) / float64(totalPaths) * 100
		}

		resultPaths = append(resultPaths, shared.PathItem{
			Path:       paths[i].path,
			UserCount:  paths[i].count,
			Percentage: percentage,
		})
	}

	return &shared.PathResult{
		Paths: resultPaths,
	}, nil
}

func splitPath(pathKey string) []string {
	result := make([]string, 0)
	current := ""

	for i := 0; i < len(pathKey); i++ {
		if i+2 < len(pathKey) && pathKey[i] == '-' && pathKey[i+1] == '>' {
			result = append(result, current)
			current = ""
			i += 2
			continue
		}
		current += string(pathKey[i])
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

func (e *AnalyticsEngine) GenerateUserProfile(userID string) (*shared.UserProfile, error) {
	behaviors := e.store.GetUserBehaviors(userID)
	sessions := e.store.GetUserSessions(userID)

	if len(behaviors) == 0 {
		return nil, nil
	}

	interestTags := make(map[string]int)
	var firstActive, lastActive time.Time

	for _, b := range behaviors {
		if b.IsFiltered {
			continue
		}

		if firstActive.IsZero() || b.Timestamp.Before(firstActive) {
			firstActive = b.Timestamp
		}
		if lastActive.IsZero() || b.Timestamp.After(lastActive) {
			lastActive = b.Timestamp
		}

		switch b.Type {
		case shared.BehaviorTypePageView:
			if b.PageView != nil {
				interestTags[b.PageView.PageKey]++
			}
		case shared.BehaviorTypeButtonClick:
			if b.ButtonClick != nil {
				interestTags["button_"+b.ButtonClick.ButtonID]++
			}
		case shared.BehaviorTypeFeatureUse:
			if b.FeatureUse != nil {
				interestTags["feature_"+b.FeatureUse.FeatureName]++
			}
		}
	}

	totalDays := int(lastActive.Sub(firstActive).Hours()/24) + 1
	behaviorCount := 0
	for _, b := range behaviors {
		if !b.IsFiltered {
			behaviorCount++
		}
	}

	avgBehaviorsPerDay := float64(behaviorCount) / float64(totalDays)
	activityLevel := "low"
	if avgBehaviorsPerDay > 5 {
		activityLevel = "high"
	} else if avgBehaviorsPerDay > 1 {
		activityLevel = "medium"
	}

	return &shared.UserProfile{
		UserID:       userID,
		InterestTags: interestTags,
		ActivityLevel: activityLevel,
		LastActive:   lastActive,
		FirstActive:  firstActive,
		TotalSessions: len(sessions),
	}, nil
}

func (e *AnalyticsEngine) GetRealtimeStats() (*shared.RealtimeStats, error) {
	activeSessions := e.store.GetActiveSessions()

	onlineUsers := make(map[string]bool)
	for _, session := range activeSessions {
		onlineUsers[session.UserID] = true
	}

	pageViews := make(map[string]int)
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	recentBehaviors := e.store.GetBehaviorsByDateRange(fiveMinutesAgo, now)
	for _, b := range recentBehaviors {
		if b.IsFiltered {
			continue
		}
		if b.Type == shared.BehaviorTypePageView && b.PageView != nil {
			pageViews[b.PageView.PageKey]++
		}
	}

	type pageCount struct {
		pageKey string
		count   int
	}

	pages := make([]pageCount, 0, len(pageViews))
	for pageKey, count := range pageViews {
		pages = append(pages, pageCount{pageKey: pageKey, count: count})
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].count > pages[j].count
	})

	rankings := make([]shared.PageRanking, 0)
	for i, p := range pages {
		if i >= 10 {
			break
		}
		rankings = append(rankings, shared.PageRanking{
			PageKey:  p.pageKey,
			UserCount: p.count,
			Rank:     i + 1,
		})
	}

	return &shared.RealtimeStats{
		OnlineUserCount: len(onlineUsers),
		ActivePages:     rankings,
	}, nil
}
