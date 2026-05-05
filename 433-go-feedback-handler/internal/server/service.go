package server

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-feedback-handler/pkg/protocol"
)

var (
	ErrFeedbackNotFound      = errors.New("feedback not found")
	ErrTagNotFound           = errors.New("tag not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrFeedbackMerged        = errors.New("feedback has been merged")
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateFeedback(req *protocol.CreateFeedbackRequest) (*protocol.CreateFeedbackResponse, error) {
	userLimit := s.store.GetOrCreateUserLimit(req.UserID)

	needsReview := userLimit.IsUnderReview

	existingFeedbacks := s.store.GetFeedbacksForMergeCheck(req.UserID, time.Now())
	var mergedFeedback *protocol.Feedback
	for _, fb := range existingFeedbacks {
		similarity := calculateSimilarity(req.Content, fb.Content)
		if similarity >= protocol.SimilarityThreshold {
			mergedFeedback = fb
			break
		}
	}

	if mergedFeedback != nil {
		newFB := s.createFeedbackObject(req)
		newFB.MergedInto = mergedFeedback.ID
		s.store.CreateFeedback(newFB)

		if mergedFeedback.MergedFrom == nil {
			mergedFeedback.MergedFrom = make([]string, 0)
		}
		mergedFeedback.MergedFrom = append(mergedFeedback.MergedFrom, newFB.ID)
		mergedFeedback.LastUpdatedAt = time.Now()
		s.store.UpdateFeedback(mergedFeedback)

		comment := &protocol.Comment{
			ID:           generateID(),
			FeedbackID:   mergedFeedback.ID,
			UserID:       "system",
			UserName:     "系统",
			Content:      fmt.Sprintf("用户提交了相似反馈，已自动合并。相似度: %.0f%%", calculateSimilarity(req.Content, mergedFeedback.Content)*100),
			IsFromSupport: true,
			CreatedAt:    time.Now(),
		}
		s.store.AddComment(mergedFeedback.ID, comment)

		return &protocol.CreateFeedbackResponse{
			Feedback:   newFB,
			IsMerged:   true,
			MergedInto: mergedFeedback.ID,
		}, nil
	}

	fb := s.createFeedbackObject(req)
	fb.InReviewQueue = needsReview

	if protocol.IsHighPriority(fb.Priority) {
		s.createHighPriorityNotification(fb)
	}

	s.store.CreateFeedback(fb)

	return &protocol.CreateFeedbackResponse{
		Feedback: fb,
		IsMerged: false,
		InReview: needsReview,
	}, nil
}

func (s *Service) createFeedbackObject(req *protocol.CreateFeedbackRequest) *protocol.Feedback {
	now := time.Now()
	priority := protocol.GetPriorityForFeedback(req.Type)

	return &protocol.Feedback{
		ID:            generateID(),
		UserID:        req.UserID,
		UserName:      req.UserName,
		Type:          req.Type,
		Content:       req.Content,
		Priority:      priority,
		Status:        protocol.StatusPending,
		Comments:      make([]*protocol.Comment, 0),
		Tags:          make([]*protocol.Tag, 0),
		TagIDs:        make([]string, 0),
		MergedFrom:    make([]string, 0),
		LastUpdatedAt: now,
		CreatedAt:     now,
	}
}

func (s *Service) GetFeedback(id string) (*protocol.Feedback, error) {
	fb := s.store.GetFeedback(id)
	if fb == nil {
		return nil, ErrFeedbackNotFound
	}
	return fb, nil
}

func (s *Service) AddComment(req *protocol.AddCommentRequest) error {
	fb := s.store.GetFeedback(req.FeedbackID)
	if fb == nil {
		return ErrFeedbackNotFound
	}

	if fb.MergedInto != "" {
		return ErrFeedbackMerged
	}

	if fb.Status == protocol.StatusClosed && !req.IsFromSupport {
		fb.Status = protocol.StatusPending
		fb.ClosedAt = time.Time{}
		fb.LastUpdatedAt = time.Now()
		s.store.UpdateFeedback(fb)

		notif := &protocol.Notification{
			ID:        generateID(),
			Type:      protocol.NotificationTypeReopened,
			TargetID:  fb.ID,
			Message:   fmt.Sprintf("反馈 %s 被用户追加评论，已自动重开", fb.ID[:8]),
			UserID:    fb.HandlerID,
			CreatedAt: time.Now(),
		}
		s.store.CreateNotification(notif)
	}

	comment := &protocol.Comment{
		ID:            generateID(),
		FeedbackID:    req.FeedbackID,
		UserID:        req.UserID,
		UserName:      req.UserName,
		Content:       req.Content,
		IsFromSupport: req.IsFromSupport,
		CreatedAt:     time.Now(),
	}

	if req.IsFromSupport && fb.FirstResponseAt.IsZero() {
		fb.FirstResponseAt = time.Now()
	}

	s.store.AddComment(req.FeedbackID, comment)
	return nil
}

func (s *Service) UpdateStatus(req *protocol.UpdateStatusRequest) error {
	fb := s.store.GetFeedback(req.FeedbackID)
	if fb == nil {
		return ErrFeedbackNotFound
	}

	if fb.MergedInto != "" {
		return ErrFeedbackMerged
	}

	if !protocol.IsValidStatusTransition(fb.Status, req.Status) {
		return ErrInvalidStatusTransition
	}

	oldStatus := fb.Status
	fb.Status = req.Status
	fb.HandlerID = req.HandlerID
	fb.HandlerName = req.HandlerName
	fb.LastUpdatedAt = time.Now()

	switch req.Status {
	case protocol.StatusResolved:
		fb.ResolvedAt = time.Now()
	case protocol.StatusClosed:
		fb.ClosedAt = time.Now()
		fb.IsInvalid = req.IsInvalid

		if req.IsInvalid {
			userLimit := s.store.GetOrCreateUserLimit(fb.UserID)
			userLimit.InvalidCloseCount++

			if userLimit.InvalidCloseCount >= protocol.MaxInvalidCloses {
				userLimit.IsUnderReview = true
				userLimit.ReviewStartedAt = time.Now()
			}

			s.store.UpdateUserLimit(userLimit)
		}
	case protocol.StatusProcessing:
		if oldStatus == protocol.StatusPending {
			fb.FirstResponseAt = time.Now()
		}
	}

	s.store.UpdateFeedback(fb)
	return nil
}

func (s *Service) AssignHandler(req *protocol.AssignHandlerRequest) error {
	fb := s.store.GetFeedback(req.FeedbackID)
	if fb == nil {
		return ErrFeedbackNotFound
	}

	fb.HandlerID = req.HandlerID
	fb.HandlerName = req.HandlerName
	fb.LastUpdatedAt = time.Now()

	s.store.UpdateFeedback(fb)
	return nil
}

func (s *Service) ListFeedbacks(req *protocol.ListFeedbackRequest) (*protocol.ListFeedbackResponse, error) {
	allFeedbacks := s.store.ListAllFeedbacks()

	filtered := make([]*protocol.Feedback, 0)
	for _, fb := range allFeedbacks {
		if fb.MergedInto != "" {
			continue
		}

		if req.Status != "" && fb.Status != req.Status {
			continue
		}
		if req.Type != "" && fb.Type != req.Type {
			continue
		}
		if req.Priority != "" && fb.Priority != req.Priority {
			continue
		}
		if req.UserID != "" && fb.UserID != req.UserID {
			continue
		}
		if req.HandlerID != "" && fb.HandlerID != req.HandlerID {
			continue
		}
		if req.TagID != "" && !containsString(fb.TagIDs, req.TagID) {
			continue
		}
		if req.InReview != nil && fb.InReviewQueue != *req.InReview {
			continue
		}

		filtered = append(filtered, fb)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			priorityOrder := map[protocol.FeedbackPriority]int{
				protocol.PriorityUrgent: 4,
				protocol.PriorityHigh:   3,
				protocol.PriorityMedium: 2,
				protocol.PriorityLow:    1,
			}
			return priorityOrder[filtered[i].Priority] > priorityOrder[filtered[j].Priority]
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	page := req.Page
	if page <= 0 {
		page = protocol.DefaultPage
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > protocol.MaxPageSize {
		pageSize = protocol.DefaultPageSize
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filtered) {
		return &protocol.ListFeedbackResponse{
			Feedbacks: []*protocol.Feedback{},
			Total:     len(filtered),
			Page:      page,
			PageSize:  pageSize,
		}, nil
	}

	if end > len(filtered) {
		end = len(filtered)
	}

	return &protocol.ListFeedbackResponse{
		Feedbacks: filtered[start:end],
		Total:     len(filtered),
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

func (s *Service) CreateTag(req *protocol.CreateTagRequest) (*protocol.Tag, error) {
	existing := s.store.GetTagByName(req.Name)
	if existing != nil {
		return existing, nil
	}

	tag := &protocol.Tag{
		ID:        generateID(),
		Name:      req.Name,
		Color:     req.Color,
		IsSystem:  false,
		CreatedAt: time.Now(),
	}

	s.store.CreateTag(tag)
	return tag, nil
}

func (s *Service) GetTag(id string) (*protocol.Tag, error) {
	tag := s.store.GetTag(id)
	if tag == nil {
		return nil, ErrTagNotFound
	}
	return tag, nil
}

func (s *Service) ListTags() []*protocol.Tag {
	return s.store.ListAllTags()
}

func (s *Service) AddTagToFeedback(req *protocol.AddTagRequest) error {
	fb := s.store.GetFeedback(req.FeedbackID)
	if fb == nil {
		return ErrFeedbackNotFound
	}

	tag := s.store.GetTag(req.TagID)
	if tag == nil {
		return ErrTagNotFound
	}

	if fb.MergedInto != "" {
		return ErrFeedbackMerged
	}

	s.store.AddTagToFeedback(req.FeedbackID, req.TagID)
	return nil
}

func (s *Service) RemoveTagFromFeedback(req *protocol.RemoveTagRequest) error {
	fb := s.store.GetFeedback(req.FeedbackID)
	if fb == nil {
		return ErrFeedbackNotFound
	}

	if fb.MergedInto != "" {
		return ErrFeedbackMerged
	}

	s.store.RemoveTagFromFeedback(req.FeedbackID, req.TagID)
	return nil
}

func (s *Service) createHighPriorityNotification(fb *protocol.Feedback) {
	notifPM := &protocol.Notification{
		ID:        generateID(),
		Type:      protocol.NotificationTypeHighPriority,
		TargetID:  fb.ID,
		Message:   fmt.Sprintf("收到高优先级反馈 [%s]: %s", fb.Type, truncateString(fb.Content, 50)),
		UserID:    "product_manager",
		CreatedAt: time.Now(),
	}
	s.store.CreateNotification(notifPM)

	notifTL := &protocol.Notification{
		ID:        generateID(),
		Type:      protocol.NotificationTypeHighPriority,
		TargetID:  fb.ID,
		Message:   fmt.Sprintf("收到高优先级反馈 [%s]: %s", fb.Type, truncateString(fb.Content, 50)),
		UserID:    "tech_lead",
		CreatedAt: time.Now(),
	}
	s.store.CreateNotification(notifTL)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (s *Service) GetNotifications(userID string) []*protocol.Notification {
	return s.store.ListNotificationsByUser(userID)
}

func (s *Service) MarkNotificationRead(id string) {
	s.store.MarkNotificationRead(id)
}

func (s *Service) GetUserLimit(userID string) *protocol.UserLimitInfo {
	return s.store.GetOrCreateUserLimit(userID)
}

func (s *Service) GetKPIStats(handlerID, period string) *protocol.KPIStats {
	feedbacks := s.store.GetFeedbacksByHandler(handlerID)

	var resolvedFeedbacks []*protocol.Feedback
	var closedFeedbacks []*protocol.Feedback
	var escalatedFeedbacks []*protocol.Feedback
	var totalResponseTime float64
	var totalResolutionTime float64
	var slaCompliantCount int
	var respondedCount int

	for _, fb := range feedbacks {
		if fb.Status == protocol.StatusResolved {
			resolvedFeedbacks = append(resolvedFeedbacks, fb)
		}
		if fb.Status == protocol.StatusClosed {
			closedFeedbacks = append(closedFeedbacks, fb)
		}
		if fb.IsEscalated {
			escalatedCount := len(escalatedFeedbacks)
			_ = escalatedCount
			escalatedFeedbacks = append(escalatedFeedbacks, fb)
		}

		if !fb.FirstResponseAt.IsZero() {
			respondedCount++
			responseTime := fb.FirstResponseAt.Sub(fb.CreatedAt).Hours()
			totalResponseTime += responseTime

			if fb.Type == protocol.FeedbackTypeComplaint {
				if responseTime <= float64(protocol.TwentyFourHours) {
					slaCompliantCount++
				}
			}
		}

		if !fb.ResolvedAt.IsZero() {
			resolutionTime := fb.ResolvedAt.Sub(fb.CreatedAt).Hours()
			totalResolutionTime += resolutionTime
		}
	}

	totalHandled := len(feedbacks)
	var avgResponseTime float64
	var avgResolutionTime float64
	var slaComplianceRate float64

	if respondedCount > 0 {
		avgResponseTime = roundFloat(totalResponseTime/float64(respondedCount), 2)
	}

	resolvedCount := len(resolvedFeedbacks)
	if resolvedCount > 0 {
		avgResolutionTime = roundFloat(totalResolutionTime/float64(resolvedCount), 2)
	}

	complaintCount := 0
	for _, fb := range feedbacks {
		if fb.Type == protocol.FeedbackTypeComplaint && !fb.FirstResponseAt.IsZero() {
			complaintCount++
		}
	}
	if complaintCount > 0 {
		slaComplianceRate = roundFloat(float64(slaCompliantCount)/float64(complaintCount)*100, 2)
	}

	return &protocol.KPIStats{
		HandlerID:           handlerID,
		HandlerName:         "",
		Period:              period,
		TotalHandled:        totalHandled,
		AvgResolutionTime:   avgResolutionTime,
		AvgFirstResponseTime: avgResponseTime,
		SLAComplianceRate:   slaComplianceRate,
		ResolvedCount:       resolvedCount,
		ClosedCount:         len(closedFeedbacks),
		EscalatedCount:      len(escalatedFeedbacks),
	}
}

func (s *Service) GenerateMonthlyReport(month string) (*protocol.MonthlyReport, error) {
	existing := s.store.GetReport(month)
	if existing != nil {
		return existing, nil
	}

	allFeedbacks := s.store.ListAllFeedbacks()

	var monthFeedbacks []*protocol.Feedback
	for _, fb := range allFeedbacks {
		fbMonth := getMonthKey(fb.CreatedAt)
		if fbMonth == month {
			monthFeedbacks = append(monthFeedbacks, fb)
		}
	}

	if len(monthFeedbacks) == 0 {
		return nil, errors.New("no feedbacks for this month")
	}

	byType := make(map[protocol.FeedbackType]int)
	byStatus := make(map[protocol.FeedbackStatus]int)
	byPriority := make(map[protocol.FeedbackPriority]int)
	tagCounts := make(map[string]int)
	tagNames := make(map[string]string)

	var totalResolutionTime float64
	var totalResponseTime float64
	var resolvedCount int
	var respondedCount int
	var slaCompliantCount int
	var complaintCount int
	var mergedCount int
	var escalatedCount int
	var invalidCount int

	for _, fb := range monthFeedbacks {
		byType[fb.Type]++
		byStatus[fb.Status]++
		byPriority[fb.Priority]++

		if len(fb.MergedFrom) > 0 {
			mergedCount += len(fb.MergedFrom)
		}
		if fb.IsEscalated {
			escalatedCount++
		}
		if fb.IsInvalid {
			invalidCount++
		}

		for _, tagID := range fb.TagIDs {
			tagCounts[tagID]++
			tag := s.store.GetTag(tagID)
			if tag != nil {
				tagNames[tagID] = tag.Name
			}
		}

		if !fb.ResolvedAt.IsZero() {
			resolvedCount++
			resolutionTime := fb.ResolvedAt.Sub(fb.CreatedAt).Hours()
			totalResolutionTime += resolutionTime
		}

		if !fb.FirstResponseAt.IsZero() {
			respondedCount++
			responseTime := fb.FirstResponseAt.Sub(fb.CreatedAt).Hours()
			totalResponseTime += responseTime

			if fb.Type == protocol.FeedbackTypeComplaint {
				complaintCount++
				if responseTime <= float64(protocol.TwentyFourHours) {
					slaCompliantCount++
				}
			}
		}
	}

	var avgResolutionTime float64
	var avgResponseTime float64
	var slaComplianceRate float64

	if resolvedCount > 0 {
		avgResolutionTime = roundFloat(totalResolutionTime/float64(resolvedCount), 2)
	}
	if respondedCount > 0 {
		avgResponseTime = roundFloat(totalResponseTime/float64(respondedCount), 2)
	}
	if complaintCount > 0 {
		slaComplianceRate = roundFloat(float64(slaCompliantCount)/float64(complaintCount)*100, 2)
	}

	topTags := make([]protocol.TagCount, 0, len(tagCounts))
	for tagID, count := range tagCounts {
		topTags = append(topTags, protocol.TagCount{
			TagID:   tagID,
			TagName: tagNames[tagID],
			Count:   count,
		})
	}
	sort.Slice(topTags, func(i, j int) bool {
		return topTags[i].Count > topTags[j].Count
	})
	if len(topTags) > 10 {
		topTags = topTags[:10]
	}

	trendAnalysis := s.generateTrendAnalysis(monthFeedbacks, byType, byPriority)
	suggestions := s.generateImprovementSuggestions(byType, slaComplianceRate, avgResponseTime, topTags)

	report := &protocol.MonthlyReport{
		Month:                  month,
		TotalFeedbacks:         len(monthFeedbacks),
		ByType:                 byType,
		ByStatus:               byStatus,
		ByPriority:             byPriority,
		AvgResolutionTime:      avgResolutionTime,
		AvgFirstResponseTime:   avgResponseTime,
		SLAComplianceRate:      slaComplianceRate,
		TopTags:                topTags,
		MergedCount:            mergedCount,
		EscalatedCount:         escalatedCount,
		InvalidCount:           invalidCount,
		TrendAnalysis:          trendAnalysis,
		ImprovementSuggestions: suggestions,
		CreatedAt:              time.Now(),
	}

	s.store.CreateReport(report)
	return report, nil
}

func (s *Service) generateTrendAnalysis(feedbacks []*protocol.Feedback, byType map[protocol.FeedbackType]int, byPriority map[protocol.FeedbackPriority]int) string {
	var parts []string

	total := len(feedbacks)
	if total == 0 {
		return "本月无反馈数据"
	}

	bugCount := byType[protocol.FeedbackTypeBug]
	complaintCount := byType[protocol.FeedbackTypeComplaint]
	featureCount := byType[protocol.FeedbackTypeFeature]

	parts = append(parts, fmt.Sprintf("本月共收到 %d 条反馈", total))

	if bugCount > total/3 {
		parts = append(parts, fmt.Sprintf("Bug报告占比较高（%.0f%%），建议关注产品稳定性", float64(bugCount)/float64(total)*100))
	}

	if complaintCount > 0 {
		parts = append(parts, fmt.Sprintf("收到 %d 条投诉，需重点关注用户体验", complaintCount))
	}

	if featureCount > total/2 {
		parts = append(parts, fmt.Sprintf("功能建议较多（%.0f%%），用户需求旺盛", float64(featureCount)/float64(total)*100))
	}

	highPrio := byPriority[protocol.PriorityHigh] + byPriority[protocol.PriorityUrgent]
	if highPrio > total/4 {
		parts = append(parts, fmt.Sprintf("高优先级反馈占比 %.0f%%，需加强资源调配", float64(highPrio)/float64(total)*100))
	}

	if len(parts) == 1 {
		parts = append(parts, "整体反馈分布较为均衡")
	}

	return strings.Join(parts, "。")
}

func (s *Service) generateImprovementSuggestions(byType map[protocol.FeedbackType]int, slaRate float64, avgRespTime float64, topTags []protocol.TagCount) []string {
	var suggestions []string

	if byType[protocol.FeedbackTypeBug] > 0 {
		suggestions = append(suggestions, "建议增加自动化测试覆盖率，减少Bug数量")
		suggestions = append(suggestions, "建立Bug快速响应机制，缩短修复周期")
	}

	if slaRate < 90 && byType[protocol.FeedbackTypeComplaint] > 0 {
		suggestions = append(suggestions, "投诉类反馈SLA达标率不足，建议增加客服轮值")
		suggestions = append(suggestions, "建立投诉预警机制，提前识别潜在风险")
	}

	if avgRespTime > 4 {
		suggestions = append(suggestions, fmt.Sprintf("平均响应时间 %.1f 小时，建议优化首响效率", avgRespTime))
	}

	if byType[protocol.FeedbackTypeFeature] > 0 {
		suggestions = append(suggestions, "建立需求收集和评审机制，定期梳理用户建议")
		suggestions = append(suggestions, "考虑设置产品反馈看板，让用户了解需求进度")
	}

	for _, tag := range topTags {
		if tag.TagName == protocol.SystemTagReproduced && tag.Count > 5 {
			suggestions = append(suggestions, "已复现问题较多，建议优先排期修复")
		}
		if tag.TagName == protocol.SystemTagNextVersionFix && tag.Count > 3 {
			suggestions = append(suggestions, "下版本修复的问题积累较多，需确认版本计划")
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "各项指标表现良好，建议保持当前运营节奏")
		suggestions = append(suggestions, "可考虑优化用户反馈入口，提升反馈便利性")
	}

	return suggestions
}

func (s *Service) CheckAndEscalateComplaints() int {
	feedbacks := s.store.GetFeedbacksByType(protocol.FeedbackTypeComplaint)
	escalated := 0

	for _, fb := range feedbacks {
		if fb.Status == protocol.StatusClosed || fb.Status == protocol.StatusResolved {
			continue
		}
		if fb.IsEscalated {
			continue
		}
		if !fb.FirstResponseAt.IsZero() {
			continue
		}

		hoursSinceCreated := hoursSince(fb.CreatedAt)
		if hoursSinceCreated >= float64(protocol.TwentyFourHours) {
			fb.IsEscalated = true
			fb.EscalatedAt = time.Now()
			s.store.UpdateFeedback(fb)

			notif := &protocol.Notification{
				ID:        generateID(),
				Type:      protocol.NotificationTypeEscalation,
				TargetID:  fb.ID,
				Message:   fmt.Sprintf("投诉反馈 %s 24小时未响应，已自动升级", fb.ID[:8]),
				UserID:    "product_manager",
				CreatedAt: time.Now(),
			}
			s.store.CreateNotification(notif)

			escalated++
		}
	}

	return escalated
}

func (s *Service) CheckAndSendReminders() int {
	feedbacks := s.store.ListAllFeedbacks()
	reminded := 0

	for _, fb := range feedbacks {
		if fb.Status == protocol.StatusClosed || fb.Status == protocol.StatusResolved {
			continue
		}
		if fb.HandlerID == "" {
			continue
		}

		daysSinceUpdate := daysSince(fb.LastUpdatedAt)
		if daysSinceUpdate >= float64(7) {
			if !fb.LastNotifiedAt.IsZero() {
				daysSinceLastNotify := daysSince(fb.LastNotifiedAt)
				if daysSinceLastNotify < 7 {
					continue
				}
			}

			fb.LastNotifiedAt = time.Now()
			s.store.UpdateFeedback(fb)

			notif := &protocol.Notification{
				ID:        generateID(),
				Type:      protocol.NotificationTypeReminder,
				TargetID:  fb.ID,
				Message:   fmt.Sprintf("反馈 %s 已7天未更新，请及时处理", fb.ID[:8]),
				UserID:    fb.HandlerID,
				CreatedAt: time.Now(),
			}
			s.store.CreateNotification(notif)

			reminded++
		}
	}

	return reminded
}

func (s *Service) CheckAndRestoreUserLimits() int {
	limits := s.store.ListAllUserLimits()
	restored := 0

	for _, limit := range limits {
		if !limit.IsUnderReview {
			continue
		}

		daysInReview := daysSince(limit.ReviewStartedAt)
		if daysInReview >= float64(protocol.ReviewDurationDays) {
			limit.IsUnderReview = false
			limit.InvalidCloseCount = 0
			s.store.UpdateUserLimit(limit)

			feedbacks := s.store.GetFeedbacksInReview()
			for _, fb := range feedbacks {
				if fb.UserID == limit.UserID {
					fb.InReviewQueue = false
					s.store.UpdateFeedback(fb)
				}
			}

			restored++
		}
	}

	return restored
}
