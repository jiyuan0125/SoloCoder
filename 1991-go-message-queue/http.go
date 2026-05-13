package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

func NewHTTPServer(broker *Broker) *http.Server {
	mux := http.NewServeMux()

	hs := &HTTPServer{broker: broker}

	mux.HandleFunc("/topics", hs.handleTopics)
	mux.HandleFunc("/publish", hs.handlePublish)
	mux.HandleFunc("/consume", hs.handleConsume)
	mux.HandleFunc("/ack", hs.handleAck)
	mux.HandleFunc("/nack", hs.handleNack)
	mux.HandleFunc("/register", hs.handleRegister)
	mux.HandleFunc("/unregister", hs.handleUnregister)
	mux.HandleFunc("/heartbeat", hs.handleHeartbeat)
	mux.HandleFunc("/messages", hs.handleGetMessages)
	mux.HandleFunc("/stats", hs.handleStats)
	mux.HandleFunc("/dead-letter", hs.handleDeadLetter)
	mux.HandleFunc("/republish", hs.handleRepublish)

	return &http.Server{Handler: mux}
}

type HTTPServer struct {
	broker *Broker
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (hs *HTTPServer) handleTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hs.broker.mu.RLock()
	topics := make([]string, 0, len(hs.broker.topics))
	for name := range hs.broker.topics {
		topics = append(topics, name)
	}
	hs.broker.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{"topics": topics})
}

func (hs *HTTPServer) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	if topicName == "" {
		writeError(w, http.StatusBadRequest, "topic is required")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}
	defer r.Body.Close()

	msgID := hs.broker.Publish(topicName, body)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": msgID,
		"topic":      topicName,
	})
}

func (hs *HTTPServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	groupName := r.URL.Query().Get("group")
	consumerID := r.URL.Query().Get("consumer")
	maxStr := r.URL.Query().Get("max")

	if topicName == "" || groupName == "" || consumerID == "" {
		writeError(w, http.StatusBadRequest, "topic, group, and consumer are required")
		return
	}

	maxMessages := 1
	if maxStr != "" {
		if m, err := strconv.Atoi(maxStr); err == nil && m > 0 {
			maxMessages = m
		}
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]interface{}{"messages": []interface{}{}})
		return
	}

	group := topic.GetOrCreateGroup(groupName)

	if !group.HasConsumer(consumerID) {
		writeError(w, http.StatusForbidden, "consumer not registered")
		return
	}

	messages := group.Consume(topic, consumerID, maxMessages)

	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		result = append(result, map[string]interface{}{
			"id":         msg.ID,
			"topic":      msg.Topic,
			"body":       string(msg.Body),
			"timestamp":  msg.Timestamp,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"messages": result})
}

func (hs *HTTPServer) handleAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	groupName := r.URL.Query().Get("group")
	consumerID := r.URL.Query().Get("consumer")
	msgIDStr := r.URL.Query().Get("message_id")

	if topicName == "" || groupName == "" || consumerID == "" || msgIDStr == "" {
		writeError(w, http.StatusBadRequest, "topic, group, consumer, and message_id are required")
		return
	}

	msgID, err := strconv.ParseUint(msgIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	group, ok := topic.GetGroup(groupName)
	if !ok {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}

	if !group.Ack(topic, consumerID, msgID) {
		writeError(w, http.StatusNotFound, "message not found or not owned by consumer")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

func (hs *HTTPServer) handleNack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	groupName := r.URL.Query().Get("group")
	consumerID := r.URL.Query().Get("consumer")
	msgIDStr := r.URL.Query().Get("message_id")

	if topicName == "" || groupName == "" || consumerID == "" || msgIDStr == "" {
		writeError(w, http.StatusBadRequest, "topic, group, consumer, and message_id are required")
		return
	}

	msgID, err := strconv.ParseUint(msgIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	group, ok := topic.GetGroup(groupName)
	if !ok {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}

	if !group.Nack(topic, consumerID, msgID) {
		writeError(w, http.StatusNotFound, "message not found or not owned by consumer")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "nacknowledged"})
}

func (hs *HTTPServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	groupName := r.URL.Query().Get("group")
	consumerID := r.URL.Query().Get("consumer")

	if topicName == "" || groupName == "" || consumerID == "" {
		writeError(w, http.StatusBadRequest, "topic, group, and consumer are required")
		return
	}

	topic := hs.broker.GetOrCreateTopic(topicName)
	group := topic.GetOrCreateGroup(groupName)

	_, err := group.RegisterConsumer(consumerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "registered",
		"topic":    topicName,
		"group":    groupName,
		"consumer": consumerID,
	})
}

func (hs *HTTPServer) handleUnregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	groupName := r.URL.Query().Get("group")
	consumerID := r.URL.Query().Get("consumer")

	if topicName == "" || groupName == "" || consumerID == "" {
		writeError(w, http.StatusBadRequest, "topic, group, and consumer are required")
		return
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]string{"status": "unregistered"})
		return
	}

	group, ok := topic.GetGroup(groupName)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]string{"status": "unregistered"})
		return
	}

	group.UnregisterConsumer(consumerID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "unregistered"})
}

func (hs *HTTPServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	groupName := r.URL.Query().Get("group")
	consumerID := r.URL.Query().Get("consumer")

	if topicName == "" || groupName == "" || consumerID == "" {
		writeError(w, http.StatusBadRequest, "topic, group, and consumer are required")
		return
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	group, ok := topic.GetGroup(groupName)
	if !ok {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}

	if !group.Heartbeat(consumerID) {
		writeError(w, http.StatusNotFound, "consumer not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (hs *HTTPServer) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	startStr := r.URL.Query().Get("start_id")
	endStr := r.URL.Query().Get("end_id")

	if topicName == "" {
		writeError(w, http.StatusBadRequest, "topic is required")
		return
	}

	var startID uint64
	var endID uint64
	var err error

	if startStr != "" {
		startID, err = strconv.ParseUint(startStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_id")
			return
		}
	}

	if endStr != "" {
		endID, err = strconv.ParseUint(endStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_id")
			return
		}
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]interface{}{"messages": []interface{}{}})
		return
	}

	messages := topic.GetMessagesByIDRange(startID, endID)

	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		result = append(result, map[string]interface{}{
			"id":        msg.ID,
			"topic":     msg.Topic,
			"body":      string(msg.Body),
			"timestamp": msg.Timestamp,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"messages": result})
}

func (hs *HTTPServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")

	var stats []map[string]interface{}

	if topicName != "" {
		topic, ok := hs.broker.GetTopic(topicName)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]interface{}{"topics": []interface{}{}})
			return
		}
		stats = append(stats, topic.GetStats())
	} else {
		hs.broker.mu.RLock()
		topics := make([]*Topic, 0, len(hs.broker.topics))
		for _, t := range hs.broker.topics {
			topics = append(topics, t)
		}
		hs.broker.mu.RUnlock()

		stats = make([]map[string]interface{}, 0, len(topics))
		for _, t := range topics {
			stats = append(stats, t.GetStats())
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"topics": stats})
}

func (hs *HTTPServer) handleDeadLetter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")

	if topicName == "" {
		writeError(w, http.StatusBadRequest, "topic is required")
		return
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]interface{}{"messages": []interface{}{}})
		return
	}

	messages := topic.GetDeadLetter()

	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		result = append(result, map[string]interface{}{
			"id":        msg.ID,
			"topic":     msg.Topic,
			"body":      string(msg.Body),
			"timestamp": msg.Timestamp,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"messages": result})
}

func (hs *HTTPServer) handleRepublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topicName := r.URL.Query().Get("topic")
	msgIDStr := r.URL.Query().Get("message_id")

	if topicName == "" || msgIDStr == "" {
		writeError(w, http.StatusBadRequest, "topic and message_id are required")
		return
	}

	msgID, err := strconv.ParseUint(msgIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	topic, ok := hs.broker.GetTopic(topicName)
	if !ok {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	if !topic.RepublishDeadLetter(msgID) {
		writeError(w, http.StatusNotFound, "message not found in dead letter queue")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "republished",
		"message_id": msgID,
	})
}

func StartBackgroundTasks(broker *Broker, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				runMaintenance(broker)
			}
		}
	}()
}

func runMaintenance(broker *Broker) {
	broker.mu.RLock()
	topics := make([]*Topic, 0, len(broker.topics))
	for _, t := range broker.topics {
		topics = append(topics, t)
	}
	broker.mu.RUnlock()

	for _, topic := range topics {
		topic.mu.RLock()
		groups := make([]*GroupState, 0, len(topic.Groups))
		for _, g := range topic.Groups {
			groups = append(groups, g)
		}
		topic.mu.RUnlock()

		for _, group := range groups {
			group.CleanupExpiredConsumers()
			group.RequeueExpired(topic)
		}
	}
}
