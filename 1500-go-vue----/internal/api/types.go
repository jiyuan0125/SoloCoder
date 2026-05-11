package api

import "time"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateOwnerRequest struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	RoomNumber string `json:"room_number"`
	Area       int64  `json:"area"`
	Status     string `json:"status"`
}

type CreateDelegateRequest struct {
	PrincipalID string `json:"principal_id"`
	AgentID     string `json:"agent_id"`
}

type AcceptDelegateRequest struct {
	DelegateID string `json:"delegate_id"`
	AgentID    string `json:"agent_id"`
}

type CreateElectionRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	Rules         string `json:"rules"`
	CommitteeSize int    `json:"committee_size"`
}

type AddCandidateRequest struct {
	ElectionID string `json:"election_id"`
	OwnerID    string `json:"owner_id"`
	Reason     string `json:"reason"`
}

type CastVoteRequest struct {
	ElectionID  string `json:"election_id"`
	VoterID     string `json:"voter_id"`
	CandidateID string `json:"candidate_id,omitempty"`
	Abstain     bool   `json:"abstain"`
}

type CreateProposalRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	ProposerID string `json:"proposer_id"`
}

type SecondProposalRequest struct {
	ProposalID string `json:"proposal_id"`
	SeconderID string `json:"seconder_id"`
}

type ReviewProposalRequest struct {
	ProposalID       string `json:"proposal_id"`
	Decision         string `json:"decision"`
	Opinion          string `json:"opinion"`
	ResponsibleDept  string `json:"responsible_dept,omitempty"`
	DeadlineDays     int    `json:"deadline_days,omitempty"`
}

type CreateAnnouncementRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ValidityDays int   `json:"validity_days"`
}

type OwnerDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	RoomNumber string `json:"room_number"`
	Area       int64  `json:"area"`
	Status     string `json:"status"`
	VoteWeight int64  `json:"vote_weight"`
}

type DelegateRelationDTO struct {
	ID          string     `json:"id"`
	PrincipalID string     `json:"principal_id"`
	AgentID     string     `json:"agent_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
}

type CandidateDTO struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	VoteCount int64  `json:"vote_count"`
}

type ElectionDTO struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Rules         string        `json:"rules"`
	CommitteeSize int           `json:"committee_size"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Status        string        `json:"status"`
	Candidates    []CandidateDTO `json:"candidates"`
	Winners       []string      `json:"winners"`
}

type ProposalDTO struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	ProposerID      string     `json:"proposer_id"`
	Seconders       []string   `json:"seconders"`
	Status          string     `json:"status"`
	CommitteeOpinion string    `json:"committee_opinion,omitempty"`
	ResponsibleDept string     `json:"responsible_dept,omitempty"`
	Deadline        *time.Time `json:"deadline,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
}

type AnnouncementDTO struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	PublishedAt time.Time `json:"published_at"`
	ExpireAt    time.Time `json:"expire_at"`
	Status      string    `json:"status"`
}
