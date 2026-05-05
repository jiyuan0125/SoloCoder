package server

import (
	"sync"
	"time"

	"go-mailing-list/pkg/protocol"
)

type Store struct {
	mu sync.RWMutex

	mailingLists     map[string]*protocol.MailingList
	subscribers      map[string]map[string]*protocol.SubscriberListEntry
	globalSubscribers map[string]*protocol.Subscriber
	templates        map[string]*protocol.EmailTemplate
	tasks            map[string]*protocol.SendTask
	sendRecords      map[string][]*protocol.SendRecord
	alerts           map[string]*protocol.Alert
	abTests          map[string]*protocol.ABTest
	sendHistory      map[string]map[string]time.Time
}

func NewStore() *Store {
	return &Store{
		mailingLists:      make(map[string]*protocol.MailingList),
		subscribers:       make(map[string]map[string]*protocol.SubscriberListEntry),
		globalSubscribers: make(map[string]*protocol.Subscriber),
		templates:         make(map[string]*protocol.EmailTemplate),
		tasks:             make(map[string]*protocol.SendTask),
		sendRecords:       make(map[string][]*protocol.SendRecord),
		alerts:            make(map[string]*protocol.Alert),
		abTests:           make(map[string]*protocol.ABTest),
		sendHistory:       make(map[string]map[string]time.Time),
	}
}

func (s *Store) CreateMailingList(list *protocol.MailingList) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mailingLists[list.ID] = list
	s.subscribers[list.ID] = make(map[string]*protocol.SubscriberListEntry)
}

func (s *Store) GetMailingList(id string) (*protocol.MailingList, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list, exists := s.mailingLists[id]
	return list, exists
}

func (s *Store) ListMailingLists() []*protocol.MailingList {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lists := make([]*protocol.MailingList, 0, len(s.mailingLists))
	for _, list := range s.mailingLists {
		lists = append(lists, list)
	}
	return lists
}

func (s *Store) UpdateMailingList(list *protocol.MailingList) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list.UpdatedAt = time.Now()
	s.mailingLists[list.ID] = list
}

func (s *Store) Subscribe(listID string, entry *protocol.SubscriberListEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.mailingLists[listID]; !exists {
		return ErrListNotFound
	}

	if s.subscribers[listID] == nil {
		s.subscribers[listID] = make(map[string]*protocol.SubscriberListEntry)
	}

	globalSub, exists := s.globalSubscribers[entry.Email]
	if !exists {
		globalSub = &protocol.Subscriber{
			Email:           entry.Email,
			Name:            entry.Name,
			Status:          protocol.StatusSubscribed,
			RegisteredAt:    time.Now(),
			InvalidAttempts: 0,
			CustomFields:    entry.CustomFields,
		}
		s.globalSubscribers[entry.Email] = globalSub
	}

	entry.Subscriber = *globalSub
	entry.ListID = listID
	entry.SubscribedAt = time.Now()
	entry.Status = protocol.StatusSubscribed

	s.subscribers[listID][entry.Email] = entry
	return nil
}

func (s *Store) Unsubscribe(listID, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.mailingLists[listID]; !exists {
		return ErrListNotFound
	}

	entry, exists := s.subscribers[listID][email]
	if !exists {
		return ErrSubscriberNotFound
	}

	entry.Status = protocol.StatusUnsubscribed
	s.subscribers[listID][email] = entry
	return nil
}

func (s *Store) GetSubscriber(listID, email string) (*protocol.SubscriberListEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if list, exists := s.subscribers[listID]; exists {
		entry, exists := list[email]
		return entry, exists
	}
	return nil, false
}

func (s *Store) ListSubscribers(listID string) ([]*protocol.SubscriberListEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.mailingLists[listID]; !exists {
		return nil, ErrListNotFound
	}

	subs := make([]*protocol.SubscriberListEntry, 0, len(s.subscribers[listID]))
	for _, sub := range s.subscribers[listID] {
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *Store) GetActiveSubscribers(listID string) ([]*protocol.SubscriberListEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.mailingLists[listID]; !exists {
		return nil, ErrListNotFound
	}

	subs := make([]*protocol.SubscriberListEntry, 0)
	for _, sub := range s.subscribers[listID] {
		if sub.Status == protocol.StatusSubscribed {
			subs = append(subs, sub)
		}
	}
	return subs, nil
}

func (s *Store) UpdateInvalidAttempt(email string) (shouldUnsubscribe bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, exists := s.globalSubscribers[email]
	if !exists {
		return false
	}

	sub.InvalidAttempts++
	now := time.Now()
	sub.LastInvalidAt = &now
	s.globalSubscribers[email] = sub

	return sub.InvalidAttempts >= protocol.MaxInvalidAttempts
}

func (s *Store) CreateTemplate(template *protocol.EmailTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[template.ID] = template
}

func (s *Store) GetTemplate(id string) (*protocol.EmailTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	template, exists := s.templates[id]
	return template, exists
}

func (s *Store) ListTemplates() []*protocol.EmailTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	templates := make([]*protocol.EmailTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, t)
	}
	return templates
}

func (s *Store) CreateTask(task *protocol.SendTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *Store) GetTask(id string) (*protocol.SendTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[id]
	return task, exists
}

func (s *Store) ListTasks() []*protocol.SendTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*protocol.SendTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (s *Store) UpdateTask(task *protocol.SendTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *Store) GetPendingScheduledTasks() []*protocol.SendTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	pending := make([]*protocol.SendTask, 0)
	for _, task := range s.tasks {
		if task.Status == protocol.TaskStatusScheduled &&
			task.ScheduledAt != nil &&
			task.ScheduledAt.Before(now) {
			pending = append(pending, task)
		}
	}
	return pending
}

func (s *Store) AddSendRecord(record *protocol.SendRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sendRecords[record.TaskID] == nil {
		s.sendRecords[record.TaskID] = make([]*protocol.SendRecord, 0)
	}
	s.sendRecords[record.TaskID] = append(s.sendRecords[record.TaskID], record)
}

func (s *Store) GetSendRecords(taskID string) []*protocol.SendRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sendRecords[taskID]
}

func (s *Store) GetSendRecordByEmail(taskID, email string) (*protocol.SendRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.sendRecords[taskID]
	for _, r := range records {
		if r.Email == email {
			return r, true
		}
	}
	return nil, false
}

func (s *Store) UpdateSendRecord(record *protocol.SendRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.sendRecords[record.TaskID]
	for i, r := range records {
		if r.ID == record.ID {
			records[i] = record
			break
		}
	}
}

func (s *Store) RecordSendHistory(taskID, email string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sendHistory[email] == nil {
		s.sendHistory[email] = make(map[string]time.Time)
	}
	s.sendHistory[email][taskID] = time.Now()
}

func (s *Store) WasSentRecently(email string, window time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.sendHistory[email]
	if !exists {
		return false
	}

	cutoff := time.Now().Add(-window)
	for _, sentAt := range history {
		if sentAt.After(cutoff) {
			return true
		}
	}
	return false
}

func (s *Store) CreateAlert(alert *protocol.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts[alert.ID] = alert
}

func (s *Store) GetAlerts() []*protocol.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*protocol.Alert, 0, len(s.alerts))
	for _, a := range s.alerts {
		alerts = append(alerts, a)
	}
	return alerts
}

func (s *Store) ResolveAlert(alertID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[alertID]
	if !exists {
		return false
	}
	alert.Resolved = true
	now := time.Now()
	alert.ResolvedAt = &now
	return true
}

func (s *Store) GetActiveAlerts() []*protocol.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*protocol.Alert, 0)
	for _, a := range s.alerts {
		if !a.Resolved {
			alerts = append(alerts, a)
		}
	}
	return alerts
}

func (s *Store) CreateABTest(abTest *protocol.ABTest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.abTests[abTest.ID] = abTest
}

func (s *Store) GetABTest(id string) (*protocol.ABTest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	abTest, exists := s.abTests[id]
	return abTest, exists
}

func (s *Store) UpdateABTest(abTest *protocol.ABTest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.abTests[abTest.ID] = abTest
}

func (s *Store) GetStats() (totalLists, totalSubs, totalTasks, totalSent, totalBounced, activeAlerts int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalLists = len(s.mailingLists)
	totalSubs = len(s.globalSubscribers)
	totalTasks = len(s.tasks)

	for _, records := range s.sendRecords {
		for _, r := range records {
			if r.Status == protocol.SendStatusSuccess {
				totalSent++
			}
			if r.Status == protocol.SendStatusBounced {
				totalBounced++
			}
		}
	}

	for _, a := range s.alerts {
		if !a.Resolved {
			activeAlerts++
		}
	}

	return
}
