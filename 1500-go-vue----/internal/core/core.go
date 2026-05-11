package core

import (
	"time"
)

type OwnerStatus string

const (
	OwnerStatusResident OwnerStatus = "resident"
	OwnerStatusVacant   OwnerStatus = "vacant"
	OwnerStatusRented   OwnerStatus = "rented"
)

type Owner struct {
	ID           string
	Name         string
	Phone        string
	RoomNumber   string
	Area         int64
	Status       OwnerStatus
	VoteWeight   int64
}

type DelegateStatus string

const (
	DelegateStatusPending  DelegateStatus = "pending"
	DelegateStatusAccepted DelegateStatus = "accepted"
	DelegateStatusRejected DelegateStatus = "rejected"
	DelegateStatusSuperseded DelegateStatus = "superseded"
)

type DelegateRelation struct {
	ID            string
	PrincipalID   string
	AgentID       string
	Status        DelegateStatus
	CreatedAt     time.Time
	AcceptedAt    *time.Time
}

type ElectionStatus string

const (
	ElectionStatusDraft    ElectionStatus = "draft"
	ElectionStatusActive   ElectionStatus = "active"
	ElectionStatusFinished ElectionStatus = "finished"
)

type Candidate struct {
	ID           string
	ElectionID   string
	OwnerID      string
	Name         string
	Reason       string
	VoteCount    int64
}

type Election struct {
	ID            string
	Title         string
	Description   string
	Rules         string
	CommitteeSize int
	StartTime     time.Time
	EndTime       time.Time
	Status        ElectionStatus
	Candidates    []Candidate
	Winners       []string
}

type VoteChoice string

const (
	VoteChoiceAbstain VoteChoice = "abstain"
	VoteChoiceCandidate VoteChoice = "candidate"
)

type Vote struct {
	ID           string
	ElectionID   string
	VoterID      string
	Choice       VoteChoice
	CandidateID  string
	VoteWeight   int64
	SubmittedAt  time.Time
}

type ProposalStatus string

const (
	ProposalStatusPending    ProposalStatus = "pending"
	ProposalStatusReviewing  ProposalStatus = "reviewing"
	ProposalStatusAccepted   ProposalStatus = "accepted"
	ProposalStatusPartially  ProposalStatus = "partially"
	ProposalStatusPostponed  ProposalStatus = "postponed"
	ProposalStatusRejected   ProposalStatus = "rejected"
)

type Proposal struct {
	ID              string
	Title           string
	Content         string
	ProposerID      string
	Seconders       []string
	Status          ProposalStatus
	CommitteeOpinion string
	ResponsibleDept string
	Deadline        *time.Time
	CreatedAt       time.Time
	ReviewedAt      *time.Time
}

type AnnouncementStatus string

const (
	AnnouncementStatusActive AnnouncementStatus = "active"
	AnnouncementStatusArchived AnnouncementStatus = "archived"
)

type Announcement struct {
	ID         string
	Title      string
	Content    string
	PublishedAt time.Time
	ExpireAt   time.Time
	Status     AnnouncementStatus
}
