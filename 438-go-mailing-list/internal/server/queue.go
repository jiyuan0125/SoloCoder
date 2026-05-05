package server

import (
	"sync"
	"time"

	"go-mailing-list/pkg/protocol"
)

type QueueItem struct {
	Task        *protocol.SendTask
	Subscribers []*protocol.SubscriberListEntry
	Priority    int
}

type SendQueue struct {
	queue     chan *QueueItem
	wg        sync.WaitGroup
	running   bool
	stopChan  chan struct{}
	store     *Store
	alertMgr  *AlertManager

	maxConcurrent int
	rateLimit     time.Duration
}

func NewSendQueue() *SendQueue {
	return &SendQueue{
		queue:         make(chan *QueueItem, 100),
		stopChan:      make(chan struct{}),
		maxConcurrent: 10,
		rateLimit:     100 * time.Millisecond,
	}
}

func (q *SendQueue) SetStore(store *Store) {
	q.store = store
}

func (q *SendQueue) SetAlertManager(alertMgr *AlertManager) {
	q.alertMgr = alertMgr
}

func (q *SendQueue) Start() {
	q.running = true
	go q.process()
}

func (q *SendQueue) Stop() {
	q.running = false
	close(q.stopChan)
	q.wg.Wait()
}

func (q *SendQueue) AddTask(task *protocol.SendTask, subscribers []*protocol.SubscriberListEntry) {
	item := &QueueItem{
		Task:        task,
		Subscribers: subscribers,
		Priority:    0,
	}
	q.queue <- item
}

func (q *SendQueue) process() {
	for q.running {
		select {
		case item := <-q.queue:
			q.wg.Add(1)
			go q.processItem(item)
		case <-q.stopChan:
			return
		}
	}
}

func (q *SendQueue) processItem(item *QueueItem) {
	defer q.wg.Done()

	task := item.Task
	task.Status = protocol.TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	q.store.UpdateTask(task)

	batches := splitIntoBatches(item.Subscribers, protocol.MaxRecipientsPerBatch)

	for i, batch := range batches {
		if i > 0 {
			time.Sleep(time.Duration(protocol.BatchIntervalSeconds) * time.Second)
		}
		q.sendBatch(task, batch)
	}

	task.Status = protocol.TaskStatusCompleted
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	q.store.UpdateTask(task)
}

func (q *SendQueue) sendBatch(task *protocol.SendTask, subscribers []*protocol.SubscriberListEntry) {
	limiter := time.Tick(q.rateLimit)

	for _, sub := range subscribers {
		<-limiter

		if sub.Status != protocol.StatusSubscribed {
			continue
		}

		if !isValidEmail(sub.Email) {
			q.recordInvalidEmail(task, sub)
			continue
		}

		window := time.Duration(protocol.ResendWindowHours) * time.Hour
		if q.store.WasSentRecently(sub.Email, window) {
			continue
		}

		q.sendToSubscriber(task, sub)
	}
}

func (q *SendQueue) sendToSubscriber(task *protocol.SendTask, sub *protocol.SubscriberListEntry) {
	htmlBody := replaceVariables(task.HTMLBody, &sub.Subscriber)
	textBody := replaceVariables(task.TextBody, &sub.Subscriber)
	subject := replaceVariables(task.Subject, &sub.Subscriber)

	_ = htmlBody
	_ = textBody
	_ = subject

	status := protocol.SendStatusSuccess
	now := time.Now()

	record := &protocol.SendRecord{
		ID:        generateID(),
		TaskID:    task.ID,
		Email:     sub.Email,
		Status:    status,
		SentAt:    &now,
		CreatedAt: time.Now(),
	}

	q.store.AddSendRecord(record)
	q.store.RecordSendHistory(task.ID, sub.Email)

	task.SentCount++
	q.store.UpdateTask(task)

	if q.alertMgr != nil {
		q.alertMgr.CheckBounceRate(task.ListID)
	}
}

func (q *SendQueue) recordInvalidEmail(task *protocol.SendTask, sub *protocol.SubscriberListEntry) {
	record := &protocol.SendRecord{
		ID:        generateID(),
		TaskID:    task.ID,
		Email:     sub.Email,
		Status:    protocol.SendStatusInvalid,
		ErrorMsg:  "Invalid email format",
		CreatedAt: time.Now(),
	}

	q.store.AddSendRecord(record)

	shouldUnsubscribe := q.store.UpdateInvalidAttempt(sub.Email)
	if shouldUnsubscribe {
		q.store.Unsubscribe(task.ListID, sub.Email)
	}
}
