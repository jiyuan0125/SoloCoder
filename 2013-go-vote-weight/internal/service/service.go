package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"votingsystem/internal/db"
	"votingsystem/internal/model"
)

var (
	ErrVotingNotFound   = errors.New("voting not found")
	ErrVotingEnded      = errors.New("表决已结束")
	ErrOwnerNotFound    = errors.New("owner not found")
	ErrAlreadyVoted     = errors.New("owner already voted")
	ErrAlreadyDelegated = errors.New("已委托给他人")
	ErrDuplicateProxy   = errors.New("duplicate proxy")
	ErrCannotDelegate   = errors.New("cannot delegate to yourself")
)

type Service struct {
	store *db.Store
}

func NewService(store *db.Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListVotings() ([]model.Voting, error) {
	return s.store.ListVotings()
}

func (s *Service) GetVoting(id int64) (*model.Voting, error) {
	v, err := s.store.GetVoting(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVotingNotFound
		}
		return nil, err
	}
	return v, nil
}

func (s *Service) CreateVoting(topic string, deadline time.Time) (*model.Voting, error) {
	v, err := s.store.CreateVoting(topic, deadline)
	if err != nil {
		return nil, err
	}
	_ = s.store.CreateAuditLog(v.ID, model.ActionTypeCreate, nil, fmt.Sprintf("创建表决: %s, 截止时间: %s", topic, deadline.Format(time.RFC3339)))
	return v, nil
}

func (s *Service) ListVotes(votingID int64) ([]model.Vote, error) {
	_, err := s.GetVoting(votingID)
	if err != nil {
		return nil, err
	}
	return s.store.ListVotes(votingID)
}

func (s *Service) ListProxies(votingID int64) ([]model.Proxy, error) {
	_, err := s.GetVoting(votingID)
	if err != nil {
		return nil, err
	}
	return s.store.ListProxies(votingID)
}

func (s *Service) ListAuditLogs(votingID int64) ([]model.VotingAuditLog, error) {
	_, err := s.GetVoting(votingID)
	if err != nil {
		return nil, err
	}
	return s.store.ListAuditLogs(votingID)
}

func (s *Service) Delegate(votingID int64, delegatorID int64, trusteeID int64) (*model.Proxy, error) {
	v, err := s.GetVoting(votingID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(v.Deadline) || v.Status != model.VotingStatusActive {
		return nil, ErrVotingEnded
	}

	if delegatorID == trusteeID {
		return nil, ErrCannotDelegate
	}

	_, err = s.store.GetOwner(delegatorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOwnerNotFound
		}
		return nil, err
	}

	_, err = s.store.GetOwner(trusteeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOwnerNotFound
		}
		return nil, err
	}

	hasVoted, err := s.store.CheckVote(nil, votingID, delegatorID)
	if err != nil {
		return nil, err
	}
	if hasVoted {
		return nil, ErrAlreadyVoted
	}

	hasDelegated, err := s.store.CheckDelegate(votingID, delegatorID)
	if err != nil {
		return nil, err
	}
	if hasDelegated {
		return nil, ErrDuplicateProxy
	}

	var p *model.Proxy
	err = s.store.WithTx(func(tx *sql.Tx) error {
		var txErr error
		p, txErr = s.store.CreateProxy(tx, votingID, delegatorID, trusteeID)
		if txErr != nil {
			if db.IsUniqueConstraintErr(txErr) {
				return ErrDuplicateProxy
			}
			return txErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.store.CreateAuditLog(votingID, model.ActionTypeProxy, &delegatorID, fmt.Sprintf("业主 %d 委托给 %d", delegatorID, trusteeID))
	return p, nil
}

func (s *Service) Vote(votingID int64, ownerID int64, choice model.VoteChoice) (*model.Vote, error) {
	v, err := s.GetVoting(votingID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(v.Deadline) || v.Status != model.VotingStatusActive {
		return nil, ErrVotingEnded
	}

	owner, err := s.store.GetOwner(ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOwnerNotFound
		}
		return nil, err
	}

	hasDelegated, err := s.store.CheckDelegate(votingID, ownerID)
	if err != nil {
		return nil, err
	}
	if hasDelegated {
		return nil, ErrAlreadyDelegated
	}

	hasVoted, err := s.store.CheckVote(nil, votingID, ownerID)
	if err != nil {
		return nil, err
	}
	if hasVoted {
		return nil, ErrAlreadyVoted
	}

	delegatees, err := s.store.GetDelegatees(votingID)
	if err != nil {
		return nil, err
	}

	var vote *model.Vote
	err = s.store.WithTx(func(tx *sql.Tx) error {
		totalWeight := int64(owner.Area)
		if addWeight, ok := delegatees[ownerID]; ok {
			totalWeight += addWeight
		}

		var txErr error
		vote, txErr = s.store.CreateVote(tx, votingID, ownerID, nil, choice, totalWeight)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.store.CreateAuditLog(votingID, model.ActionTypeVote, &ownerID, fmt.Sprintf("业主 %d 投票 %s, 权重 %d", ownerID, choice, vote.VoteWeight))
	return vote, nil
}

func (s *Service) CloseVoting(votingID int64) error {
	v, err := s.GetVoting(votingID)
	if err != nil {
		return err
	}

	if v.Status != model.VotingStatusActive {
		return nil
	}

	totalPower, err := s.store.GetTotalVotingPower()
	if err != nil {
		return err
	}

	yesVotes, totalVotes, err := s.store.CalculateVoteStats(votingID)
	if err != nil {
		return err
	}

	var participation float64
	if totalPower > 0 {
		participation = float64(totalVotes) / float64(totalPower)
	}

	var status model.VotingStatus
	if participation > 0.5 && totalVotes > 0 && float64(yesVotes) > float64(totalVotes)*(2.0/3.0) {
		status = model.VotingStatusPassed
	} else {
		status = model.VotingStatusRejected
	}

	if err := s.store.UpdateVotingResult(votingID, status, participation, yesVotes, totalVotes); err != nil {
		return err
	}

	_ = s.store.CreateAuditLog(votingID, model.ActionTypeClose, nil,
		fmt.Sprintf("表决结束, 状态: %s, 参与率: %.2f, 赞成: %d, 总票数: %d", status, participation, yesVotes, totalVotes))
	return nil
}

func (s *Service) ProcessExpiredVotings() error {
	expired, err := s.store.ListExpiredActiveVotings()
	if err != nil {
		return err
	}
	for _, v := range expired {
		if err := s.CloseVoting(v.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListOwners() ([]model.Owner, error) {
	return s.store.ListOwners()
}

func (s *Service) CreateOwner(name string, area int) (*model.Owner, error) {
	if area < 0 {
		area = 0
	}
	return s.store.CreateOwner(name, area)
}
