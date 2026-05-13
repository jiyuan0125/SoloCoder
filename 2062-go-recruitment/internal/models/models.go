package models

import (
	"time"
)

type JobStatus string

const (
	JobStatusOpen   JobStatus = "open"
	JobStatusClosed JobStatus = "closed"
)

type CandidateStatus string

const (
	CandidateStatusApplied    CandidateStatus = "applied"
	CandidateStatusScreening  CandidateStatus = "screening"
	CandidateStatusInterview  CandidateStatus = "interview"
	CandidateStatusOffer      CandidateStatus = "offer"
	CandidateStatusOnboard    CandidateStatus = "onboard"
	CandidateStatusRejected   CandidateStatus = "rejected"
	CandidateStatusWithdrawn  CandidateStatus = "withdrawn"
	CandidateStatusOfferAbandoned CandidateStatus = "offer_abandoned"
)

type InterviewResult string

const (
	InterviewResultPass     InterviewResult = "pass"
	InterviewResultFail     InterviewResult = "fail"
	InterviewResultPending  InterviewResult = "pending"
)

type InterviewStatus string

const (
	InterviewStatusScheduled InterviewStatus = "scheduled"
	InterviewStatusCompleted InterviewStatus = "completed"
)

type OfferStatus string

const (
	OfferStatusPending   OfferStatus = "pending"
	OfferStatusAccepted  OfferStatus = "accepted"
	OfferStatusRejected  OfferStatus = "rejected"
	OfferStatusAbandoned OfferStatus = "abandoned"
	OfferStatusWithdrawn OfferStatus = "withdrawn"
)

type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Department  string    `json:"department"`
	Requirements string   `json:"requirements"`
	SalaryMin   float64   `json:"salary_min"`
	SalaryMax   float64   `json:"salary_max"`
	Status      JobStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

type Candidate struct {
	ID             string            `json:"id"`
	ResumeID       string            `json:"resume_id"`
	Name           string            `json:"name"`
	Email          string            `json:"email"`
	Phone          string            `json:"phone"`
	JobID          string            `json:"job_id"`
	Status         CandidateStatus   `json:"status"`
	AppliedAt      time.Time         `json:"applied_at"`
	RejectionReason string           `json:"rejection_reason,omitempty"`
	RejectedAt     *time.Time        `json:"rejected_at,omitempty"`
}

type Interview struct {
	ID             string          `json:"id"`
	CandidateID    string          `json:"candidate_id"`
	JobID          string          `json:"job_id"`
	Interviewer    string          `json:"interviewer"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	Status         InterviewStatus `json:"status"`
	Result         *InterviewResult `json:"result,omitempty"`
	Feedback       string          `json:"feedback,omitempty"`
	IsRetry        bool            `json:"is_retry"`
	Round          int             `json:"round"`
	SubmittedAt    *time.Time      `json:"submitted_at,omitempty"`
}

type Offer struct {
	ID           string      `json:"id"`
	CandidateID  string      `json:"candidate_id"`
	JobID        string      `json:"job_id"`
	Salary       float64     `json:"salary"`
	StartDate    time.Time   `json:"start_date"`
	ValidUntil   time.Time   `json:"valid_until"`
	Status       OfferStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	RepliedAt    *time.Time  `json:"replied_at,omitempty"`
}

type CalendarEvent struct {
	Interviewer string    `json:"interviewer"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
}

type BudgetStage struct {
	Stage       string  `json:"stage"`
	PlannedAmount float64 `json:"planned_amount"`
}

type Budget struct {
	JobID         string        `json:"job_id"`
	TotalAmount   float64       `json:"total_amount"`
	Stages        []BudgetStage `json:"stages"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type Database struct {
	Jobs       []Job       `json:"jobs"`
	Candidates []Candidate `json:"candidates"`
	Interviews []Interview `json:"interviews"`
	Offers     []Offer     `json:"offers"`
	Budgets    []Budget    `json:"budgets"`
}

func NewDatabase() *Database {
	return &Database{
		Jobs:       []Job{},
		Candidates: []Candidate{},
		Interviews: []Interview{},
		Offers:     []Offer{},
		Budgets:    []Budget{},
	}
}
