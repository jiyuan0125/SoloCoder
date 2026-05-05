package server

import (
	"strings"
	"time"

	"github.com/comment-moderator/pkg/common"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) SubmitComment(userID, content string) (*common.SubmitCommentResponse, error) {
	if userID == "" {
		return &common.SubmitCommentResponse{
			Success: false,
			Message: "用户ID不能为空",
		}, nil
	}

	if content == "" {
		return &common.SubmitCommentResponse{
			Success: false,
			Message: "评论内容不能为空",
		}, nil
	}

	user := s.store.GetOrCreateUser(userID)

	now := time.Now()

	s.refreshUserStatus(user, now)

	if s.shouldForceManualReview(user, now) {
		return s.submitWithStatus(userID, content, common.StatusManualReview, "")
	}

	hasSensitive, sensitiveWord := s.store.CheckSensitiveWords(content)
	if hasSensitive {
		return s.rejectComment(userID, content, sensitiveWord)
	}

	if len([]rune(content)) > common.MaxCommentLength {
		return s.submitWithStatus(userID, content, common.StatusManualReview, "")
	}

	return s.submitWithStatus(userID, content, common.StatusAutoApproved, "")
}

func (s *Service) refreshUserStatus(user *common.User, now time.Time) {
	if now.Sub(user.Last24hResetTime) > 24*time.Hour {
		user.Comments24h = 0
		user.Last24hResetTime = now
	}

	if user.UnderManualReview && now.After(user.ManualReviewEndTime) {
		user.UnderManualReview = false
		user.ConsecutiveRejects = 0
	}

	s.store.UpdateUser(user)
}

func (s *Service) shouldForceManualReview(user *common.User, now time.Time) bool {
	if user.UnderManualReview {
		return true
	}

	if user.Comments24h >= common.MaxComments24h {
		return true
	}

	return false
}

func (s *Service) submitWithStatus(userID, content string, status common.CommentStatus, rejectReason string) (*common.SubmitCommentResponse, error) {
	user := s.store.GetOrCreateUser(userID)
	user.Comments24h++
	s.store.UpdateUser(user)

	now := time.Now()
	comment := &common.Comment{
		ID:           common.GenerateID(),
		UserID:       userID,
		Content:      content,
		Status:       status,
		RejectReason: rejectReason,
		CreatedAt:    now,
		UpdatedAt:    now,
		EditDeadline: now.Add(time.Duration(common.EditWindowMinutes) * time.Minute),
		EditCount:    0,
		ReportedBy:   make([]string, 0),
		ReportCount:  0,
	}

	if status == common.StatusManualReview {
		moderator := s.store.GetLeastBusyModerator()
		if moderator != nil {
			comment.AssignedTo = moderator.ID
		}
	}

	if status == common.StatusAutoApproved {
		comment.Status = common.StatusPublished
	}

	s.store.AddComment(comment)

	var message string
	switch status {
	case common.StatusAutoApproved, common.StatusPublished:
		message = "评论已自动通过并发布"
	case common.StatusManualReview:
		message = "评论已进入人工审核队列"
	case common.StatusRejected:
		message = "评论已被拒绝"
	}

	return &common.SubmitCommentResponse{
		Success:   true,
		Message:   message,
		CommentID: comment.ID,
		Status:    comment.Status,
	}, nil
}

func (s *Service) rejectComment(userID, content string, sensitiveWord *common.SensitiveWord) (*common.SubmitCommentResponse, error) {
	user := s.store.GetOrCreateUser(userID)
	user.ConsecutiveRejects++
	user.LastRejectTime = time.Now()

	if user.ConsecutiveRejects >= common.ConsecutiveRejectsLimit {
		user.UnderManualReview = true
		user.ManualReviewEndTime = time.Now().Add(time.Duration(common.ManualReviewDays) * 24 * time.Hour)
	}

	s.store.UpdateUser(user)

	rejectReason := common.RejectReasonTemplates[sensitiveWord.Level]
	if rejectReason == "" {
		rejectReason = "您的评论包含违规内容，已被拒绝。"
	}

	now := time.Now()
	comment := &common.Comment{
		ID:             common.GenerateID(),
		UserID:         userID,
		Content:        content,
		Status:         common.StatusRejected,
		RejectReason:   rejectReason,
		CreatedAt:      now,
		UpdatedAt:      now,
		EditDeadline:   now.Add(time.Duration(common.EditWindowMinutes) * time.Minute),
		EditCount:      0,
		ReportedBy:     make([]string, 0),
		ReportCount:    0,
		ViolationLevel: sensitiveWord.Level,
	}

	s.store.AddComment(comment)

	s.addAuditLog("system", "auto_reject", comment.ID, "自动拒绝包含敏感词的评论，级别："+string(sensitiveWord.Level))

	return &common.SubmitCommentResponse{
		Success:   true,
		Message:   "评论已被拒绝",
		CommentID: comment.ID,
		Status:    common.StatusRejected,
	}, nil
}

func (s *Service) EditComment(commentID, userID, newContent string) (*common.EditCommentResponse, error) {
	if commentID == "" {
		return &common.EditCommentResponse{
			Success: false,
			Message: "评论ID不能为空",
		}, nil
	}

	if userID == "" {
		return &common.EditCommentResponse{
			Success: false,
			Message: "用户ID不能为空",
		}, nil
	}

	if newContent == "" {
		return &common.EditCommentResponse{
			Success: false,
			Message: "评论内容不能为空",
		}, nil
	}

	comment := s.store.GetComment(commentID)
	if comment == nil {
		return &common.EditCommentResponse{
			Success: false,
			Message: "评论不存在",
		}, nil
	}

	if comment.UserID != userID {
		return &common.EditCommentResponse{
			Success: false,
			Message: "只能编辑自己的评论",
		}, nil
	}

	if comment.EditCount >= common.MaxEditTimes {
		return &common.EditCommentResponse{
			Success: false,
			Message: "已达到最大编辑次数",
		}, nil
	}

	now := time.Now()
	if now.After(comment.EditDeadline) {
		return &common.EditCommentResponse{
			Success: false,
			Message: "编辑窗口已过期",
		}, nil
	}

	comment.Content = newContent
	comment.EditCount++
	comment.UpdatedAt = now
	comment.EditDeadline = now.Add(time.Duration(common.EditWindowMinutes) * time.Minute)

	hasSensitive, sensitiveWord := s.store.CheckSensitiveWords(newContent)
	if hasSensitive {
		comment.Status = common.StatusRejected
		comment.RejectReason = common.RejectReasonTemplates[sensitiveWord.Level]
		comment.ViolationLevel = sensitiveWord.Level

		user := s.store.GetOrCreateUser(userID)
		user.ConsecutiveRejects++
		user.LastRejectTime = now
		if user.ConsecutiveRejects >= common.ConsecutiveRejectsLimit {
			user.UnderManualReview = true
			user.ManualReviewEndTime = now.Add(time.Duration(common.ManualReviewDays) * 24 * time.Hour)
		}
		s.store.UpdateUser(user)

		s.addAuditLog("system", "auto_reject_edit", comment.ID, "编辑后的评论包含敏感词被拒绝")
	} else if len([]rune(newContent)) > common.MaxCommentLength {
		comment.Status = common.StatusManualReview
		moderator := s.store.GetLeastBusyModerator()
		if moderator != nil {
			comment.AssignedTo = moderator.ID
		}
	} else {
		comment.Status = common.StatusPending
	}

	s.store.UpdateComment(comment)

	return &common.EditCommentResponse{
		Success: true,
		Message: "评论已更新，重新进入审核流程",
		Status:  comment.Status,
	}, nil
}

func (s *Service) ReportComment(commentID, userID, reason string) (*common.ReportCommentResponse, error) {
	if commentID == "" {
		return &common.ReportCommentResponse{
			Success: false,
			Message: "评论ID不能为空",
		}, nil
	}

	if userID == "" {
		return &common.ReportCommentResponse{
			Success: false,
			Message: "用户ID不能为空",
		}, nil
	}

	comment := s.store.GetComment(commentID)
	if comment == nil {
		return &common.ReportCommentResponse{
			Success: false,
			Message: "评论不存在",
		}, nil
	}

	if comment.UserID == userID {
		return &common.ReportCommentResponse{
			Success: false,
			Message: "不能举报自己的评论",
		}, nil
	}

	for _, reporter := range comment.ReportedBy {
		if reporter == userID {
			return &common.ReportCommentResponse{
				Success: false,
				Message: "您已经举报过该评论",
			}, nil
		}
	}

	comment.ReportedBy = append(comment.ReportedBy, userID)
	comment.ReportCount++

	if comment.ReportCount >= common.ReportThreshold && comment.Status != common.StatusManualReview {
		comment.Status = common.StatusManualReview
		moderator := s.store.GetLeastBusyModerator()
		if moderator != nil {
			comment.AssignedTo = moderator.ID
		}
		s.addAuditLog("system", "report_trigger", comment.ID, "举报次数达到阈值，进入人工审核")
	}

	s.store.UpdateComment(comment)

	return &common.ReportCommentResponse{
		Success:     true,
		Message:     "举报已提交",
		ReportCount: comment.ReportCount,
	}, nil
}

func (s *Service) BatchApprove(moderatorID string, commentIDs []string) (*common.BatchActionResponse, error) {
	if len(commentIDs) > common.MaxBatchOperations {
		return &common.BatchActionResponse{
			Success: false,
			Message: "一次最多处理50条评论",
		}, nil
	}

	moderator := s.store.GetModerator(moderatorID)
	if moderator == nil {
		return &common.BatchActionResponse{
			Success: false,
			Message: "审核员不存在",
		}, nil
	}

	processed := 0
	failedIDs := make([]string, 0)
	today := common.GetTodayString()

	for _, commentID := range commentIDs {
		comment := s.store.GetComment(commentID)
		if comment == nil {
			failedIDs = append(failedIDs, commentID)
			continue
		}

		if comment.Status != common.StatusManualReview && comment.Status != common.StatusPending {
			failedIDs = append(failedIDs, commentID)
			continue
		}

		comment.Status = common.StatusPublished
		comment.UpdatedAt = time.Now()
		s.store.UpdateComment(comment)

		moderator.TotalCount++
		if moderator.DailyCount == nil {
			moderator.DailyCount = make(map[string]int)
		}
		moderator.DailyCount[today]++
		s.store.AddModerator(moderator)

		s.addAuditLog(moderatorID, "approve", commentID, "审核通过")
		processed++
	}

	return &common.BatchActionResponse{
		Success:   true,
		Message:   "批量操作完成",
		Processed: processed,
		FailedIDs: failedIDs,
	}, nil
}

func (s *Service) BatchReject(moderatorID string, commentIDs []string, reason string) (*common.BatchActionResponse, error) {
	if len(commentIDs) > common.MaxBatchOperations {
		return &common.BatchActionResponse{
			Success: false,
			Message: "一次最多处理50条评论",
		}, nil
	}

	moderator := s.store.GetModerator(moderatorID)
	if moderator == nil {
		return &common.BatchActionResponse{
			Success: false,
			Message: "审核员不存在",
		}, nil
	}

	processed := 0
	failedIDs := make([]string, 0)
	today := common.GetTodayString()

	for _, commentID := range commentIDs {
		comment := s.store.GetComment(commentID)
		if comment == nil {
			failedIDs = append(failedIDs, commentID)
			continue
		}

		if comment.Status != common.StatusManualReview && comment.Status != common.StatusPending {
			failedIDs = append(failedIDs, commentID)
			continue
		}

		comment.Status = common.StatusRejected
		comment.RejectReason = reason
		comment.UpdatedAt = time.Now()
		s.store.UpdateComment(comment)

		user := s.store.GetOrCreateUser(comment.UserID)
		user.ConsecutiveRejects++
		user.LastRejectTime = time.Now()
		if user.ConsecutiveRejects >= common.ConsecutiveRejectsLimit {
			user.UnderManualReview = true
			user.ManualReviewEndTime = time.Now().Add(time.Duration(common.ManualReviewDays) * 24 * time.Hour)
		}
		s.store.UpdateUser(user)

		moderator.TotalCount++
		if moderator.DailyCount == nil {
			moderator.DailyCount = make(map[string]int)
		}
		moderator.DailyCount[today]++
		s.store.AddModerator(moderator)

		s.addAuditLog(moderatorID, "reject", commentID, "审核拒绝，原因："+reason)
		processed++
	}

	return &common.BatchActionResponse{
		Success:   true,
		Message:   "批量操作完成",
		Processed: processed,
		FailedIDs: failedIDs,
	}, nil
}

func (s *Service) AddSensitiveWord(adminID, word string, level common.SensitiveWordLevel) (*common.AddSensitiveWordResponse, error) {
	if strings.TrimSpace(word) == "" {
		return &common.AddSensitiveWordResponse{
			Success: false,
			Message: "敏感词不能为空",
		}, nil
	}

	sw := &common.SensitiveWord{
		Word:  word,
		Level: level,
	}

	s.store.AddSensitiveWord(sw)
	s.addAuditLog(adminID, "add_sensitive", "", "添加敏感词："+word+"，级别："+string(level))

	return &common.AddSensitiveWordResponse{
		Success: true,
		Message: "敏感词添加成功",
	}, nil
}

func (s *Service) RemoveSensitiveWord(adminID, word string) (*common.RemoveSensitiveWordResponse, error) {
	if s.store.RemoveSensitiveWord(word) {
		s.addAuditLog(adminID, "remove_sensitive", "", "删除敏感词："+word)
		return &common.RemoveSensitiveWordResponse{
			Success: true,
			Message: "敏感词删除成功",
		}, nil
	}

	return &common.RemoveSensitiveWordResponse{
		Success: false,
		Message: "敏感词不存在",
	}, nil
}

func (s *Service) GetPendingComments(moderatorID string, limit, offset int) (*common.GetPendingCommentsResponse, error) {
	if moderatorID == "" {
		return &common.GetPendingCommentsResponse{
			Success: false,
			Message: "需要指定审核员ID",
		}, nil
	}

	moderator := s.store.GetModerator(moderatorID)
	if moderator == nil {
		return &common.GetPendingCommentsResponse{
			Success: false,
			Message: "审核员不存在",
		}, nil
	}

	allPending := s.store.GetPendingComments(limit, offset)
	total := s.store.GetPendingCount()

	return &common.GetPendingCommentsResponse{
		Success:  true,
		Message:  "获取成功",
		Comments: allPending,
		Total:    total,
		HasMore:  offset+limit < total,
	}, nil
}

func (s *Service) ViewRejectedComment(userID, commentID string) (*common.ViewRejectedCommentResponse, error) {
	comment := s.store.GetComment(commentID)
	if comment == nil {
		return &common.ViewRejectedCommentResponse{
			Success: false,
			Message: "评论不存在",
		}, nil
	}

	if comment.UserID != userID {
		return &common.ViewRejectedCommentResponse{
			Success: false,
			Message: "只能查看自己的评论",
		}, nil
	}

	if comment.Status != common.StatusRejected {
		return &common.ViewRejectedCommentResponse{
			Success: false,
			Message: "该评论未被拒绝",
		}, nil
	}

	return &common.ViewRejectedCommentResponse{
		Success:      true,
		Message:      "获取成功",
		RejectReason: comment.RejectReason,
	}, nil
}

func (s *Service) RegisterModerator(adminID, name string) (*common.RegisterModeratorResponse, error) {
	id := common.GenerateID()
	moderator := &common.Moderator{
		ID:          id,
		Name:        name,
		DailyCount:  make(map[string]int),
		TotalCount:  0,
		AssignedIDs: make([]string, 0),
	}

	s.store.AddModerator(moderator)
	s.addAuditLog(adminID, "register_moderator", id, "注册审核员："+name)

	return &common.RegisterModeratorResponse{
		Success:     true,
		Message:     "审核员注册成功",
		ModeratorID: id,
	}, nil
}

func (s *Service) GetSensitiveWords() (*common.GetSensitiveWordsResponse, error) {
	words := s.store.GetSensitiveWords()
	return &common.GetSensitiveWordsResponse{
		Success: true,
		Message: "获取成功",
		Words:   words,
	}, nil
}

func (s *Service) GetAuditLogs() (*common.GetAuditLogsResponse, error) {
	logs := s.store.GetAuditLogs()
	return &common.GetAuditLogsResponse{
		Success: true,
		Message: "获取成功",
		Logs:    logs,
	}, nil
}

func (s *Service) GetModeratorStats() (*common.GetModeratorStatsResponse, error) {
	moderators := s.store.GetAllModerators()
	today := common.GetTodayString()

	stats := make(map[string]common.ModeratorStats)
	for id, m := range moderators {
		stats[id] = common.ModeratorStats{
			Name:       m.Name,
			DailyCount: m.DailyCount[today],
			TotalCount: m.TotalCount,
		}
	}

	return &common.GetModeratorStatsResponse{
		Success:    true,
		Message:    "获取成功",
		Moderators: stats,
	}, nil
}

func (s *Service) addAuditLog(moderatorID, action, commentID, details string) {
	log := &common.AuditLog{
		ID:          common.GenerateID(),
		ModeratorID: moderatorID,
		Action:      action,
		CommentID:   commentID,
		Timestamp:   time.Now(),
		Details:     details,
	}
	s.store.AddAuditLog(log)
}
