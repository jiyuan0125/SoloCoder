package common

import "time"

type CreateAnnouncementRequest struct {
	Title          string               `json:"title"`
	Content        string               `json:"content"`
	Scope          AnnouncementScope    `json:"scope"`
	TargetDeptIDs  []string             `json:"target_dept_ids,omitempty"`
	IsPinned       bool                 `json:"is_pinned"`
	EffectiveStart time.Time            `json:"effective_start"`
	EffectiveEnd   time.Time            `json:"effective_end"`
	Priority       AnnouncementPriority `json:"priority"`
	Attachments    []AttachmentInfo     `json:"attachments,omitempty"`
}

type AttachmentInfo struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

type UpdateAnnouncementRequest struct {
	Content string `json:"content"`
}

type ListAnnouncementsRequest struct {
	Keyword string `json:"keyword,omitempty"`
}

type SubmitApprovalRequest struct {
	AnnouncementID string `json:"announcement_id"`
}

type ApproveAnnouncementRequest struct {
	AnnouncementID string `json:"announcement_id"`
	Comment        string `json:"comment,omitempty"`
}

type RejectAnnouncementRequest struct {
	AnnouncementID string `json:"announcement_id"`
	Comment        string `json:"comment"`
}

type RecordViewRequest struct {
	AnnouncementID string `json:"announcement_id"`
}

type APIRequest struct {
	UserID string      `json:"user_id"`
	Data   interface{} `json:"data,omitempty"`
}
