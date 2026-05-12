package models

import (
	"time"
)

type UserRole string

const (
	RoleAuthor  UserRole = "author"
	RoleEditor  UserRole = "editor"
	RoleReviewer UserRole = "reviewer"
	RoleAdmin   UserRole = "admin"
)

type PaperStatus string

const (
	StatusDraft        PaperStatus = "draft"
	StatusPendingReview PaperStatus = "pending_review"
	StatusInReview     PaperStatus = "in_review"
	StatusRevision     PaperStatus = "revision"
	StatusAccepted     PaperStatus = "accepted"
	StatusPublished    PaperStatus = "published"
	StatusRejected     PaperStatus = "rejected"
)

type ReviewDecision string

const (
	DecisionAccept    ReviewDecision = "accept"
	DecisionRevision  ReviewDecision = "revision"
	DecisionReject    ReviewDecision = "reject"
)

type Author struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Affiliation  string `json:"affiliation"`
	IsFirst      bool   `json:"is_first"`
	IsCorresponding bool `json:"is_corresponding"`
}

type Reviewer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Affiliation string    `json:"affiliation"`
	AssignedAt  time.Time `json:"assigned_at"`
	Submitted   bool      `json:"submitted"`
}

type Review struct {
	ID         string         `json:"id"`
	ReviewerID string         `json:"reviewer_id"`
	ReviewerName string       `json:"reviewer_name"`
	Decision   ReviewDecision `json:"decision"`
	Score      int            `json:"score"`
	Comments   string         `json:"comments"`
	SubmittedAt time.Time     `json:"submitted_at"`
}

type PublicationInfo struct {
	JournalName string `json:"journal_name"`
	Volume      string `json:"volume"`
	Issue       string `json:"issue"`
	PageStart   string `json:"page_start"`
	PageEnd     string `json:"page_end"`
	DOI         string `json:"doi"`
}

type VersionHistory struct {
	ID        string      `json:"id"`
	PaperID   string      `json:"paper_id"`
	Status    PaperStatus `json:"status"`
	OperatorID string     `json:"operator_id"`
	OperatorName string   `json:"operator_name"`
	ChangedAt time.Time   `json:"changed_at"`
	Description string    `json:"description"`
}

type Paper struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Abstract        string           `json:"abstract"`
	Keywords        []string         `json:"keywords"`
	Authors         []Author         `json:"authors"`
	SubjectCategory string           `json:"subject_category"`
	TargetJournal   string           `json:"target_journal"`
	Status          PaperStatus      `json:"status"`
	AssignedReviewers []Reviewer     `json:"assigned_reviewers"`
	Reviews         []Review         `json:"reviews"`
	PublicationInfo *PublicationInfo `json:"publication_info,omitempty"`
	VersionHistory  []VersionHistory `json:"version_history"`
	CreatedBy       string           `json:"created_by"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type User struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Role UserRole `json:"role"`
}

type CreatePaperRequest struct {
	Title           string   `json:"title" binding:"required"`
	Abstract        string   `json:"abstract"`
	Keywords        string   `json:"keywords"`
	Authors         []Author `json:"authors" binding:"required"`
	SubjectCategory string   `json:"subject_category"`
	TargetJournal   string   `json:"target_journal"`
}

type UpdatePaperRequest struct {
	Title           string   `json:"title"`
	Abstract        string   `json:"abstract"`
	Keywords        string   `json:"keywords"`
	Authors         []Author `json:"authors"`
	SubjectCategory string   `json:"subject_category"`
	TargetJournal   string   `json:"target_journal"`
}

type AssignReviewersRequest struct {
	Reviewers []Reviewer `json:"reviewers" binding:"required"`
}

type SubmitReviewRequest struct {
	ReviewerID string         `json:"reviewer_id" binding:"required"`
	Decision   ReviewDecision `json:"decision" binding:"required"`
	Score      int            `json:"score" binding:"required"`
	Comments   string         `json:"comments"`
}

type PublishPaperRequest struct {
	PublicationInfo PublicationInfo `json:"publication_info" binding:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
