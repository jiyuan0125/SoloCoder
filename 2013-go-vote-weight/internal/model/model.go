package model

import "time"

type Owner struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Area   int    `json:"area"`
}

type VotingStatus string

const (
	VotingStatusPending   VotingStatus = "pending"
	VotingStatusActive    VotingStatus = "active"
	VotingStatusPassed    VotingStatus = "passed"
	VotingStatusRejected  VotingStatus = "rejected"
)

type Voting struct {
	ID              int64        `json:"id"`
	Topic           string       `json:"topic"`
	Deadline        time.Time    `json:"deadline"`
	Status          VotingStatus `json:"status"`
	Participation   float64      `json:"participation"`
	YesVotes        int64        `json:"yes_votes"`
	TotalVotes      int64        `json:"total_votes"`
	CreatedAt       time.Time    `json:"created_at"`
}

type VoteChoice string

const (
	VoteChoiceYes VoteChoice = "yes"
	VoteChoiceNo  VoteChoice = "no"
)

type Vote struct {
	ID         int64      `json:"id"`
	VotingID   int64      `json:"voting_id"`
	OwnerID    int64      `json:"owner_id"`
	ProxyForID *int64     `json:"proxy_for_id"`
	Choice     VoteChoice `json:"choice"`
	VoteWeight int64      `json:"vote_weight"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Proxy struct {
	ID             int64     `json:"id"`
	VotingID       int64     `json:"voting_id"`
	DelegatorID    int64     `json:"delegator_id"`
	TrusteeID      int64     `json:"trustee_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type VotingActionType string

const (
	ActionTypeCreate VotingActionType = "create"
	ActionTypeVote   VotingActionType = "vote"
	ActionTypeProxy  VotingActionType = "proxy"
	ActionTypeClose  VotingActionType = "close"
)

type VotingAuditLog struct {
	ID           int64            `json:"id"`
	VotingID     int64            `json:"voting_id"`
	ActionType   VotingActionType `json:"action_type"`
	OwnerID      *int64           `json:"owner_id"`
	Details      string           `json:"details"`
	CreatedAt    time.Time        `json:"created_at"`
}
