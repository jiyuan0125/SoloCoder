package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Subscription struct {
	ID           string   `json:"id"`
	CallbackURL  string   `json:"callback_url"`
	EventTypes   []string `json:"event_types"`
	Filters      []Filter `json:"filters,omitempty"`
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
}

type DeadLetter struct {
	ID           string    `json:"id"`
	EventID      string    `json:"event_id"`
	EventType    string    `json:"event_type"`
	Payload      json.RawMessage `json:"payload"`
	SubscriptionID string  `json:"subscription_id"`
	CallbackURL  string    `json:"callback_url"`
	Attempts     int       `json:"attempts"`
	LastError    string    `json:"last_error"`
	Timestamp    time.Time `json:"timestamp"`
}

type EventStats struct {
	Published int `json:"published"`
}

type Stats struct {
	Events      map[string]EventStats   `json:"events"`
	Subscriptions []SubscriptionStats   `json:"subscriptions"`
}

type SubscriptionStats struct {
	ID           string `json:"id"`
	CallbackURL  string `json:"callback_url"`
	SuccessCount int    `json:"success_count"`
	FailureCount int    `json:"failure_count"`
}

type EventBus struct {
	subscriptions  map[string]*Subscription
	deadLetters    map[string]*DeadLetter
	eventStats     map[string]*EventStats
	eventQueue     chan Event
	retryBackoffs  []time.Duration
	mu             sync.RWMutex
	httpClient     *http.Client
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscriptions: make(map[string]*Subscription),
		deadLetters:   make(map[string]*DeadLetter),
		eventStats:    make(map[string]*EventStats),
		eventQueue:    make(chan Event, 1000),
		retryBackoffs: []time.Duration{1 * time.Second, 5 * time.Second, 15 * time.Second},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (eb *EventBus) Start() {
	for i := 0; i < 10; i++ {
		go eb.worker()
	}
}

func (eb *EventBus) worker() {
	for event := range eb.eventQueue {
		eb.processEvent(event)
	}
}

func (eb *EventBus) processEvent(event Event) {
	eb.mu.RLock()
	subs := make([]*Subscription, 0, len(eb.subscriptions))
	for _, sub := range eb.subscriptions {
		subs = append(subs, sub)
	}
	eb.mu.RUnlock()

	for _, sub := range subs {
		if eb.matchesEventType(sub, event.Type) && eb.matchesFilters(sub, event.Payload) {
			go eb.deliverEvent(event, sub, 0)
		}
	}
}

func (eb *EventBus) matchesEventType(sub *Subscription, eventType string) bool {
	for _, pattern := range sub.EventTypes {
		if patternMatch(pattern, eventType) {
			return true
		}
	}
	return false
}

func patternMatch(pattern, str string) bool {
	rePattern := "^" + regexp.QuoteMeta(pattern)
	rePattern = regexp.MustCompile(`\\\*`).ReplaceAllString(rePattern, ".*")
	rePattern += "$"
	re, err := regexp.Compile(rePattern)
	if err != nil {
		return false
	}
	return re.MatchString(str)
}

func (eb *EventBus) matchesFilters(sub *Subscription, payload json.RawMessage) bool {
	if len(sub.Filters) == 0 {
		return true
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return false
	}

	for _, filter := range sub.Filters {
		if !applyFilter(filter, data) {
			return false
		}
	}
	return true
}

func applyFilter(filter Filter, data map[string]interface{}) bool {
	val, exists := data[filter.Field]
	if !exists {
		return false
	}

	switch filter.Operator {
	case "==", "=":
		return compareValues(val, filter.Value) == 0
	case "!=":
		return compareValues(val, filter.Value) != 0
	case ">":
		return compareValues(val, filter.Value) > 0
	case ">=":
		return compareValues(val, filter.Value) >= 0
	case "<":
		return compareValues(val, filter.Value) < 0
	case "<=":
		return compareValues(val, filter.Value) <= 0
	case "contains":
		str, ok := val.(string)
		if !ok {
			return false
		}
		return bytes.Contains([]byte(str), []byte(filter.Value))
	default:
		return false
	}
}

func compareValues(a interface{}, bStr string) int {
	switch aVal := a.(type) {
	case float64:
		bVal, err := strconv.ParseFloat(bStr, 64)
		if err != nil {
			return 0
		}
		if aVal > bVal {
			return 1
		} else if aVal < bVal {
			return -1
		}
		return 0
	case string:
		if aVal > bStr {
			return 1
		} else if aVal < bStr {
			return -1
		}
		return 0
	case int:
		bVal, err := strconv.Atoi(bStr)
		if err != nil {
			return 0
		}
		if aVal > bVal {
			return 1
		} else if aVal < bVal {
			return -1
		}
		return 0
	default:
		return 0
	}
}

func (eb *EventBus) deliverEvent(event Event, sub *Subscription, attempt int) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"event_id":  event.ID,
		"type":      event.Type,
		"payload":   event.Payload,
		"timestamp": event.Timestamp,
	})

	req, err := http.NewRequest("POST", sub.CallbackURL, bytes.NewBuffer(reqBody))
	if err != nil {
		eb.handleDeliveryFailure(event, sub, attempt, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := eb.httpClient.Do(req)
	if err != nil {
		eb.handleDeliveryFailure(event, sub, attempt, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		eb.handleDeliveryFailure(event, sub, attempt, fmt.Sprintf("status code %d", resp.StatusCode))
		return
	}

	eb.mu.Lock()
	sub.SuccessCount++
	eb.mu.Unlock()
}

func (eb *EventBus) handleDeliveryFailure(event Event, sub *Subscription, attempt int, errMsg string) {
	eb.mu.Lock()
	sub.FailureCount++
	eb.mu.Unlock()

	if attempt >= len(eb.retryBackoffs) {
		eb.addToDeadLetter(event, sub, attempt+1, errMsg)
		return
	}

	time.Sleep(eb.retryBackoffs[attempt])
	eb.deliverEvent(event, sub, attempt+1)
}

func (eb *EventBus) addToDeadLetter(event Event, sub *Subscription, attempts int, errMsg string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	deadLetter := &DeadLetter{
		ID:             uuid.New().String(),
		EventID:        event.ID,
		EventType:      event.Type,
		Payload:        event.Payload,
		SubscriptionID: sub.ID,
		CallbackURL:    sub.CallbackURL,
		Attempts:       attempts,
		LastError:      errMsg,
		Timestamp:      time.Now(),
	}

	eb.deadLetters[deadLetter.ID] = deadLetter
}

func (eb *EventBus) Publish(eventType string, payload json.RawMessage) string {
	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	eb.mu.Lock()
	if eb.eventStats[eventType] == nil {
		eb.eventStats[eventType] = &EventStats{}
	}
	eb.eventStats[eventType].Published++
	eb.mu.Unlock()

	eb.eventQueue <- event

	return event.ID
}

func (eb *EventBus) Subscribe(callbackURL string, eventTypes []string, filters []Filter) *Subscription {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub := &Subscription{
		ID:          uuid.New().String(),
		CallbackURL: callbackURL,
		EventTypes:  eventTypes,
		Filters:     filters,
	}

	eb.subscriptions[sub.ID] = sub
	return sub
}

func (eb *EventBus) ListSubscriptions() []*Subscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs := make([]*Subscription, 0, len(eb.subscriptions))
	for _, sub := range eb.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

func (eb *EventBus) ListDeadLetters() []*DeadLetter {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	dls := make([]*DeadLetter, 0, len(eb.deadLetters))
	for _, dl := range eb.deadLetters {
		dls = append(dls, dl)
	}
	return dls
}

func (eb *EventBus) GetDeadLetter(id string) (*DeadLetter, bool) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	dl, exists := eb.deadLetters[id]
	return dl, exists
}

func (eb *EventBus) RetryDeadLetter(id string) bool {
	eb.mu.RLock()
	dl, exists := eb.deadLetters[id]
	if !exists {
		eb.mu.RUnlock()
		return false
	}

	event := Event{
		ID:        dl.EventID,
		Type:      dl.EventType,
		Payload:   dl.Payload,
		Timestamp: time.Now(),
	}
	eb.mu.RUnlock()

	eb.mu.Lock()
	delete(eb.deadLetters, id)
	eb.mu.Unlock()

	eb.processEvent(event)
	return true
}

func (eb *EventBus) GetStats() Stats {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	stats := Stats{
		Events: make(map[string]EventStats),
	}

	for eventType, eventStat := range eb.eventStats {
		stats.Events[eventType] = EventStats{
			Published: eventStat.Published,
		}
	}

	for _, sub := range eb.subscriptions {
		stats.Subscriptions = append(stats.Subscriptions, SubscriptionStats{
			ID:           sub.ID,
			CallbackURL:  sub.CallbackURL,
			SuccessCount: sub.SuccessCount,
			FailureCount: sub.FailureCount,
		})
	}

	return stats
}

func main() {
	eb := NewEventBus()
	eb.Start()

	app := fiber.New()

	app.Post("/events/publish", func(c *fiber.Ctx) error {
		var body struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if body.Type == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Event type is required",
			})
		}

		eventID := eb.Publish(body.Type, body.Payload)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"event_id": eventID,
			"status":   "accepted",
		})
	})

	app.Post("/subscriptions", func(c *fiber.Ctx) error {
		var body struct {
			CallbackURL string   `json:"callback_url"`
			EventTypes  []string `json:"event_types"`
			Filters     []Filter `json:"filters,omitempty"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if body.CallbackURL == "" || len(body.EventTypes) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "callback_url and event_types are required",
			})
		}

		sub := eb.Subscribe(body.CallbackURL, body.EventTypes, body.Filters)

		return c.Status(fiber.StatusCreated).JSON(sub)
	})

	app.Get("/subscriptions", func(c *fiber.Ctx) error {
		subs := eb.ListSubscriptions()
		return c.JSON(subs)
	})

	app.Get("/events/dead-letter", func(c *fiber.Ctx) error {
		dls := eb.ListDeadLetters()
		return c.JSON(dls)
	})

	app.Post("/events/dead-letter/:id/retry", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if eb.RetryDeadLetter(id) {
			return c.JSON(fiber.Map{
				"status": "retrying",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Dead letter not found",
		})
	})

	app.Get("/events/stats", func(c *fiber.Ctx) error {
		stats := eb.GetStats()
		return c.JSON(stats)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Event bus running on port %s\n", port)
	app.Listen(":" + port)
}
