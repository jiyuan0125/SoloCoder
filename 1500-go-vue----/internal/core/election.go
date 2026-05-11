package core

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ElectionService struct {
	store        *Store
	ownerService *OwnerService
	voteMu       sync.Mutex
}

func NewElectionService(store *Store, ownerService *OwnerService) *ElectionService {
	return &ElectionService{
		store:        store,
		ownerService: ownerService,
	}
}

func (s *ElectionService) CreateElection(title, description, rules string, committeeSize int) (*Election, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if committeeSize <= 0 {
		return nil, errors.New("committee size must be positive")
	}

	now := time.Now()
	election := &Election{
		ID:            uuid.NewString(),
		Title:         title,
		Description:   description,
		Rules:         rules,
		CommitteeSize: committeeSize,
		StartTime:     now,
		EndTime:       now.Add(7 * 24 * time.Hour),
		Status:        ElectionStatusActive,
		Candidates:    []Candidate{},
		Winners:       []string{},
	}
	s.store.AddElection(election)
	return election, nil
}

func (s *ElectionService) GetElection(id string) (*Election, error) {
	election := s.store.GetElection(id)
	if election == nil {
		return nil, errors.New("election not found")
	}
	return election, nil
}

func (s *ElectionService) ListElections() []*Election {
	return s.store.GetElections()
}

func (s *ElectionService) AddCandidate(electionID, ownerID, reason string) (*Candidate, error) {
	election, err := s.GetElection(electionID)
	if err != nil {
		return nil, err
	}
	if election.Status != ElectionStatusActive {
		return nil, errors.New("election is not active")
	}

	owner, err := s.ownerService.GetOwner(ownerID)
	if err != nil {
		return nil, err
	}

	for _, c := range election.Candidates {
		if c.OwnerID == ownerID {
			return nil, errors.New("owner is already a candidate")
		}
	}

	candidate := Candidate{
		ID:         uuid.NewString(),
		ElectionID: electionID,
		OwnerID:    ownerID,
		Name:       owner.Name,
		Reason:     reason,
		VoteCount:  0,
	}

	election.Candidates = append(election.Candidates, candidate)
	s.store.UpdateElection(election)
	return &candidate, nil
}

func (s *ElectionService) RecommendCandidate(electionID, candidateID, recommenderID string) error {
	if _, err := s.GetElection(electionID); err != nil {
		return err
	}
	if _, err := s.ownerService.GetOwner(recommenderID); err != nil {
		return err
	}

	return nil
}

func (s *ElectionService) CastVote(electionID, voterID, candidateID string, abstain bool) error {
	election, err := s.GetElection(electionID)
	if err != nil {
		return err
	}

	now := time.Now()
	if now.Before(election.StartTime) {
		return errors.New("voting has not started")
	}
	if now.After(election.EndTime) {
		return errors.New("voting has ended")
	}
	if election.Status != ElectionStatusActive {
		return errors.New("election is not active")
	}

	if _, err := s.ownerService.GetOwner(voterID); err != nil {
		return err
	}

	s.voteMu.Lock()
	defer s.voteMu.Unlock()

	if s.store.HasVoted(electionID, voterID) {
		return errors.New("already voted")
	}

	voteWeight, err := s.ownerService.GetVoteWeight(voterID)
	if err != nil {
		return err
	}

	choice := VoteChoiceCandidate
	if abstain {
		choice = VoteChoiceAbstain
		candidateID = ""
	} else {
		found := false
		for _, c := range election.Candidates {
			if c.ID == candidateID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("candidate not found in this election")
		}
	}

	vote := &Vote{
		ID:          uuid.NewString(),
		ElectionID:  electionID,
		VoterID:     voterID,
		Choice:      choice,
		CandidateID: candidateID,
		VoteWeight:  voteWeight,
		SubmittedAt: now,
	}

	s.store.AddVote(electionID, vote)

	if !abstain {
		for i := range election.Candidates {
			if election.Candidates[i].ID == candidateID {
				election.Candidates[i].VoteCount += voteWeight
				break
			}
		}
		s.store.UpdateElection(election)
	}

	return nil
}

func (s *ElectionService) FinalizeElection(electionID string) (*Election, error) {
	election, err := s.GetElection(electionID)
	if err != nil {
		return nil, err
	}
	if election.Status != ElectionStatusActive {
		return nil, errors.New("election is not active")
	}

	now := time.Now()
	if now.Before(election.EndTime) {
		return nil, errors.New("voting has not ended yet")
	}

	s.voteMu.Lock()
	defer s.voteMu.Unlock()

	votes := s.store.GetVotes(electionID)
	candidateVotes := make(map[string]int64)
	for _, c := range election.Candidates {
		candidateVotes[c.ID] = c.VoteCount
	}

	for _, vote := range votes {
		if vote.Choice == VoteChoiceCandidate {
			candidateVotes[vote.CandidateID] += vote.VoteWeight
		}
	}

	for i := range election.Candidates {
		election.Candidates[i].VoteCount = candidateVotes[election.Candidates[i].ID]
	}

	sortedCandidates := make([]Candidate, len(election.Candidates))
	copy(sortedCandidates, election.Candidates)

	sort.SliceStable(sortedCandidates, func(i, j int) bool {
		if sortedCandidates[i].VoteCount == sortedCandidates[j].VoteCount {
			return rand.Intn(2) == 0
		}
		return sortedCandidates[i].VoteCount > sortedCandidates[j].VoteCount
	})

	winners := []string{}
	for i := 0; i < election.CommitteeSize && i < len(sortedCandidates); i++ {
		winners = append(winners, sortedCandidates[i].ID)
	}

	election.Winners = winners
	election.Status = ElectionStatusFinished
	s.store.UpdateElection(election)

	return election, nil
}

func (s *ElectionService) GetVoteCount(electionID string) (map[string]int64, error) {
	election, err := s.GetElection(electionID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, c := range election.Candidates {
		result[c.ID] = c.VoteCount
	}
	return result, nil
}

func (s *ElectionService) CheckTimeRemaining(electionID string) (time.Duration, error) {
	election, err := s.GetElection(electionID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	if now.After(election.EndTime) {
		return 0, fmt.Errorf("voting has ended")
	}
	return election.EndTime.Sub(now), nil
}
