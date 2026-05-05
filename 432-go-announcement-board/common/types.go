package common

import "time"

type UserRole string

const (
	RoleEmployee UserRole = "employee"
	RoleManager  UserRole = "manager"
	RoleAdmin    UserRole = "admin"
)

type AnnouncementStatus string

const (
	StatusDraft     AnnouncementStatus = "draft"
	StatusPending   AnnouncementStatus = "pending"
	StatusPublished AnnouncementStatus = "published"
	StatusArchived  AnnouncementStatus = "archived"
)

type AnnouncementScope string

const (
	ScopeAll      AnnouncementScope = "all"
	ScopeDept     AnnouncementScope = "dept"
)

type AnnouncementPriority string

const (
	PriorityNormal AnnouncementPriority = "normal"
	PriorityUrgent AnnouncementPriority = "urgent"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type User struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	DeptID   string   `json:"dept_id"`
	DeptName string   `json:"dept_name"`
	Role     UserRole `json:"role"`
}

type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Attachment struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	UploadAt  time.Time `json:"upload_at"`
}

type ChangeHistory struct {
	ID           string    `json:"id"`
	ContentBefore string   `json:"content_before"`
	ContentAfter  string   `json:"content_after"`
	ModifiedBy   string    `json:"modified_by"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type ApprovalRecord struct {
	ID            string         `json:"id"`
	AnnouncementID string        `json:"announcement_id"`
	RequesterID   string         `json:"requester_id"`
	ApproverID    string         `json:"approver_id,omitempty"`
	Status        ApprovalStatus `json:"status"`
	RequestAt     time.Time      `json:"request_at"`
	ApprovedAt    *time.Time     `json:"approved_at,omitempty"`
	Comment       string         `json:"comment,omitempty"`
}

type Announcement struct {
	ID             string               `json:"id"`
	Title          string               `json:"title"`
	Content        string               `json:"content"`
	Scope          AnnouncementScope    `json:"scope"`
	TargetDeptIDs  []string             `json:"target_dept_ids,omitempty"`
	IsPinned       bool                 `json:"is_pinned"`
	PinnedAt       *time.Time           `json:"pinned_at,omitempty"`
	EffectiveStart time.Time            `json:"effective_start"`
	EffectiveEnd   time.Time            `json:"effective_end"`
	Priority       AnnouncementPriority `json:"priority"`
	Status         AnnouncementStatus   `json:"status"`
	CreatedBy      string               `json:"created_by"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Attachments    []Attachment         `json:"attachments"`
	ChangeHistory  []ChangeHistory      `json:"change_history"`
	ViewCount      int                  `json:"view_count"`
	ViewedUserIDs  []string             `json:"viewed_user_ids"`
	ApprovalRecord *ApprovalRecord      `json:"approval_record,omitempty"`
}
