package models

import "time"

type Topic struct {
	Name            string    `json:"name"`
	TTL             int64     `json:"ttl"`
	MaxDeadMessages int       `json:"max_dead_messages"`
	CreatedAt       time.Time `json:"created_at"`
}

type Subscription struct {
	Name        string    `json:"name"`
	TopicName   string    `json:"topic_name"`
	Mode        string    `json:"mode"`
	AckTimeout  int64     `json:"ack_timeout"`
	CreatedAt   time.Time `json:"created_at"`
}

type Message struct {
	ID          string    `json:"id"`
	TopicName   string    `json:"topic_name"`
	Payload     string    `json:"payload"`
	PublishedAt time.Time `json:"published_at"`
	ExpireAt    time.Time `json:"expire_at"`
}

type InFlightMessage struct {
	ID             string    `json:"id"`
	SubscriptionID int64     `json:"subscription_id"`
	MessageID      string    `json:"message_id"`
	ConsumerID     string    `json:"consumer_id"`
	DeliveredAt    time.Time `json:"delivered_at"`
	DeadlineAt     time.Time `json:"deadline_at"`
	DeliveryCount  int       `json:"delivery_count"`
}

type DeadLetter struct {
	ID          string    `json:"id"`
	TopicName   string    `json:"topic_name"`
	Payload     string    `json:"payload"`
	OriginalID  string    `json:"original_id"`
	PublishedAt time.Time `json:"published_at"`
	DeadAt      time.Time `json:"dead_at"`
	Reason      string    `json:"reason"`
}

type Consumer struct {
	ID             string    `json:"id"`
	SubscriptionID int64     `json:"subscription_id"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
}

type TopicStats struct {
	TopicName         string `json:"topic_name"`
	TotalMessages     int64  `json:"total_messages"`
	QueuedMessages    int64  `json:"queued_messages"`
	InFlightMessages  int64  `json:"in_flight_messages"`
	DeadLetterCount   int64  `json:"dead_letter_count"`
	SubscriptionCount int64  `json:"subscription_count"`
}

type SubscriptionStats struct {
	SubscriptionName  string `json:"subscription_name"`
	TopicName         string `json:"topic_name"`
	Mode              string `json:"mode"`
	QueuedMessages    int64  `json:"queued_messages"`
	InFlightMessages  int64  `json:"in_flight_messages"`
	ActiveConsumers   int64  `json:"active_consumers"`
}

type SubscriptionMessage struct {
	MessageID   string `json:"message_id"`
	Payload     string `json:"payload"`
	PublishedAt string `json:"published_at"`
}
