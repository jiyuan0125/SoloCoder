package subscription

import (
	"errors"
	"go-mq-admin/internal/models"
	"go-mq-admin/internal/storage"
)

var (
	ErrInvalidMode        = errors.New("mode must be 'broadcast' or 'cluster'")
	ErrSubscriptionExists = errors.New("subscription already exists")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Manager struct {
	store *storage.Storage
}

func NewManager(store *storage.Storage) *Manager {
	return &Manager{store: store}
}

func (m *Manager) validateMode(mode string) error {
	if mode != "broadcast" && mode != "cluster" {
		return ErrInvalidMode
	}
	return nil
}

func (m *Manager) Create(topicName, name, mode string, ackTimeout int64) (*models.Subscription, error) {
	if mode == "" {
		mode = "cluster"
	}
	if err := m.validateMode(mode); err != nil {
		return nil, err
	}

	if ackTimeout <= 0 {
		ackTimeout = 30000
	}

	exists, err := m.store.SubscriptionExists(topicName, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSubscriptionExists
	}

	sub := &models.Subscription{
		Name:       name,
		TopicName:  topicName,
		Mode:       mode,
		AckTimeout: ackTimeout,
	}

	if err := m.store.CreateSubscription(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (m *Manager) Get(topicName, name string) (*models.Subscription, error) {
	sub, _, err := m.store.GetSubscription(topicName, name)
	return sub, err
}

func (m *Manager) List(topicName string) ([]*models.Subscription, error) {
	return m.store.ListSubscriptions(topicName)
}

func (m *Manager) Delete(topicName, name string) error {
	exists, err := m.store.SubscriptionExists(topicName, name)
	if err != nil {
		return err
	}
	if !exists {
		return ErrSubscriptionNotFound
	}
	return m.store.DeleteSubscription(topicName, name)
}

func (m *Manager) GetStats(topicName, subName string) (*models.SubscriptionStats, error) {
	sub, subID, err := m.store.GetSubscription(topicName, subName)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()
	stats := &models.SubscriptionStats{
		SubscriptionName: subName,
		TopicName:        topicName,
		Mode:             sub.Mode,
	}

	err = db.QueryRow(
		`SELECT 
			(SELECT COUNT(*) FROM queue_messages WHERE subscription_id = ?),
			(SELECT COUNT(*) FROM in_flight_messages WHERE subscription_id = ?),
			(SELECT COUNT(*) FROM consumers WHERE subscription_id = ? AND last_heartbeat > datetime('now', '-30 seconds'))`,
		subID, subID, subID,
	).Scan(
		&stats.QueuedMessages,
		&stats.InFlightMessages,
		&stats.ActiveConsumers,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
