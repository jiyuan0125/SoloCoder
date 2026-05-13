package dispatcher

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"go-mq-admin/internal/models"
	"go-mq-admin/internal/storage"
	"time"
)

var (
	ErrEmptyPayload  = errors.New("message payload cannot be empty")
	ErrBatchTooLarge = errors.New("batch size exceeds maximum of 1000")
)

const MaxBatchSize = 1000

type Dispatcher struct {
	store *storage.Storage
}

func NewDispatcher(store *storage.Storage) *Dispatcher {
	return &Dispatcher{store: store}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (d *Dispatcher) Publish(topicName string, payloads []string) ([]string, error) {
	if len(payloads) == 0 {
		return nil, ErrEmptyPayload
	}
	if len(payloads) > MaxBatchSize {
		return nil, ErrBatchTooLarge
	}

	for _, p := range payloads {
		if p == "" {
			return nil, ErrEmptyPayload
		}
	}

	d.store.Lock()
	defer d.store.Unlock()

	exists, err := d.store.TopicExists(topicName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, sql.ErrNoRows
	}

	topic, err := d.store.GetTopic(topicName)
	if err != nil {
		return nil, err
	}

	subs, err := d.store.GetTopicSubscriptions(topicName)
	if err != nil {
		return nil, err
	}

	tx, err := d.store.BeginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	messageIDs := make([]string, len(payloads))
	for i, payload := range payloads {
		msgID := generateID()
		messageIDs[i] = msgID

		msg := &models.Message{
			ID:          msgID,
			TopicName:   topicName,
			Payload:     payload,
			PublishedAt: time.Now(),
			ExpireAt:    time.Now().Add(time.Duration(topic.TTL) * time.Millisecond),
		}

		query := "INSERT INTO messages (id, topic_name, payload, published_at, expire_at) VALUES (?, ?, ?, ?, ?)"
		if _, err := tx.Exec(query, msg.ID, msg.TopicName, msg.Payload, msg.PublishedAt, msg.ExpireAt); err != nil {
			return nil, err
		}

		for _, sub := range subs {
			var pos int
			err := tx.QueryRow(
				"SELECT COALESCE(MAX(position), 0) + 1 FROM queue_messages WHERE subscription_id = ?",
				sub.ID,
			).Scan(&pos)
			if err != nil {
				return nil, err
			}

			_, err = tx.Exec(
				"INSERT INTO queue_messages (subscription_id, message_id, position) VALUES (?, ?, ?)",
				sub.ID, msgID, pos,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return messageIDs, nil
}

func (d *Dispatcher) Pull(topicName, subName, consumerID string, maxMessages int) ([]*models.SubscriptionMessage, error) {
	d.store.Lock()
	defer d.store.Unlock()

	_, subID, err := d.store.GetSubscription(topicName, subName)
	if err != nil {
		return nil, err
	}

	if consumerID == "" {
		consumerID = generateID()
	}

	db := d.store.DB()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO consumers (id, subscription_id, last_heartbeat)
		 VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(id) DO UPDATE SET last_heartbeat = CURRENT_TIMESTAMP`,
		consumerID, subID,
	)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		`SELECT qm.id, qm.message_id, m.payload, m.published_at
		 FROM queue_messages qm
		 JOIN messages m ON qm.message_id = m.id
		 WHERE qm.subscription_id = ?
		 ORDER BY qm.position ASC
		 LIMIT ?`,
		subID, maxMessages,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.SubscriptionMessage
	var queueIDs []int64
	var messageIDs []string

	for rows.Next() {
		var qmID int64
		var msg models.SubscriptionMessage
		var publishedAt time.Time
		if err := rows.Scan(&qmID, &msg.MessageID, &msg.Payload, &publishedAt); err != nil {
			return nil, err
		}
		msg.PublishedAt = publishedAt.Format(time.RFC3339)
		messages = append(messages, &msg)
		queueIDs = append(queueIDs, qmID)
		messageIDs = append(messageIDs, msg.MessageID)
	}

	for i, qmID := range queueIDs {
		inFlightID := generateID()
		_, err = tx.Exec(
			`INSERT INTO in_flight_messages (id, subscription_id, message_id, consumer_id, delivered_at, deadline_at, delivery_count)
			 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, datetime(CURRENT_TIMESTAMP, '+30 seconds'), 1)`,
			inFlightID, subID, messageIDs[i], consumerID,
		)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec("DELETE FROM queue_messages WHERE id = ?", qmID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *Dispatcher) Ack(topicName, subName, messageID string) error {
	d.store.Lock()
	defer d.store.Unlock()

	_, subID, err := d.store.GetSubscription(topicName, subName)
	if err != nil {
		return err
	}

	db := d.store.DB()

	_, err = db.Exec(
		"DELETE FROM in_flight_messages WHERE subscription_id = ? AND message_id = ?",
		subID, messageID,
	)
	return err
}

func (d *Dispatcher) Nack(topicName, subName, messageID string) error {
	d.store.Lock()
	defer d.store.Unlock()

	_, subID, err := d.store.GetSubscription(topicName, subName)
	if err != nil {
		return err
	}

	db := d.store.DB()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM in_flight_messages WHERE subscription_id = ? AND message_id = ?)",
		subID, messageID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return tx.Commit()
	}

	var pos int
	err = tx.QueryRow(
		"SELECT COALESCE(MAX(position), 0) + 1 FROM queue_messages WHERE subscription_id = ?",
		subID,
	).Scan(&pos)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO queue_messages (subscription_id, message_id, position) VALUES (?, ?, ?)",
		subID, messageID, pos,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"DELETE FROM in_flight_messages WHERE subscription_id = ? AND message_id = ?",
		subID, messageID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Dispatcher) Heartbeat(consumerID string) error {
	d.store.Lock()
	defer d.store.Unlock()

	_, err := d.store.DB().Exec(
		"UPDATE consumers SET last_heartbeat = CURRENT_TIMESTAMP WHERE id = ?",
		consumerID,
	)
	return err
}
