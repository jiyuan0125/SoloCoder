package storage

import (
	"database/sql"
	"fmt"
	"go-mq-admin/internal/models"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
	mu sync.Mutex
}

const schema = `
CREATE TABLE IF NOT EXISTS topics (
	name TEXT PRIMARY KEY,
	ttl INTEGER NOT NULL DEFAULT 3600000,
	max_dead_messages INTEGER NOT NULL DEFAULT 100,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	topic_name TEXT NOT NULL,
	mode TEXT NOT NULL DEFAULT 'cluster',
	ack_timeout INTEGER NOT NULL DEFAULT 30000,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(name, topic_name),
	FOREIGN KEY (topic_name) REFERENCES topics(name) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	topic_name TEXT NOT NULL,
	payload TEXT NOT NULL,
	published_at DATETIME NOT NULL,
	expire_at DATETIME NOT NULL,
	FOREIGN KEY (topic_name) REFERENCES topics(name) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS queue_messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	subscription_id INTEGER NOT NULL,
	message_id TEXT NOT NULL,
	position INTEGER NOT NULL,
	FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE,
	FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS in_flight_messages (
	id TEXT PRIMARY KEY,
	subscription_id INTEGER NOT NULL,
	message_id TEXT NOT NULL,
	consumer_id TEXT NOT NULL,
	delivered_at DATETIME NOT NULL,
	deadline_at DATETIME NOT NULL,
	delivery_count INTEGER NOT NULL DEFAULT 1,
	FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE,
	FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS dead_letters (
	id TEXT PRIMARY KEY,
	topic_name TEXT NOT NULL,
	payload TEXT NOT NULL,
	original_message_id TEXT NOT NULL,
	published_at DATETIME NOT NULL,
	dead_at DATETIME NOT NULL,
	reason TEXT NOT NULL,
	position INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS consumers (
	id TEXT PRIMARY KEY,
	subscription_id INTEGER NOT NULL,
	last_heartbeat DATETIME NOT NULL,
	FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS dead_letter_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	topic_name TEXT NOT NULL,
	dead_letter_id TEXT NOT NULL,
	dropped_at DATETIME NOT NULL,
	reason TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_expire ON messages(expire_at);
CREATE INDEX IF NOT EXISTS idx_queue_messages_sub ON queue_messages(subscription_id);
CREATE INDEX IF NOT EXISTS idx_in_flight_deadline ON in_flight_messages(deadline_at);
CREATE INDEX IF NOT EXISTS idx_dead_letters_topic ON dead_letters(topic_name);
CREATE INDEX IF NOT EXISTS idx_consumers_sub ON consumers(subscription_id);
`

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=10000&_journal_mode=WAL&_txlock=IMMEDIATE")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &Storage{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) init() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

func (s *Storage) DB() *sql.DB {
	return s.db
}

func (s *Storage) Lock() {
	s.mu.Lock()
}

func (s *Storage) Unlock() {
	s.mu.Unlock()
}

func (s *Storage) CreateTopic(topic *models.Topic) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		"INSERT INTO topics (name, ttl, max_dead_messages) VALUES (?, ?, ?)",
		topic.Name, topic.TTL, topic.MaxDeadMessages,
	)
	return err
}

func (s *Storage) GetTopic(name string) (*models.Topic, error) {
	var t models.Topic
	err := s.db.QueryRow(
		"SELECT name, ttl, max_dead_messages, created_at FROM topics WHERE name = ?",
		name,
	).Scan(&t.Name, &t.TTL, &t.MaxDeadMessages, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Storage) ListTopics() ([]*models.Topic, error) {
	rows, err := s.db.Query("SELECT name, ttl, max_dead_messages, created_at FROM topics ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []*models.Topic
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.Name, &t.TTL, &t.MaxDeadMessages, &t.CreatedAt); err != nil {
			return nil, err
		}
		topics = append(topics, &t)
	}
	return topics, nil
}

func (s *Storage) DeleteTopic(name string) error {
	_, err := s.db.Exec("DELETE FROM topics WHERE name = ?", name)
	return err
}

func (s *Storage) TopicExists(name string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM topics WHERE name = ?", name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Storage) HasActiveSubscriptions(topicName string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM subscriptions WHERE topic_name = ?",
		topicName,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Storage) CreateSubscription(sub *models.Subscription) error {
	result, err := s.db.Exec(
		"INSERT INTO subscriptions (name, topic_name, mode, ack_timeout) VALUES (?, ?, ?, ?)",
		sub.Name, sub.TopicName, sub.Mode, sub.AckTimeout,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO queue_messages (subscription_id, message_id, position)
		 SELECT ?, id, COALESCE((SELECT MAX(position) FROM queue_messages WHERE subscription_id = ?), 0) + ROW_NUMBER() OVER (ORDER BY published_at)
		 FROM messages WHERE topic_name = ? AND expire_at > ?`,
		id, id, sub.TopicName, time.Now(),
	)
	return err
}

func (s *Storage) GetSubscription(topicName, subName string) (*models.Subscription, int64, error) {
	var sub models.Subscription
	var id int64
	err := s.db.QueryRow(
		"SELECT id, name, topic_name, mode, ack_timeout, created_at FROM subscriptions WHERE topic_name = ? AND name = ?",
		topicName, subName,
	).Scan(&id, &sub.Name, &sub.TopicName, &sub.Mode, &sub.AckTimeout, &sub.CreatedAt)
	if err != nil {
		return nil, 0, err
	}
	return &sub, id, nil
}

func (s *Storage) ListSubscriptions(topicName string) ([]*models.Subscription, error) {
	rows, err := s.db.Query(
		"SELECT name, topic_name, mode, ack_timeout, created_at FROM subscriptions WHERE topic_name = ? ORDER BY name",
		topicName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		var s models.Subscription
		if err := rows.Scan(&s.Name, &s.TopicName, &s.Mode, &s.AckTimeout, &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, nil
}

func (s *Storage) DeleteSubscription(topicName, subName string) error {
	_, err := s.db.Exec(
		"DELETE FROM subscriptions WHERE topic_name = ? AND name = ?",
		topicName, subName,
	)
	return err
}

func (s *Storage) SubscriptionExists(topicName, subName string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM subscriptions WHERE topic_name = ? AND name = ?",
		topicName, subName,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Storage) PublishMessage(msg *models.Message, tx *sql.Tx) error {
	exec := s.db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec(
		"INSERT INTO messages (id, topic_name, payload, published_at, expire_at) VALUES (?, ?, ?, ?, ?)",
		msg.ID, msg.TopicName, msg.Payload, msg.PublishedAt, msg.ExpireAt,
	)
	return err
}

func (s *Storage) GetTopicSubscriptions(topicName string) ([]struct {
	ID   int64
	Mode string
}, error) {
	rows, err := s.db.Query(
		"SELECT id, mode FROM subscriptions WHERE topic_name = ?",
		topicName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []struct {
		ID   int64
		Mode string
	}
	for rows.Next() {
		var sub struct {
			ID   int64
			Mode string
		}
		if err := rows.Scan(&sub.ID, &sub.Mode); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *Storage) EnqueueMessage(subscriptionID int64, messageID string, tx *sql.Tx) error {
	exec := s.db.Exec
	query := s.db.QueryRow
	if tx != nil {
		exec = tx.Exec
		query = tx.QueryRow
	}

	var position int
	err := query(
		"SELECT COALESCE(MAX(position), 0) + 1 FROM queue_messages WHERE subscription_id = ?",
		subscriptionID,
	).Scan(&position)
	if err != nil {
		return err
	}

	_, err = exec(
		"INSERT INTO queue_messages (subscription_id, message_id, position) VALUES (?, ?, ?)",
		subscriptionID, messageID, position,
	)
	return err
}

func (s *Storage) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}
