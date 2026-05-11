package core

import (
	"sync"
	"time"

	"community-activity-platform/pkg/common"
)

type Storage struct {
	participants     map[string]*common.Participant
	activities       map[string]*common.Activity
	registrations    map[string]*common.Registration
	feedbacks        map[string]*common.Feedback
	activitySummaries map[string]*common.ActivitySummary
	analysisTodos    map[string]*common.AnalysisTodo

	participantsByPhone  map[string]string
	participantsByRoom   map[string]string
	registrationsByActivity map[string]map[string]string
	registrationsByParticipant map[string]map[string]string
	feedbacksByActivity    map[string]map[string]string
	waitlistByActivity     map[string][]string

	mu sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		participants:           make(map[string]*common.Participant),
		activities:             make(map[string]*common.Activity),
		registrations:          make(map[string]*common.Registration),
		feedbacks:              make(map[string]*common.Feedback),
		activitySummaries:      make(map[string]*common.ActivitySummary),
		analysisTodos:          make(map[string]*common.AnalysisTodo),
		participantsByPhone:    make(map[string]string),
		participantsByRoom:     make(map[string]string),
		registrationsByActivity: make(map[string]map[string]string),
		registrationsByParticipant: make(map[string]map[string]string),
		feedbacksByActivity:    make(map[string]map[string]string),
		waitlistByActivity:     make(map[string][]string),
	}
}

func (s *Storage) StoreParticipant(p *common.Participant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.participants[p.ID] = p
	s.participantsByPhone[p.Phone] = p.ID
	s.participantsByRoom[p.RoomNumber] = p.ID
}

func (s *Storage) GetParticipant(id string) (*common.Participant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.participants[id]
	return p, ok
}

func (s *Storage) GetParticipantByPhone(phone string) (*common.Participant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.participantsByPhone[phone]
	if !ok {
		return nil, false
	}
	return s.participants[id], true
}

func (s *Storage) GetParticipantByRoom(room string) (*common.Participant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.participantsByRoom[room]
	if !ok {
		return nil, false
	}
	return s.participants[id], true
}

func (s *Storage) PhoneExists(phone string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.participantsByPhone[phone]
	return ok
}

func (s *Storage) RoomExists(room string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.participantsByRoom[room]
	return ok
}

func (s *Storage) StoreActivity(a *common.Activity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activities[a.ID] = a
}

func (s *Storage) GetActivity(id string) (*common.Activity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.activities[id]
	return a, ok
}

func (s *Storage) ListActivities(filter func(*common.Activity) bool) []*common.Activity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*common.Activity{}
	for _, a := range s.activities {
		if filter == nil || filter(a) {
			result = append(result, a)
		}
	}
	return result
}

func (s *Storage) StoreRegistration(r *common.Registration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrations[r.ID] = r

	if _, ok := s.registrationsByActivity[r.ActivityID]; !ok {
		s.registrationsByActivity[r.ActivityID] = make(map[string]string)
	}
	s.registrationsByActivity[r.ActivityID][r.ParticipantID] = r.ID

	if _, ok := s.registrationsByParticipant[r.ParticipantID]; !ok {
		s.registrationsByParticipant[r.ParticipantID] = make(map[string]string)
	}
	s.registrationsByParticipant[r.ParticipantID][r.ActivityID] = r.ID
}

func (s *Storage) GetRegistration(id string) (*common.Registration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.registrations[id]
	return r, ok
}

func (s *Storage) GetRegistrationByParticipantActivity(participantID, activityID string) (*common.Registration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	partMap, ok := s.registrationsByParticipant[participantID]
	if !ok {
		return nil, false
	}
	regID, ok := partMap[activityID]
	if !ok {
		return nil, false
	}
	return s.registrations[regID], true
}

func (s *Storage) ListRegistrationsByActivity(activityID string, filter func(*common.Registration) bool) []*common.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*common.Registration{}
	partMap, ok := s.registrationsByActivity[activityID]
	if !ok {
		return result
	}
	for _, regID := range partMap {
		r := s.registrations[regID]
		if filter == nil || filter(r) {
			result = append(result, r)
		}
	}
	return result
}

func (s *Storage) CountRegistrationsByActivity(activityID string, status common.RegistrationStatus) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	partMap, ok := s.registrationsByActivity[activityID]
	if !ok {
		return 0
	}
	for _, regID := range partMap {
		r := s.registrations[regID]
		if r.Status == status {
			count++
		}
	}
	return count
}

func (s *Storage) AddToWaitlist(activityID, registrationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitlistByActivity[activityID] = append(s.waitlistByActivity[activityID], registrationID)
}

func (s *Storage) GetWaitlist(activityID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.waitlistByActivity[activityID]...)
}

func (s *Storage) RemoveFromWaitlist(activityID, registrationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	waitlist, ok := s.waitlistByActivity[activityID]
	if !ok {
		return
	}
	newWaitlist := []string{}
	for _, id := range waitlist {
		if id != registrationID {
			newWaitlist = append(newWaitlist, id)
		}
	}
	s.waitlistByActivity[activityID] = newWaitlist
}

func (s *Storage) PopWaitlist(activityID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	waitlist, ok := s.waitlistByActivity[activityID]
	if !ok || len(waitlist) == 0 {
		return "", false
	}
	first := waitlist[0]
	s.waitlistByActivity[activityID] = waitlist[1:]
	return first, true
}

func (s *Storage) ClearWaitlist(activityID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitlistByActivity[activityID] = []string{}
}

func (s *Storage) StoreFeedback(f *common.Feedback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedbacks[f.ID] = f

	if _, ok := s.feedbacksByActivity[f.ActivityID]; !ok {
		s.feedbacksByActivity[f.ActivityID] = make(map[string]string)
	}
	s.feedbacksByActivity[f.ActivityID][f.ParticipantID] = f.ID
}

func (s *Storage) GetFeedback(id string) (*common.Feedback, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.feedbacks[id]
	return f, ok
}

func (s *Storage) GetFeedbackByParticipantActivity(participantID, activityID string) (*common.Feedback, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	partMap, ok := s.feedbacksByActivity[activityID]
	if !ok {
		return nil, false
	}
	fbID, ok := partMap[participantID]
	if !ok {
		return nil, false
	}
	return s.feedbacks[fbID], true
}

func (s *Storage) ListFeedbacksByActivity(activityID string) []*common.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*common.Feedback{}
	partMap, ok := s.feedbacksByActivity[activityID]
	if !ok {
		return result
	}
	for _, fbID := range partMap {
		result = append(result, s.feedbacks[fbID])
	}
	return result
}

func (s *Storage) StoreActivitySummary(summary *common.ActivitySummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activitySummaries[summary.ActivityID] = summary
}

func (s *Storage) GetActivitySummary(activityID string) (*common.ActivitySummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.activitySummaries[activityID]
	return summary, ok
}

func (s *Storage) StoreAnalysisTodo(todo *common.AnalysisTodo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analysisTodos[todo.ID] = todo
}

func (s *Storage) ListAnalysisTodos() []*common.AnalysisTodo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*common.AnalysisTodo{}
	for _, t := range s.analysisTodos {
		result = append(result, t)
	}
	return result
}

func (s *Storage) Now() time.Time {
	return time.Now()
}
