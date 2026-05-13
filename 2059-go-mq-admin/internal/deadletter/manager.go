package deadletter

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"go-mq-admin/internal/models"
	"go-mq-admin/internal/storage"
	"log"
	"time"
)

type Manager struct {
	store *storage.Storage
}

func NewManager(store *storage.Storage) *Manager {
	return &Manager{store: store}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) MoveToDeadLetter(messageID, topicName, payload string, publishedAt time.Time, reason string) error {
	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()
	deadID := generateID()

	var pos int
	err := db.QueryRow(
		"SELECT COALESCE(MAX(position), 0) + 1 FROM dead_letters WHERE topic_name = ?",
		topicName,
	).Scan(&pos)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO dead_letters (id, topic_name, payload, original_message_id, published_at, dead_at, reason, position)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?)`,
		deadID, topicName, payload, messageID, publishedAt, reason, pos,
	)
	return err
}

func (m *Manager) List(topicName string) ([]*models.DeadLetter, error) {
	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()
	rows, err := db.Query(
		`SELECT id, topic_name, payload, original_message_id, published_at, dead_at, reason
		 FROM dead_letters WHERE topic_name = ? ORDER BY position ASC`,
		topicName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var letters []*models.DeadLetter
	for rows.Next() {
		var dl models.DeadLetter
		if err := rows.Scan(&dl.ID, &dl.TopicName, &dl.Payload, &dl.OriginalID, &dl.PublishedAt, &dl.DeadAt, &dl.Reason); err != nil {
			return nil, err
		}
		letters = append(letters, &dl)
	}
	return letters, nil
}

func (m *Manager) Get(topicName, deadID string) (*models.DeadLetter, error) {
	m.store.Lock()
	defer m.store.Unlock()

	var dl models.DeadLetter
	err := m.store.DB().QueryRow(
		`SELECT id, topic_name, payload, original_message_id, published_at, dead_at, reason
		 FROM dead_letters WHERE topic_name = ? AND id = ?`,
		topicName, deadID,
	).Scan(&dl.ID, &dl.TopicName, &dl.Payload, &dl.OriginalID, &dl.PublishedAt, &dl.DeadAt, &dl.Reason)
	if err != nil {
		return nil, err
	}
	return &dl, nil
}

func (m *Manager) Requeue(topicName, deadID string) error {
	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var dl models.DeadLetter
	err = tx.QueryRow(
		`SELECT id, topic_name, payload, original_message_id, published_at, dead_at, reason
		 FROM dead_letters WHERE topic_name = ? AND id = ?`,
		topicName, deadID,
	).Scan(&dl.ID, &dl.TopicName, &dl.Payload, &dl.OriginalID, &dl.PublishedAt, &dl.DeadAt, &dl.Reason)
	if err != nil {
		return err
	}

	subs, err := tx.Query(
		"SELECT id FROM subscriptions WHERE topic_name = ?",
		topicName,
	)
	if err != nil {
		return err
	}
	defer subs.Close()

	for subs.Next() {
		var subID int64
		if err := subs.Scan(&subID); err != nil {
			return err
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
			subID, dl.OriginalID, pos,
		)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM dead_letters WHERE id = ?", deadID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) Drop(topicName, deadID string) error {
	m.store.Lock()
	defer m.store.Unlock()

	_, err := m.store.DB().Exec(
		"DELETE FROM dead_letters WHERE topic_name = ? AND id = ?",
		topicName, deadID,
	)
	return err
}

func (m *Manager) ProcessExpiredAndOverflow() error {
	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	expiredRows, err := tx.Query(
		`SELECT m.id, m.topic_name, m.payload, m.published_at
		 FROM messages m
		 WHERE m.expire_at < CURRENT_TIMESTAMP
		 AND NOT EXISTS (SELECT 1 FROM dead_letters dl WHERE dl.original_message_id = m.id)`,
	)
	if err != nil {
		return err
	}

	for expiredRows.Next() {
		var msgID, topicName, payload string
		var publishedAt time.Time
		if err := expiredRows.Scan(&msgID, &topicName, &payload, &publishedAt); err != nil {
			expiredRows.Close()
			return err
		}

		var pos int
		err = tx.QueryRow(
			"SELECT COALESCE(MAX(position), 0) + 1 FROM dead_letters WHERE topic_name = ?",
			topicName,
		).Scan(&pos)
		if err != nil {
			expiredRows.Close()
			return err
		}

		deadID := generateID()
		_, err = tx.Exec(
			`INSERT INTO dead_letters (id, topic_name, payload, original_message_id, published_at, dead_at, reason, position)
			 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 'ttl_expired', ?)`,
			deadID, topicName, payload, msgID, publishedAt, pos,
		)
		if err != nil {
			expiredRows.Close()
			return err
		}

		_, err = tx.Exec("DELETE FROM in_flight_messages WHERE message_id = ?", msgID)
		if err != nil {
			expiredRows.Close()
			return err
		}
		_, err = tx.Exec("DELETE FROM queue_messages WHERE message_id = ?", msgID)
		if err != nil {
			expiredRows.Close()
			return err
		}
	}
	expiredRows.Close()

	topics, err := tx.Query("SELECT name, max_dead_messages FROM topics")
	if err != nil {
		return err
	}

	for topics.Next() {
		var topicName string
		var maxDead int
		if err := topics.Scan(&topicName, &maxDead); err != nil {
			topics.Close()
			return err
		}

		var count int
		err = tx.QueryRow("SELECT COUNT(*) FROM dead_letters WHERE topic_name = ?", topicName).Scan(&count)
		if err != nil {
			topics.Close()
			return err
		}

		if count > maxDead {
			excess := count - maxDead
			log.Printf("[DeadLetter] Topic %s has %d dead letters, exceeding max %d, dropping %d", topicName, count, maxDead, excess)

			rows, err := tx.Query(
				`SELECT id FROM dead_letters WHERE topic_name = ? ORDER BY position ASC LIMIT ?`,
				topicName, excess,
			)
			if err != nil {
				topics.Close()
				return err
			}

			for rows.Next() {
				var dlID string
				if err := rows.Scan(&dlID); err != nil {
					rows.Close()
					topics.Close()
					return err
				}

				_, err = tx.Exec(
					`INSERT INTO dead_letter_logs (topic_name, dead_letter_id, dropped_at, reason)
					 VALUES (?, ?, CURRENT_TIMESTAMP, 'max_dead_exceeded')`,
					topicName, dlID,
				)
				if err != nil {
					rows.Close()
					topics.Close()
					return err
				}

				_, err = tx.Exec("DELETE FROM dead_letters WHERE id = ?", dlID)
				if err != nil {
					rows.Close()
					topics.Close()
					return err
				}

				log.Printf("[DeadLetter] Dropped dead letter %s from topic %s due to max exceeded", dlID, topicName)
			}
			rows.Close()
		}
	}
	topics.Close()

	_, err = tx.Exec(
		`DELETE FROM messages 
		 WHERE expire_at < CURRENT_TIMESTAMP
		 AND NOT EXISTS (SELECT 1 FROM in_flight_messages WHERE message_id = messages.id)
		 AND NOT EXISTS (SELECT 1 FROM queue_messages WHERE message_id = messages.id)
		 AND NOT EXISTS (SELECT 1 FROM dead_letters WHERE original_message_id = messages.id)`,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) RequeueTimedOutMessages() error {
	m.store.Lock()
	defer m.store.Unlock()

	db := m.store.DB()

	rows, err := db.Query(
		`SELECT ifm.subscription_id, ifm.message_id
		 FROM in_flight_messages ifm
		 JOIN consumers c ON ifm.consumer_id = c.id
		 WHERE ifm.deadline_at < CURRENT_TIMESTAMP
		 OR c.last_heartbeat < datetime('now', '-30 seconds')`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var subID int64
		var msgID string
		if err := rows.Scan(&subID, &msgID); err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		var pos int
		err = tx.QueryRow(
			"SELECT COALESCE(MAX(position), 0) + 1 FROM queue_messages WHERE subscription_id = ?",
			subID,
		).Scan(&pos)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec(
			"INSERT INTO queue_messages (subscription_id, message_id, position) VALUES (?, ?, ?)",
			subID, msgID, pos,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec(
			"DELETE FROM in_flight_messages WHERE subscription_id = ? AND message_id = ?",
			subID, msgID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) CleanupInactiveConsumers() error {
	m.store.Lock()
	defer m.store.Unlock()

	_, err := m.store.DB().Exec(
		"DELETE FROM consumers WHERE last_heartbeat < datetime('now', '-5 minutes')",
	)
	return err
}

var _ = sql.ErrNoRows
