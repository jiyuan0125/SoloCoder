package model

import (
	"time"
)

type MeetingStatus string

const (
	MeetingStatusPreparing    MeetingStatus = "preparing"
	MeetingStatusAcceptingSubmissions MeetingStatus = "accepting_submissions"
	MeetingStatusReviewing   MeetingStatus = "reviewing"
	MeetingStatusNotifying  MeetingStatus = "notifying"
	MeetingStatusEnded     MeetingStatus = "ended"
)

type PaperStatus string

const (
	PaperStatusDraft     PaperStatus = "draft"
	PaperStatusSubmitted PaperStatus = "submitted"
	PaperStatusFormatChecking PaperStatus = "format_checking"
	PaperStatusUnderReview PaperStatus = "under_review"
	PaperStatusRevision  PaperStatus = "revision"
	PaperStatusAccepted PaperStatus = "accepted"
	PaperStatusRejected PaperStatus = "rejected"
	PaperStatusPublished PaperStatus = "published"
)

type ReviewRecommendation string

const (
	StrongAccept  ReviewRecommendation = "strong_accept"
	WeakAccept    ReviewRecommendation = "weak_accept"
	Borderline    ReviewRecommendation = "borderline"
	WeakReject    ReviewRecommendation = "weak_reject"
	StrongReject  ReviewRecommendation = "strong_reject"
)

type UserRole string

const (
	RoleAuthor      UserRole = "author"
	RoleReviewer UserRole = "reviewer"
	RoleChair    UserRole = "chair"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	UUID     string    `json:"uuid" gorm:"uniqueIndex"`
	Name     string    `json:"name"`
	Email    string    `json:"email" gorm:"uniqueIndex"`
	Password string    `json:"-"`
	Role     UserRole  `json:"role"`
	Affiliation string `json:"affiliation"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Meeting struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	UUID          string        `json:"uuid" gorm:"uniqueIndex"`
	Name          string        `json:"name"`
	Abbreviation  string        `json:"abbreviation" gorm:"uniqueIndex"`
	StartDate     time.Time     `json:"start_date"`
	EndDate       time.Time     `json:"end_date"`
	Location      string        `json:"location"`
	Topics        string        `json:"topics"`
	SubmissionDeadline time.Time `json:"submission_deadline"`
	ReviewDeadline time.Time   `json:"review_deadline"`
	NotificationDeadline time.Time `json:"notification_deadline"`
	Status        MeetingStatus `json:"status"`
	ChairID       *uint         `json:"chair_id"`
	Chair         *User       `json:"chair,omitempty" gorm:"foreignKey:ChairID"`
	ResponsibleID *uint         `json:"responsible_id"`
	Responsible  *User         `json:"responsible,omitempty" gorm:"foreignKey:ResponsibleID"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type PaperAuthor struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	PaperID       uint   `json:"paper_id"`
	Name          string `json:"name"`
	Affiliation   string `json:"affiliation"`
	Email         string `json:"email"`
	IsFirstAuthor bool  `json:"is_first_author"`
	IsCorresponding bool `json:"is_corresponding"`
	Order         int    `json:"order"`
}

type Paper struct {
	ID              uint        `json:"id" gorm:"primaryKey"`
	UUID            string      `json:"uuid" gorm:"uniqueIndex"`
	MeetingID       uint        `json:"meeting_id"`
	PaperNumber     string      `json:"paper_number" gorm:"uniqueIndex"`
	Title           string      `json:"title"`
	Abstract        string      `json:"abstract"`
	Keywords        string      `json:"keywords"`
	TopicArea       string      `json:"topic_area"`
	Content         string      `json:"content"`
	Status           PaperStatus `json:"status"`
	Authors          []PaperAuthor `json:"authors" gorm:"foreignKey:PaperID"`
	Reviews          []Review    `json:"reviews,omitempty" gorm:"foreignKey:PaperID"`
	Schedule         *Schedule `json:"schedule,omitempty" gorm:"foreignKey:PaperID"`
	SubmissionTime   time.Time  `json:"submission_time"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Review struct {
	ID               uint                 `json:"id" gorm:"primaryKey"`
	UUID             string               `json:"uuid" gorm:"uniqueIndex"`
	PaperID          uint                 `json:"paper_id"`
	ReviewerID       uint                 `json:"reviewer_id"`
	Reviewer         *User                `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
	Originality      int                  `json:"originality"`
	TechnicalQuality int                  `json:"technical_quality"`
	Relevance        int                  `json:"relevance"`
	Clarity          int                  `json:"clarity"`
	Recommendation   ReviewRecommendation `json:"recommendation"`
	Comments         string               `json:"comments"`
	ConfidentialComments string           `json:"confidential_comments"`
	SubmittedAt      *time.Time          `json:"submitted_at"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type Schedule struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UUID        string    `json:"uuid" gorm:"uniqueIndex"`
	MeetingID   uint      `json:"meeting_id"`
	PaperID     uint      `json:"paper_id"`
	Day         int       `json:"day"`
	Session    string    `json:"session"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	SessionID   *uint     `json:"session_id"`
	SessionObj  *Session  `json:"session_obj,omitempty" gorm:"foreignKey:SessionID"`
	HasConflict bool     `json:"has_conflict"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Session struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UUID      string    `json:"uuid" gorm:"uniqueIndex"`
	MeetingID uint      `json:"meeting_id"`
	Day       int       `json:"day"`
	TimeOfDay string    `json:"time_of_day"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Reminder struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UUID      string    `json:"uuid" gorm:"uniqueIndex"`
	MeetingID uint      `json:"meeting_id"`
	Type      string    `json:"type"`
	TargetUserID uint     `json:"target_user_id"`
	PaperID   *uint     `json:"paper_id"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	SentAt    time.Time `json:"sent_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UUID      string    `json:"uuid" gorm:"uniqueIndex"`
	Action    string    `json:"action"`
	EntityType string   `json:"entity_type"`
	EntityID  string    `json:"entity_id"`
	Details   string    `json:"details"`
	UserID    *uint     `json:"user_id"`
	HasError  bool      `json:"has_error"`
	ErrorMessage string `json:"error_message"`
	CreatedAt time.Time `json:"created_at"`
}
