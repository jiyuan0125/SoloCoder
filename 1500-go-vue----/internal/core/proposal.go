package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const MinSecondersRequired = 10

type ProposalService struct {
	store        *Store
	ownerService *OwnerService
}

func NewProposalService(store *Store, ownerService *OwnerService) *ProposalService {
	return &ProposalService{
		store:        store,
		ownerService: ownerService,
	}
}

func (s *ProposalService) CreateProposal(title, content, proposerID string) (*Proposal, error) {
	if title == "" || content == "" {
		return nil, errors.New("title and content are required")
	}
	if _, err := s.ownerService.GetOwner(proposerID); err != nil {
		return nil, err
	}

	proposal := &Proposal{
		ID:         uuid.NewString(),
		Title:      title,
		Content:    content,
		ProposerID: proposerID,
		Seconders:  []string{},
		Status:     ProposalStatusPending,
		CreatedAt:  time.Now(),
	}
	s.store.AddProposal(proposal)
	return proposal, nil
}

func (s *ProposalService) GetProposal(id string) (*Proposal, error) {
	proposal := s.store.GetProposal(id)
	if proposal == nil {
		return nil, errors.New("proposal not found")
	}
	return proposal, nil
}

func (s *ProposalService) ListProposals() []*Proposal {
	return s.store.GetProposals()
}

func (s *ProposalService) SecondProposal(proposalID, seconderID string) (*Proposal, error) {
	proposal, err := s.GetProposal(proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != ProposalStatusPending {
		return nil, errors.New("proposal is not in pending status")
	}
	if _, err := s.ownerService.GetOwner(seconderID); err != nil {
		return nil, err
	}
	if seconderID == proposal.ProposerID {
		return nil, errors.New("proposer cannot second their own proposal")
	}

	for _, sID := range proposal.Seconders {
		if sID == seconderID {
			return nil, errors.New("already seconded this proposal")
		}
	}

	proposal.Seconders = append(proposal.Seconders, seconderID)

	if len(proposal.Seconders) >= MinSecondersRequired {
		proposal.Status = ProposalStatusReviewing
	}

	s.store.UpdateProposal(proposal)
	return proposal, nil
}

func (s *ProposalService) ReviewProposal(proposalID string, decision ProposalStatus, opinion, responsibleDept string, deadline *time.Time) (*Proposal, error) {
	proposal, err := s.GetProposal(proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != ProposalStatusReviewing {
		return nil, errors.New("proposal is not in reviewing status")
	}

	if decision != ProposalStatusAccepted &&
		decision != ProposalStatusPartially &&
		decision != ProposalStatusPostponed &&
		decision != ProposalStatusRejected {
		return nil, errors.New("invalid decision")
	}

	now := time.Now()
	proposal.Status = decision
	proposal.CommitteeOpinion = opinion
	proposal.ReviewedAt = &now

	if decision == ProposalStatusAccepted {
		proposal.ResponsibleDept = responsibleDept
		proposal.Deadline = deadline
	}

	s.store.UpdateProposal(proposal)
	return proposal, nil
}
