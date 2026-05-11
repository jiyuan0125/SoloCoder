package core

import (
	"sync"
)

type Store struct {
	mu               sync.RWMutex
	owners           map[string]*Owner
	delegates        map[string]*DelegateRelation
	elections        map[string]*Election
	votes            map[string][]*Vote
	proposals        map[string]*Proposal
	announcements    map[string]*Announcement
	ownerIndex       map[string]string
}

func NewStore() *Store {
	return &Store{
		owners:        make(map[string]*Owner),
		delegates:     make(map[string]*DelegateRelation),
		elections:     make(map[string]*Election),
		votes:         make(map[string][]*Vote),
		proposals:     make(map[string]*Proposal),
		announcements: make(map[string]*Announcement),
		ownerIndex:    make(map[string]string),
	}
}

func (s *Store) Lock()   { s.mu.Lock() }
func (s *Store) Unlock() { s.mu.Unlock() }
func (s *Store) RLock()  { s.mu.RLock() }
func (s *Store) RUnlock() { s.mu.RUnlock() }

func (s *Store) GetOwner(id string) *Owner {
	s.RLock()
	defer s.RUnlock()
	return s.owners[id]
}

func (s *Store) GetOwnerByPhone(phone string) *Owner {
	s.RLock()
	defer s.RUnlock()
	id, ok := s.ownerIndex[phone]
	if !ok {
		return nil
	}
	return s.owners[id]
}

func (s *Store) GetOwners() []*Owner {
	s.RLock()
	defer s.RUnlock()
	result := make([]*Owner, 0, len(s.owners))
	for _, o := range s.owners {
		result = append(result, o)
	}
	return result
}

func (s *Store) AddOwner(owner *Owner) {
	s.Lock()
	defer s.Unlock()
	s.owners[owner.ID] = owner
	s.ownerIndex[owner.Phone] = owner.ID
}

func (s *Store) UpdateOwner(owner *Owner) {
	s.Lock()
	defer s.Unlock()
	s.owners[owner.ID] = owner
}

func (s *Store) AddDelegate(d *DelegateRelation) {
	s.Lock()
	defer s.Unlock()
	s.delegates[d.ID] = d
}

func (s *Store) GetDelegates() []*DelegateRelation {
	s.RLock()
	defer s.RUnlock()
	result := make([]*DelegateRelation, 0, len(s.delegates))
	for _, d := range s.delegates {
		result = append(result, d)
	}
	return result
}

func (s *Store) GetDelegate(id string) *DelegateRelation {
	s.RLock()
	defer s.RUnlock()
	return s.delegates[id]
}

func (s *Store) UpdateDelegate(d *DelegateRelation) {
	s.Lock()
	defer s.Unlock()
	s.delegates[d.ID] = d
}

func (s *Store) GetActiveDelegateForAgent(agentID string) *DelegateRelation {
	s.RLock()
	defer s.RUnlock()
	for _, d := range s.delegates {
		if d.AgentID == agentID && d.Status == DelegateStatusAccepted {
			return d
		}
	}
	return nil
}

func (s *Store) GetPendingDelegatesForAgent(agentID string) []*DelegateRelation {
	s.RLock()
	defer s.RUnlock()
	var result []*DelegateRelation
	for _, d := range s.delegates {
		if d.AgentID == agentID && d.Status == DelegateStatusPending {
			result = append(result, d)
		}
	}
	return result
}

func (s *Store) AddElection(e *Election) {
	s.Lock()
	defer s.Unlock()
	s.elections[e.ID] = e
}

func (s *Store) GetElection(id string) *Election {
	s.RLock()
	defer s.RUnlock()
	return s.elections[id]
}

func (s *Store) GetElections() []*Election {
	s.RLock()
	defer s.RUnlock()
	result := make([]*Election, 0, len(s.elections))
	for _, e := range s.elections {
		result = append(result, e)
	}
	return result
}

func (s *Store) UpdateElection(e *Election) {
	s.Lock()
	defer s.Unlock()
	s.elections[e.ID] = e
}

func (s *Store) AddVote(electionID string, v *Vote) {
	s.Lock()
	defer s.Unlock()
	s.votes[electionID] = append(s.votes[electionID], v)
}

func (s *Store) GetVotes(electionID string) []*Vote {
	s.RLock()
	defer s.RUnlock()
	votes := s.votes[electionID]
	result := make([]*Vote, len(votes))
	copy(result, votes)
	return result
}

func (s *Store) HasVoted(electionID, voterID string) bool {
	s.RLock()
	defer s.RUnlock()
	for _, v := range s.votes[electionID] {
		if v.VoterID == voterID {
			return true
		}
	}
	return false
}

func (s *Store) AddProposal(p *Proposal) {
	s.Lock()
	defer s.Unlock()
	s.proposals[p.ID] = p
}

func (s *Store) GetProposal(id string) *Proposal {
	s.RLock()
	defer s.RUnlock()
	return s.proposals[id]
}

func (s *Store) GetProposals() []*Proposal {
	s.RLock()
	defer s.RUnlock()
	result := make([]*Proposal, 0, len(s.proposals))
	for _, p := range s.proposals {
		result = append(result, p)
	}
	return result
}

func (s *Store) UpdateProposal(p *Proposal) {
	s.Lock()
	defer s.Unlock()
	s.proposals[p.ID] = p
}

func (s *Store) AddAnnouncement(a *Announcement) {
	s.Lock()
	defer s.Unlock()
	s.announcements[a.ID] = a
}

func (s *Store) GetAnnouncement(id string) *Announcement {
	s.RLock()
	defer s.RUnlock()
	return s.announcements[id]
}

func (s *Store) GetAnnouncements() []*Announcement {
	s.RLock()
	defer s.RUnlock()
	result := make([]*Announcement, 0, len(s.announcements))
	for _, a := range s.announcements {
		result = append(result, a)
	}
	return result
}

func (s *Store) UpdateAnnouncement(a *Announcement) {
	s.Lock()
	defer s.Unlock()
	s.announcements[a.ID] = a
}
