package core

import (
	"complaint-system/common"
	"sync"
	"time"
)

type Ticket struct {
	mu                    sync.RWMutex
	TicketNo              string
	ComplainerName        string
	ContactPhone          string
	Content               string
	ComplaintChannel      common.ComplaintChannel
	ComplaintType         common.ComplaintType
	Region                string
	Status                common.TicketStatus
	UrgencyLevel          *common.UrgencyLevel
	ResponsibleDepartment string
	ResponsiblePerson     string
	DispatcherID          string
	ProcessingResult      string
	ReviewResult          *common.ReviewResult
	ReviewRemark          string
	RetryCount            int
	IsEscalated           bool
	IsTimeout             bool
	ExternalSystemID      string
	CreatedAt             time.Time
	DispatchedAt          *time.Time
	ProcessedAt           *time.Time
	ClosedAt              *time.Time
	ResponseDeadline      time.Time
	LastReviewTime        *time.Time
}

type TicketStore struct {
	mu      sync.RWMutex
	tickets map[string]*Ticket
	dailySeq map[string]int
	dayNow  string
}

func NewTicketStore() *TicketStore {
	return &TicketStore{
		tickets:  make(map[string]*Ticket),
		dailySeq: make(map[string]int),
	}
}

func (s *TicketStore) add(t *Ticket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[t.TicketNo] = t
}

func (s *TicketStore) get(no string) (*Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[no]
	return t, ok
}

func (s *TicketStore) list() []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Ticket, 0, len(s.tickets))
	for _, t := range s.tickets {
		result = append(result, t)
	}
	return result
}

func (s *TicketStore) nextSeq(day string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	seq := s.dailySeq[day]
	seq++
	s.dailySeq[day] = seq
	return seq
}
