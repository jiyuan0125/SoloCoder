package models

import "time"

type Stage string

const (
	StageResumeScreening Stage = "resume_screening"
	StageWrittenTest     Stage = "written_test"
	StageTechInterview   Stage = "tech_interview"
	StageHRInterview     Stage = "hr_interview"
	StageOffer           Stage = "offer"
	StageOnboarding      Stage = "onboarding"
	StageRejected        Stage = "rejected"
	StageOfferExpired    Stage = "offer_expired"
)

var StageOrder = []Stage{
	StageResumeScreening,
	StageWrittenTest,
	StageTechInterview,
	StageHRInterview,
	StageOffer,
	StageOnboarding,
}

type Candidate struct {
	ID              int64
	Name            string
	Email           string
	Phone           string
	PositionID      int64
	CurrentStage    Stage
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastRejectedAt  *time.Time
}

type StageRecord struct {
	ID           int64
	CandidateID  int64
	Stage        Stage
	Owner        string
	DueDate      time.Time
	Score        *float64
	Status       string
	Remark       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TechInterviewerScore struct {
	ID           int64
	CandidateID  int64
	Interviewer  string
	Score        float64
	Submitted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Offer struct {
	ID           int64
	CandidateID  int64
	PositionID   int64
	ValidUntil   time.Time
	Accepted     bool
	Cancelled    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Position struct {
	ID             int64
	Name           string
	TotalQuota     int
	UsedQuota      int
	PassScore      float64
	TechThreshold  float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type HistoryRecord struct {
	ID           int64
	CandidateID  int64
	Stage        Stage
	Action       string
	Operator     string
	Remark       string
	Score        *float64
	CreatedAt    time.Time
}
