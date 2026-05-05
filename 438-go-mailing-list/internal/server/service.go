package server

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"

	"go-mailing-list/pkg/protocol"
)

type Service struct {
	store     *Store
	queue     *SendQueue
	scheduler *Scheduler
	alerts    *AlertManager

	mu sync.Mutex
}

func NewService(store *Store) *Service {
	queue := NewSendQueue()
	alertMgr := NewAlertManager(store)
	scheduler := NewScheduler(store, queue, alertMgr)

	queue.SetStore(store)
	queue.SetAlertManager(alertMgr)

	service := &Service{
		store:     store,
		queue:     queue,
		scheduler: scheduler,
		alerts:    alertMgr,
	}

	scheduler.SetService(service)
	return service
}

func (s *Service) Start() {
	s.queue.Start()
	s.scheduler.Start()
}

func (s *Service) Stop() {
	s.queue.Stop()
	s.scheduler.Stop()
}

func (s *Service) CreateMailingList(name, description string) (*protocol.MailingList, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	list := &protocol.MailingList{
		ID:          generateID(),
		Name:        name,
		Description: description,
		IsPaused:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.store.CreateMailingList(list)
	return list, nil
}

func (s *Service) GetMailingList(id string) (*protocol.MailingList, error) {
	list, exists := s.store.GetMailingList(id)
	if !exists {
		return nil, ErrListNotFound
	}
	return list, nil
}

func (s *Service) ListMailingLists() []*protocol.MailingList {
	return s.store.ListMailingLists()
}

func (s *Service) PauseList(listID string) error {
	list, exists := s.store.GetMailingList(listID)
	if !exists {
		return ErrListNotFound
	}
	list.IsPaused = true
	s.store.UpdateMailingList(list)
	return nil
}

func (s *Service) ResumeList(listID string) error {
	list, exists := s.store.GetMailingList(listID)
	if !exists {
		return ErrListNotFound
	}
	list.IsPaused = false
	s.store.UpdateMailingList(list)
	return nil
}

func (s *Service) Subscribe(listID, email, name string, customFields map[string]string) (*protocol.SubscriberListEntry, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmailFormat
	}

	entry := &protocol.SubscriberListEntry{
		Subscriber: protocol.Subscriber{
			Email:           email,
			Name:            name,
			Status:          protocol.StatusSubscribed,
			RegisteredAt:    time.Now(),
			InvalidAttempts: 0,
			CustomFields:    customFields,
		},
	}

	err := s.store.Subscribe(listID, entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Service) Unsubscribe(listID, email string) error {
	return s.store.Unsubscribe(listID, email)
}

func (s *Service) GetSubscriber(listID, email string) (*protocol.SubscriberListEntry, bool) {
	return s.store.GetSubscriber(listID, email)
}

func (s *Service) ListSubscribers(listID string) ([]*protocol.SubscriberListEntry, error) {
	return s.store.ListSubscribers(listID)
}

func (s *Service) ImportSubscribers(listID, filePath string) (total, imported, skipped int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	header := scanner.Text()
	emailCol := -1
	nameCol := -1

	columns := parseCSVLine(header)
	for i, col := range columns {
		col = strings.ToLower(strings.TrimSpace(col))
		if col == "email" || col == "邮箱" {
			emailCol = i
		} else if col == "name" || col == "姓名" || col == "name" {
			nameCol = i
		}
	}

	if emailCol == -1 {
		return 0, 0, 0, ErrInvalidEmailFormat
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := parseCSVLine(line)
		if len(parts) <= emailCol {
			skipped++
			total++
			continue
		}

		email := strings.TrimSpace(parts[emailCol])
		if !isValidEmail(email) {
			skipped++
			total++
			continue
		}

		name := ""
		if nameCol != -1 && len(parts) > nameCol {
			name = strings.TrimSpace(parts[nameCol])
		}

		_, err := s.Subscribe(listID, email, name, nil)
		if err != nil {
			skipped++
		} else {
			imported++
		}
		total++
	}

	if err := scanner.Err(); err != nil {
		return total, imported, skipped, err
	}

	return total, imported, skipped, nil
}

func (s *Service) CreateTemplate(name, subject, htmlBody, textBody string, trackOpens, trackClicks bool, variables map[string]string) (*protocol.EmailTemplate, error) {
	template := &protocol.EmailTemplate{
		ID:          generateID(),
		Name:        name,
		Subject:     subject,
		HTMLBody:    htmlBody,
		TextBody:    textBody,
		TrackOpens:  trackOpens,
		TrackClicks: trackClicks,
		Variables:   variables,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.store.CreateTemplate(template)
	return template, nil
}

func (s *Service) GetTemplate(id string) (*protocol.EmailTemplate, error) {
	template, exists := s.store.GetTemplate(id)
	if !exists {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

func (s *Service) ListTemplates() []*protocol.EmailTemplate {
	return s.store.ListTemplates()
}

func (s *Service) SendCampaign(req *protocol.SendCampaignRequest) (*protocol.SendTask, error) {
	list, exists := s.store.GetMailingList(req.ListID)
	if !exists {
		return nil, ErrListNotFound
	}

	if list.IsPaused {
		return nil, ErrListPaused
	}

	var subject, htmlBody, textBody string
	if req.TemplateID != "" {
		template, err := s.GetTemplate(req.TemplateID)
		if err != nil {
			return nil, err
		}
		subject = template.Subject
		htmlBody = template.HTMLBody
		textBody = template.TextBody
	} else {
		subject = req.Subject
		htmlBody = req.HTMLBody
		textBody = req.TextBody
	}

	subscribers, err := s.store.GetActiveSubscribers(req.ListID)
	if err != nil {
		return nil, err
	}

	if len(subscribers) == 0 {
		return nil, ErrNoActiveSubscribers
	}

	if req.IsABTest {
		if req.SubjectA == "" || req.SubjectB == "" {
			return nil, ErrABTestRequiresSubjects
		}
		return s.createABTestTask(req.ListID, subject, htmlBody, textBody, req.ScheduledAt, req.SubjectA, req.SubjectB, subscribers)
	}

	return s.createSendTask(req.ListID, subject, htmlBody, textBody, req.ScheduledAt, subscribers)
}

func (s *Service) createSendTask(listID, subject, htmlBody, textBody string, scheduledAt *time.Time, subscribers []*protocol.SubscriberListEntry) (*protocol.SendTask, error) {
	task := &protocol.SendTask{
		ID:              generateID(),
		ListID:          listID,
		Subject:         subject,
		HTMLBody:        htmlBody,
		TextBody:        textBody,
		ScheduledAt:     scheduledAt,
		Status:          protocol.TaskStatusPending,
		TotalRecipients: len(subscribers),
		SentCount:       0,
		FailedCount:     0,
		BouncedCount:    0,
		IsABTest:        false,
		CreatedAt:       time.Now(),
	}

	if scheduledAt != nil {
		task.Status = protocol.TaskStatusScheduled
	}

	s.store.CreateTask(task)

	if scheduledAt == nil {
		s.queue.AddTask(task, subscribers)
	}

	return task, nil
}

func (s *Service) createABTestTask(listID, subject, htmlBody, textBody string, scheduledAt *time.Time, subjectA, subjectB string, subscribers []*protocol.SubscriberListEntry) (*protocol.SendTask, error) {
	emails := make([]string, 0, len(subscribers))
	for _, sub := range subscribers {
		emails = append(emails, sub.Email)
	}

	shuffleSlice(emails)

	mid := len(emails) / 2
	groupA := emails[:mid]
	groupB := emails[mid:]

	abTest := &protocol.ABTest{
		ID:              generateID(),
		SubjectA:        subjectA,
		SubjectB:        subjectB,
		TotalRecipients: len(emails),
		GroupAEmails:    groupA,
		GroupBEmails:    groupB,
		OpensA:          0,
		OpensB:          0,
		ClicksA:         0,
		ClicksB:         0,
		Status:          "created",
		CreatedAt:       time.Now(),
	}

	task := &protocol.SendTask{
		ID:              generateID(),
		ListID:          listID,
		Subject:         subject,
		HTMLBody:        htmlBody,
		TextBody:        textBody,
		ScheduledAt:     scheduledAt,
		Status:          protocol.TaskStatusPending,
		TotalRecipients: len(subscribers),
		SentCount:       0,
		FailedCount:     0,
		BouncedCount:    0,
		IsABTest:        true,
		ABTestID:        abTest.ID,
		CreatedAt:       time.Now(),
	}

	if scheduledAt != nil {
		task.Status = protocol.TaskStatusScheduled
	}

	abTest.TaskID = task.ID

	s.store.CreateTask(task)
	s.store.CreateABTest(abTest)

	if scheduledAt == nil {
		go s.runABTest(task, abTest, subscribers)
	}

	return task, nil
}

func (s *Service) runABTest(task *protocol.SendTask, abTest *protocol.ABTest, subscribers []*protocol.SubscriberListEntry) {
	emailToSubscriber := make(map[string]*protocol.SubscriberListEntry)
	for _, sub := range subscribers {
		emailToSubscriber[sub.Email] = sub
	}

	var groupASubs []*protocol.SubscriberListEntry
	for _, email := range abTest.GroupAEmails {
		if sub, exists := emailToSubscriber[email]; exists {
			groupASubs = append(groupASubs, sub)
		}
	}

	var groupBSubs []*protocol.SubscriberListEntry
	for _, email := range abTest.GroupBEmails {
		if sub, exists := emailToSubscriber[email]; exists {
			groupBSubs = append(groupBSubs, sub)
		}
	}

	s.sendABTestGroup(task, abTest, "A", abTest.SubjectA, groupASubs)
	s.sendABTestGroup(task, abTest, "B", abTest.SubjectB, groupBSubs)

	time.Sleep(5 * time.Minute)

	abTest.Status = "evaluating"
	s.store.UpdateABTest(abTest)

	var winner string
	if abTest.OpensA >= abTest.OpensB {
		winner = "A"
	} else {
		winner = "B"
	}

	abTest.Winner = winner
	abTest.Status = "completed"
	now := time.Now()
	abTest.CompletedAt = &now
	s.store.UpdateABTest(abTest)
}

func (s *Service) sendABTestGroup(task *protocol.SendTask, abTest *protocol.ABTest, group, subject string, subscribers []*protocol.SubscriberListEntry) {
	for _, sub := range subscribers {
		if !isValidEmail(sub.Email) {
			record := &protocol.SendRecord{
				ID:        generateID(),
				TaskID:    task.ID,
				Email:     sub.Email,
				Status:    protocol.SendStatusInvalid,
				ABGroup:   group,
				CreatedAt: time.Now(),
			}
			s.store.AddSendRecord(record)
			continue
		}

		window := time.Duration(protocol.ResendWindowHours) * time.Hour
		if s.store.WasSentRecently(sub.Email, window) {
			record := &protocol.SendRecord{
				ID:        generateID(),
				TaskID:    task.ID,
				Email:     sub.Email,
				Status:    protocol.SendStatusPending,
				ABGroup:   group,
				CreatedAt: time.Now(),
			}
			s.store.AddSendRecord(record)
			continue
		}

		status := protocol.SendStatusSuccess
		now := time.Now()
		record := &protocol.SendRecord{
			ID:        generateID(),
			TaskID:    task.ID,
			Email:     sub.Email,
			Status:    status,
			SentAt:    &now,
			ABGroup:   group,
			CreatedAt: time.Now(),
		}
		s.store.AddSendRecord(record)
		s.store.RecordSendHistory(task.ID, sub.Email)
	}
}

func (s *Service) GetTask(id string) (*protocol.SendTask, error) {
	task, exists := s.store.GetTask(id)
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *Service) ListTasks() []*protocol.SendTask {
	return s.store.ListTasks()
}

func (s *Service) GetSendRecords(taskID string) []*protocol.SendRecord {
	return s.store.GetSendRecords(taskID)
}

func (s *Service) TrackOpen(taskID, email string) error {
	record, exists := s.store.GetSendRecordByEmail(taskID, email)
	if !exists {
		return nil
	}

	now := time.Now()
	record.OpenedAt = &now
	s.store.UpdateSendRecord(record)

	if record.ABGroup != "" {
		abTests := s.store.ListABTests()
		for _, abTest := range abTests {
			if abTest.TaskID == taskID {
				if record.ABGroup == "A" {
					abTest.OpensA++
				} else {
					abTest.OpensB++
				}
				s.store.UpdateABTest(abTest)
				break
			}
		}
	}

	return nil
}

func (s *Service) TrackClick(taskID, email, url string) error {
	record, exists := s.store.GetSendRecordByEmail(taskID, email)
	if !exists {
		return nil
	}

	now := time.Now()
	record.ClickedAt = &now
	s.store.UpdateSendRecord(record)

	if record.ABGroup != "" {
		abTests := s.store.ListABTests()
		for _, abTest := range abTests {
			if abTest.TaskID == taskID {
				if record.ABGroup == "A" {
					abTest.ClicksA++
				} else {
					abTest.ClicksB++
				}
				s.store.UpdateABTest(abTest)
				break
			}
		}
	}

	return nil
}

func (s *Store) ListABTests() []*protocol.ABTest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tests := make([]*protocol.ABTest, 0, len(s.abTests))
	for _, t := range s.abTests {
		tests = append(tests, t)
	}
	return tests
}

func (s *Service) GetAlerts() []*protocol.Alert {
	return s.store.GetAlerts()
}

func (s *Service) ResolveAlert(alertID string) bool {
	return s.store.ResolveAlert(alertID)
}

func (s *Service) GetStats() (totalLists, totalSubs, totalTasks, totalSent, totalBounced, activeAlerts int) {
	return s.store.GetStats()
}
