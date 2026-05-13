package topic

import (
	"database/sql"
	"errors"
	"go-mq-admin/internal/models"
	"go-mq-admin/internal/storage"
	"regexp"
)

var (
	ErrInvalidTopicName  = errors.New("topic name can only contain letters, numbers, and underscores")
	ErrTopicExists       = errors.New("topic already exists")
	ErrTopicNotFound     = errors.New("topic not found")
	ErrActiveSubscriptions = errors.New("topic has active subscriptions")
	validTopicName       = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

type Manager struct {
	store *storage.Storage
}

func NewManager(store *storage.Storage) *Manager {
	return &Manager{store: store}
}

func (m *Manager) ValidateName(name string) error {
	if !validTopicName.MatchString(name) {
		return ErrInvalidTopicName
	}
	return nil
}

func (m *Manager) Create(name string, ttl int64, maxDeadMessages int) (*models.Topic, error) {
	if err := m.ValidateName(name); err != nil {
		return nil, err
	}

	exists, err := m.store.TopicExists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTopicExists
	}

	if ttl <= 0 {
		ttl = 3600000
	}
	if maxDeadMessages <= 0 {
		maxDeadMessages = 100
	}

	topic := &models.Topic{
		Name:            name,
		TTL:             ttl,
		MaxDeadMessages: maxDeadMessages,
	}

	if err := m.store.CreateTopic(topic); err != nil {
		return nil, err
	}

	return topic, nil
}

func (m *Manager) Get(name string) (*models.Topic, error) {
	topic, err := m.store.GetTopic(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTopicNotFound
		}
		return nil, err
	}
	return topic, nil
}

func (m *Manager) List() ([]*models.Topic, error) {
	return m.store.ListTopics()
}

func (m *Manager) Delete(name string) error {
	exists, err := m.store.TopicExists(name)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTopicNotFound
	}

	hasSubs, err := m.store.HasActiveSubscriptions(name)
	if err != nil {
		return err
	}
	if hasSubs {
		return ErrActiveSubscriptions
	}

	return m.store.DeleteTopic(name)
}

func (m *Manager) GetStats(topicName string) (*models.TopicStats, error) {
	exists, err := m.store.TopicExists(topicName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTopicNotFound
	}

	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()
	stats := &models.TopicStats{TopicName: topicName}

	err = db.QueryRow(
		`SELECT 
			(SELECT COUNT(*) FROM messages WHERE topic_name = ?),
			(SELECT COUNT(DISTINCT qm.message_id) 
			 FROM queue_messages qm 
			 JOIN subscriptions s ON qm.subscription_id = s.id 
			 WHERE s.topic_name = ?),
			(SELECT COUNT(*) 
			 FROM in_flight_messages ifm 
			 JOIN subscriptions s ON ifm.subscription_id = s.id 
			 WHERE s.topic_name = ?),
			(SELECT COUNT(*) FROM dead_letters WHERE topic_name = ?),
			(SELECT COUNT(*) FROM subscriptions WHERE topic_name = ?)`,
		topicName, topicName, topicName, topicName, topicName,
	).Scan(
		&stats.TotalMessages,
		&stats.QueuedMessages,
		&stats.InFlightMessages,
		&stats.DeadLetterCount,
		&stats.SubscriptionCount,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
