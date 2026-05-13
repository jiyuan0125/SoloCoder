package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"go-mq-admin/internal/deadletter"
	"go-mq-admin/internal/dispatcher"
	"go-mq-admin/internal/models"
	"go-mq-admin/internal/storage"
	"go-mq-admin/internal/subscription"
	"go-mq-admin/internal/topic"
	"net/http"
	"os"
	"strings"
)

type Handler struct {
	topicMgr      *topic.Manager
	subMgr        *subscription.Manager
	dispatcher    *dispatcher.Dispatcher
	deadLetterMgr *deadletter.Manager
	store         *storage.Storage
}

func NewHandler(
	topicMgr *topic.Manager,
	subMgr *subscription.Manager,
	disp *dispatcher.Dispatcher,
	dlMgr *deadletter.Manager,
	store *storage.Storage,
) *Handler {
	return &Handler{
		topicMgr:      topicMgr,
		subMgr:        subMgr,
		dispatcher:    disp,
		deadLetterMgr: dlMgr,
		store:         store,
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func (h *Handler) ListTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.topicMgr.List()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, topics)
}

func (h *Handler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		TTL             int64  `json:"ttl"`
		MaxDeadMessages int    `json:"max_dead_messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.topicMgr.Create(req.Name, req.TTL, req.MaxDeadMessages)
	if err != nil {
		if errors.Is(err, topic.ErrInvalidTopicName) {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, topic.ErrTopicExists) {
			errorResponse(w, http.StatusConflict, err.Error())
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, t)
}

func (h *Handler) GetTopic(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("topic")
	if name == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	t, err := h.topicMgr.Get(name)
	if err != nil {
		if errors.Is(err, topic.ErrTopicNotFound) || errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "topic not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, t)
}

func (h *Handler) GetTopicStats(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("topic")
	if name == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	stats, err := h.topicMgr.GetStats(name)
	if err != nil {
		if errors.Is(err, topic.ErrTopicNotFound) {
			errorResponse(w, http.StatusNotFound, "topic not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, stats)
}

func (h *Handler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("topic")
	if name == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	if err := h.topicMgr.Delete(name); err != nil {
		if errors.Is(err, topic.ErrTopicNotFound) {
			errorResponse(w, http.StatusNotFound, "topic not found")
			return
		}
		if errors.Is(err, topic.ErrActiveSubscriptions) {
			errorResponse(w, http.StatusConflict, err.Error())
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	if topicName == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	exists, err := h.store.TopicExists(topicName)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		errorResponse(w, http.StatusNotFound, "topic not found")
		return
	}

	subs, err := h.subMgr.List(topicName)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, subs)
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	if topicName == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	exists, err := h.store.TopicExists(topicName)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		errorResponse(w, http.StatusNotFound, "topic not found")
		return
	}

	var req struct {
		Name       string `json:"name"`
		Mode       string `json:"mode"`
		AckTimeout int64  `json:"ack_timeout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.subMgr.Create(topicName, req.Name, req.Mode, req.AckTimeout)
	if err != nil {
		if errors.Is(err, subscription.ErrInvalidMode) {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, subscription.ErrSubscriptionExists) {
			errorResponse(w, http.StatusConflict, err.Error())
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, sub)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	subName := r.PathValue("subscription")

	sub, err := h.subMgr.Get(topicName, subName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, sub)
}

func (h *Handler) GetSubscriptionStats(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	subName := r.PathValue("subscription")

	stats, err := h.subMgr.GetStats(topicName, subName)
	if err != nil {
		if errors.Is(err, subscription.ErrSubscriptionNotFound) || errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, stats)
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	subName := r.PathValue("subscription")

	if err := h.subMgr.Delete(topicName, subName); err != nil {
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PublishMessage(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	if topicName == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	var req struct {
		Messages []string `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ids, err := h.dispatcher.Publish(topicName, req.Messages)
	if err != nil {
		if errors.Is(err, dispatcher.ErrEmptyPayload) {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, dispatcher.ErrBatchTooLarge) {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "topic not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message_ids": ids,
		"count":       len(ids),
	})
}

func (h *Handler) PullMessages(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	subName := r.PathValue("subscription")
	consumerID := r.URL.Query().Get("consumer_id")

	max := 10
	maxStr := r.URL.Query().Get("max")
	if maxStr != "" {
		var m int
		temp := struct{ Max *int }{Max: &m}
		if err := json.NewDecoder(strings.NewReader(`{"max":`+maxStr+`}`)).Decode(&temp); err == nil && m > 0 {
			max = m
		}
	}

	messages, err := h.dispatcher.Pull(topicName, subName, consumerID, max)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, messages)
}

func (h *Handler) AckMessage(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	subName := r.PathValue("subscription")
	messageID := r.PathValue("message_id")

	if err := h.dispatcher.Ack(topicName, subName, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) NackMessage(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	subName := r.PathValue("subscription")
	messageID := r.PathValue("message_id")

	if err := h.dispatcher.Nack(topicName, subName, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	consumerID := r.PathValue("consumer_id")
	if consumerID == "" {
		errorResponse(w, http.StatusBadRequest, "consumer_id is required")
		return
	}

	if err := h.dispatcher.Heartbeat(consumerID); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListDeadLetters(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	if topicName == "" {
		errorResponse(w, http.StatusBadRequest, "topic name is required")
		return
	}

	exists, err := h.store.TopicExists(topicName)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		errorResponse(w, http.StatusNotFound, "topic not found")
		return
	}

	letters, err := h.deadLetterMgr.List(topicName)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, letters)
}

func (h *Handler) GetDeadLetter(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	deadID := r.PathValue("dead_letter_id")

	dl, err := h.deadLetterMgr.Get(topicName, deadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "dead letter not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, dl)
}

func (h *Handler) RequeueDeadLetter(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	deadID := r.PathValue("dead_letter_id")

	if err := h.deadLetterMgr.Requeue(topicName, deadID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "dead letter not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DropDeadLetter(w http.ResponseWriter, r *http.Request) {
	topicName := r.PathValue("topic")
	deadID := r.PathValue("dead_letter_id")

	if err := h.deadLetterMgr.Drop(topicName, deadID); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	topics, err := h.topicMgr.List()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var statsList []*models.TopicStats
	for _, t := range topics {
		stats, err := h.topicMgr.GetStats(t.Name)
		if err == nil {
			statsList = append(statsList, stats)
		}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"topics":       topics,
		"topic_stats":  statsList,
		"total_topics": len(topics),
	})
}

func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("web/ui.html")
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to load UI: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
