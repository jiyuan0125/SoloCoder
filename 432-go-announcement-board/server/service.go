package main

import (
	"announcement-board/common"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

const MaxPinnedCount = 5
const MaxAttachments = 5

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isAdmin(userID string) bool {
	user, ok := globalStore.GetUser(userID)
	return ok && user.Role == common.RoleAdmin
}

func isManager(userID string) bool {
	user, ok := globalStore.GetUser(userID)
	return ok && user.Role == common.RoleManager
}

func checkScope(ann *common.Announcement, user *common.User) bool {
	if ann.Scope == common.ScopeAll {
		return true
	}
	for _, deptID := range ann.TargetDeptIDs {
		if deptID == user.DeptID {
			return true
		}
	}
	return false
}

func isEffectiveNow(ann *common.Announcement) bool {
	now := time.Now()
	return !ann.EffectiveStart.After(now) && !ann.EffectiveEnd.Before(now)
}

func archiveExpiredAnnouncements() {
	now := time.Now()
	for _, ann := range globalStore.GetAllAnnouncements() {
		if ann.Status == common.StatusPublished && ann.EffectiveEnd.Before(now) {
			ann.Status = common.StatusArchived
			ann.UpdatedAt = now
			globalStore.UpdateAnnouncement(ann)
		}
	}
}

func publishScheduledAnnouncements() {
	now := time.Now()
	for _, ann := range globalStore.GetAllAnnouncements() {
		if ann.Status == common.StatusDraft && ann.ApprovalRecord == nil {
			if !ann.EffectiveStart.After(now) && !ann.EffectiveEnd.Before(now) {
				ann.Status = common.StatusPublished
				ann.UpdatedAt = now
				globalStore.UpdateAnnouncement(ann)
			}
		}
	}
}

func ProcessScheduledTasks() {
	archiveExpiredAnnouncements()
	publishScheduledAnnouncements()
}

func CreateAnnouncement(userID string, req *common.CreateAnnouncementRequest) (*common.Announcement, error) {
	if !isAdmin(userID) {
		return nil, &APIError{Code: 403, Message: "只有管理员可以创建公告"}
	}

	if req.Title == "" {
		return nil, &APIError{Code: 400, Message: "公告标题不能为空"}
	}
	if req.Content == "" {
		return nil, &APIError{Code: 400, Message: "公告内容不能为空"}
	}
	if req.EffectiveEnd.Before(req.EffectiveStart) {
		return nil, &APIError{Code: 400, Message: "生效结束时间不能早于开始时间"}
	}
	if req.Scope == common.ScopeDept && len(req.TargetDeptIDs) == 0 {
		return nil, &APIError{Code: 400, Message: "指定部门范围时必须选择目标部门"}
	}
	if len(req.Attachments) > MaxAttachments {
		return nil, &APIError{Code: 400, Message: "每个公告最多允许5个附件"}
	}

	for _, deptID := range req.TargetDeptIDs {
		if _, ok := globalStore.GetDepartment(deptID); !ok {
			return nil, &APIError{Code: 400, Message: "部门不存在: " + deptID}
		}
	}

	now := time.Now()
	ann := &common.Announcement{
		ID:             "ann_" + generateID(),
		Title:          req.Title,
		Content:        req.Content,
		Scope:          req.Scope,
		TargetDeptIDs:  req.TargetDeptIDs,
		IsPinned:       false,
		EffectiveStart: req.EffectiveStart,
		EffectiveEnd:   req.EffectiveEnd,
		Priority:       req.Priority,
		Status:         common.StatusDraft,
		CreatedBy:      userID,
		CreatedAt:      now,
		UpdatedAt:      now,
		Attachments:    make([]common.Attachment, 0),
		ChangeHistory:  make([]common.ChangeHistory, 0),
		ViewCount:      0,
		ViewedUserIDs:  make([]string, 0),
	}

	for _, att := range req.Attachments {
		ann.Attachments = append(ann.Attachments, common.Attachment{
			ID:       "att_" + generateID(),
			FileName: att.FileName,
			FileSize: att.FileSize,
			UploadAt: now,
		})
	}

	if req.Priority == common.PriorityUrgent {
		ann.Status = common.StatusPublished
	} else {
		if !req.EffectiveStart.After(now) && !req.EffectiveEnd.Before(now) {
			ann.Status = common.StatusPublished
		}
	}

	if req.IsPinned {
		if err := tryPinAnnouncement(ann); err != nil {
			return nil, err
		}
	}

	globalStore.CreateAnnouncement(ann)
	return ann, nil
}

func tryPinAnnouncement(ann *common.Announcement) error {
	pinned := globalStore.GetPinnedAnnouncements()
	if len(pinned) >= MaxPinnedCount {
		sort.Slice(pinned, func(i, j int) bool {
			if pinned[i].PinnedAt == nil {
				return false
			}
			if pinned[j].PinnedAt == nil {
				return true
			}
			return pinned[i].PinnedAt.Before(*pinned[j].PinnedAt)
		})
		toUnpin := pinned[0]
		toUnpin.IsPinned = false
		toUnpin.PinnedAt = nil
		globalStore.UpdateAnnouncement(toUnpin)
	}

	now := time.Now()
	ann.IsPinned = true
	ann.PinnedAt = &now
	return nil
}

func UpdateAnnouncement(userID string, annID string, req *common.UpdateAnnouncementRequest) (*common.Announcement, error) {
	if !isAdmin(userID) {
		return nil, &APIError{Code: 403, Message: "只有管理员可以编辑公告"}
	}

	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return nil, &APIError{Code: 404, Message: "公告不存在"}
	}

	if ann.Status == common.StatusArchived {
		return nil, &APIError{Code: 400, Message: "已归档的公告不能修改"}
	}

	oldContent := ann.Content
	if oldContent == req.Content {
		return ann, nil
	}

	now := time.Now()
	history := common.ChangeHistory{
		ID:            "ch_" + generateID(),
		ContentBefore: oldContent,
		ContentAfter:  req.Content,
		ModifiedBy:    userID,
		ModifiedAt:    now,
	}
	ann.ChangeHistory = append(ann.ChangeHistory, history)
	ann.Content = req.Content
	ann.UpdatedAt = now

	globalStore.UpdateAnnouncement(ann)
	return ann, nil
}

func DeleteDraftAnnouncement(userID string, annID string) error {
	if !isAdmin(userID) {
		return &APIError{Code: 403, Message: "只有管理员可以删除公告"}
	}

	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return &APIError{Code: 404, Message: "公告不存在"}
	}

	if ann.Status != common.StatusDraft {
		return &APIError{Code: 400, Message: "只能删除草稿状态的公告"}
	}

	globalStore.DeleteAnnouncement(annID)
	return nil
}

func GetAnnouncementDetail(userID string, annID string) (*common.AnnouncementDetailResponse, error) {
	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return nil, &APIError{Code: 404, Message: "公告不存在"}
	}

	user, ok := globalStore.GetUser(userID)
	if !ok {
		return nil, &APIError{Code: 401, Message: "用户不存在"}
	}

	if !isAdmin(userID) {
		if ann.Status != common.StatusPublished {
			return nil, &APIError{Code: 403, Message: "无权限访问此公告"}
		}
		if !isEffectiveNow(ann) {
			return nil, &APIError{Code: 403, Message: "公告未在生效期内"}
		}
		if !checkScope(ann, user) {
			return nil, &APIError{Code: 403, Message: "无权限访问此公告"}
		}
	}

	totalAudience := calculateAudience(ann)
	readRate := 0.0
	if totalAudience > 0 {
		readRate = float64(len(ann.ViewedUserIDs)) / float64(totalAudience) * 100
	}

	response := &common.AnnouncementDetailResponse{
		Announcement: *ann,
		ViewStats: common.ViewStatistics{
			ViewCount:     ann.ViewCount,
			TotalAudience: totalAudience,
			ReadRate:      readRate,
			ReadUserIDs:   ann.ViewedUserIDs,
		},
	}

	if !isAdmin(userID) {
		alreadyViewed := false
		for _, uid := range ann.ViewedUserIDs {
			if uid == userID {
				alreadyViewed = true
				break
			}
		}
		if !alreadyViewed {
			ann.ViewCount++
			ann.ViewedUserIDs = append(ann.ViewedUserIDs, userID)
			globalStore.UpdateAnnouncement(ann)
			response.Announcement.ViewCount = ann.ViewCount
			response.Announcement.ViewedUserIDs = ann.ViewedUserIDs
			response.ViewStats.ViewCount = ann.ViewCount
			response.ViewStats.ReadUserIDs = ann.ViewedUserIDs
			if totalAudience > 0 {
				response.ViewStats.ReadRate = float64(len(ann.ViewedUserIDs)) / float64(totalAudience) * 100
			}
		}
	}

	return response, nil
}

func calculateAudience(ann *common.Announcement) int {
	if ann.Scope == common.ScopeAll {
		return len(globalStore.GetAllUsers())
	}
	count := 0
	for _, deptID := range ann.TargetDeptIDs {
		count += len(globalStore.GetUsersByDept(deptID))
	}
	return count
}

func ListEmployeeAnnouncements(userID string, keyword string) (*common.AnnouncementListResponse, error) {
	user, ok := globalStore.GetUser(userID)
	if !ok {
		return nil, &APIError{Code: 401, Message: "用户不存在"}
	}

	ProcessScheduledTasks()

	var announcements []*common.Announcement
	for _, ann := range globalStore.GetAllAnnouncements() {
		if ann.Status != common.StatusPublished {
			continue
		}
		if !isEffectiveNow(ann) {
			continue
		}
		if !checkScope(ann, user) {
			continue
		}
		if keyword != "" {
			if !strings.Contains(ann.Title, keyword) && !strings.Contains(ann.Content, keyword) {
				continue
			}
		}
		announcements = append(announcements, ann)
	}

	sort.Slice(announcements, func(i, j int) bool {
		if announcements[i].IsPinned != announcements[j].IsPinned {
			return announcements[i].IsPinned
		}
		if announcements[i].IsPinned && announcements[j].IsPinned {
			if announcements[i].PinnedAt == nil {
				return false
			}
			if announcements[j].PinnedAt == nil {
				return true
			}
			return announcements[i].PinnedAt.After(*announcements[j].PinnedAt)
		}
		return announcements[i].CreatedAt.After(announcements[j].CreatedAt)
	})

	result := make([]common.Announcement, 0, len(announcements))
	for _, a := range announcements {
		result = append(result, *a)
	}

	return &common.AnnouncementListResponse{
		Announcements: result,
		Total:         len(result),
	}, nil
}

func ListAdminAnnouncements(userID string, keyword string) (*common.AnnouncementListResponse, error) {
	if !isAdmin(userID) {
		return nil, &APIError{Code: 403, Message: "无权限"}
	}

	ProcessScheduledTasks()

	var announcements []*common.Announcement
	for _, ann := range globalStore.GetAllAnnouncements() {
		if keyword != "" {
			if !strings.Contains(ann.Title, keyword) && !strings.Contains(ann.Content, keyword) {
				continue
			}
		}
		announcements = append(announcements, ann)
	}

	sort.Slice(announcements, func(i, j int) bool {
		return announcements[i].CreatedAt.After(announcements[j].CreatedAt)
	})

	result := make([]common.Announcement, 0, len(announcements))
	for _, a := range announcements {
		result = append(result, *a)
	}

	return &common.AnnouncementListResponse{
		Announcements: result,
		Total:         len(result),
	}, nil
}

func SubmitApproval(userID string, annID string) error {
	if !isAdmin(userID) {
		return &APIError{Code: 403, Message: "只有管理员可以提交审批"}
	}

	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return &APIError{Code: 404, Message: "公告不存在"}
	}

	if ann.Status != common.StatusDraft {
		return &APIError{Code: 400, Message: "只有草稿状态的公告可以提交审批"}
	}

	if ann.ApprovalRecord != nil && ann.ApprovalRecord.Status == common.ApprovalPending {
		return &APIError{Code: 400, Message: "该公告已在审批中"}
	}

	now := time.Now()
	ar := &common.ApprovalRecord{
		ID:             "apr_" + generateID(),
		AnnouncementID: annID,
		RequesterID:    userID,
		Status:         common.ApprovalPending,
		RequestAt:      now,
	}

	ann.Status = common.StatusPending
	ann.ApprovalRecord = ar
	ann.UpdatedAt = now

	globalStore.CreateApprovalRecord(ar)
	globalStore.UpdateAnnouncement(ann)
	return nil
}

func ApproveAnnouncement(userID string, annID string, comment string) error {
	if !isManager(userID) {
		return &APIError{Code: 403, Message: "只有部门经理可以审批"}
	}

	user, ok := globalStore.GetUser(userID)
	if !ok {
		return &APIError{Code: 401, Message: "用户不存在"}
	}

	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return &APIError{Code: 404, Message: "公告不存在"}
	}

	if ann.Status != common.StatusPending {
		return &APIError{Code: 400, Message: "该公告不在待审批状态"}
	}

	if ann.ApprovalRecord == nil || ann.ApprovalRecord.Status != common.ApprovalPending {
		return &APIError{Code: 400, Message: "该公告没有待处理的审批"}
	}

	isApprover := false
	if ann.Scope == common.ScopeAll {
		return &APIError{Code: 400, Message: "全员公告不需要审批"}
	}
	for _, deptID := range ann.TargetDeptIDs {
		if deptID == user.DeptID {
			isApprover = true
			break
		}
	}
	if !isApprover {
		return &APIError{Code: 403, Message: "您不是该公告的审批人"}
	}

	now := time.Now()
	ann.ApprovalRecord.ApproverID = userID
	ann.ApprovalRecord.Status = common.ApprovalApproved
	ann.ApprovalRecord.ApprovedAt = &now
	ann.ApprovalRecord.Comment = comment

	ann.Status = common.StatusPublished
	ann.UpdatedAt = now

	globalStore.UpdateApprovalRecord(ann.ApprovalRecord)
	globalStore.UpdateAnnouncement(ann)
	return nil
}

func RejectAnnouncement(userID string, annID string, comment string) error {
	if !isManager(userID) {
		return &APIError{Code: 403, Message: "只有部门经理可以审批"}
	}

	user, ok := globalStore.GetUser(userID)
	if !ok {
		return &APIError{Code: 401, Message: "用户不存在"}
	}

	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return &APIError{Code: 404, Message: "公告不存在"}
	}

	if ann.Status != common.StatusPending {
		return &APIError{Code: 400, Message: "该公告不在待审批状态"}
	}

	if ann.ApprovalRecord == nil || ann.ApprovalRecord.Status != common.ApprovalPending {
		return &APIError{Code: 400, Message: "该公告没有待处理的审批"}
	}

	isApprover := false
	if ann.Scope == common.ScopeAll {
		return &APIError{Code: 400, Message: "全员公告不需要审批"}
	}
	for _, deptID := range ann.TargetDeptIDs {
		if deptID == user.DeptID {
			isApprover = true
			break
		}
	}
	if !isApprover {
		return &APIError{Code: 403, Message: "您不是该公告的审批人"}
	}

	now := time.Now()
	ann.ApprovalRecord.ApproverID = userID
	ann.ApprovalRecord.Status = common.ApprovalRejected
	ann.ApprovalRecord.ApprovedAt = &now
	ann.ApprovalRecord.Comment = comment

	ann.Status = common.StatusDraft
	ann.UpdatedAt = now

	globalStore.UpdateApprovalRecord(ann.ApprovalRecord)
	globalStore.UpdateAnnouncement(ann)
	return nil
}

func ListPendingApprovals(userID string) (*common.ApprovalListResponse, error) {
	if !isManager(userID) {
		return nil, &APIError{Code: 403, Message: "只有部门经理可以查看待审批列表"}
	}

	user, ok := globalStore.GetUser(userID)
	if !ok {
		return nil, &APIError{Code: 401, Message: "用户不存在"}
	}

	pending := globalStore.GetPendingApprovalsByApprover(user.DeptID)
	result := make([]common.ApprovalRecord, 0, len(pending))
	for _, a := range pending {
		result = append(result, *a)
	}

	return &common.ApprovalListResponse{
		Approvals: result,
		Total:     len(result),
	}, nil
}

func TogglePin(userID string, annID string) error {
	if !isAdmin(userID) {
		return &APIError{Code: 403, Message: "只有管理员可以操作置顶"}
	}

	ann, ok := globalStore.GetAnnouncement(annID)
	if !ok {
		return &APIError{Code: 404, Message: "公告不存在"}
	}

	if ann.Status != common.StatusPublished {
		return &APIError{Code: 400, Message: "只有已发布的公告可以置顶"}
	}

	if ann.IsPinned {
		ann.IsPinned = false
		ann.PinnedAt = nil
	} else {
		if err := tryPinAnnouncement(ann); err != nil {
			return err
		}
	}

	ann.UpdatedAt = time.Now()
	globalStore.UpdateAnnouncement(ann)
	return nil
}

func GetUserInfo(userID string) (*common.User, error) {
	user, ok := globalStore.GetUser(userID)
	if !ok {
		return nil, &APIError{Code: 404, Message: "用户不存在"}
	}
	return user, nil
}

func GetAllDepartments() []*common.Department {
	return globalStore.GetAllDepartments()
}
