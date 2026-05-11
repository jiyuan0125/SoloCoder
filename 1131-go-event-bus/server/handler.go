package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"eventbus-demo/common"
	"eventbus-demo/eventbus"
)

func writeJSON(w http.ResponseWriter, code int, resp common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}

func handleCreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	var req common.CreateTopicReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "invalid request body"))
		return
	}
	if req.Topic == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "topic is required"))
		return
	}

	err := bus.CreateTopic(req.Topic)
	if err != nil {
		if errors.Is(err, eventbus.ErrTopicExist) {
			writeJSON(w, http.StatusConflict, common.NewErrorResponse(common.CodeTopicExist, ""))
			return
		}
		writeJSON(w, http.StatusInternalServerError, common.NewErrorResponse(common.CodeError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"topic": req.Topic}))
}

func handleDeleteTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	var req common.DeleteTopicReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "invalid request body"))
		return
	}
	if req.Topic == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "topic is required"))
		return
	}

	err := bus.DeleteTopic(req.Topic)
	if err != nil {
		if errors.Is(err, eventbus.ErrTopicNotExist) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeTopicNotExist, ""))
			return
		}
		writeJSON(w, http.StatusInternalServerError, common.NewErrorResponse(common.CodeError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"topic": req.Topic}))
}

func handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	var req common.SubscribeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "invalid request body"))
		return
	}
	if req.Topic == "" || req.SubID == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "topic and sub_id are required"))
		return
	}

	handler := eventbus.NopHandler()
	if req.Endpoint != "" {
		handler = eventbus.WebhookHandler(req.Endpoint, 10*time.Second)
	}

	sub := &eventbus.Subscriber{
		ID:       req.SubID,
		Priority: req.Priority,
		Handler:  handler,
		Endpoint: req.Endpoint,
	}

	err := bus.Subscribe(req.Topic, sub)
	if err != nil {
		if errors.Is(err, eventbus.ErrTopicNotExist) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeTopicNotExist, ""))
			return
		}
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"topic":    req.Topic,
		"sub_id":   req.SubID,
		"priority": req.Priority,
	}))
}

func handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	var req common.UnsubscribeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "invalid request body"))
		return
	}
	if req.Topic == "" || req.SubID == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "topic and sub_id are required"))
		return
	}

	err := bus.Unsubscribe(req.Topic, req.SubID)
	if err != nil {
		if errors.Is(err, eventbus.ErrTopicNotExist) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeTopicNotExist, ""))
			return
		}
		if errors.Is(err, eventbus.ErrSubNotExist) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeSubNotExist, ""))
			return
		}
		writeJSON(w, http.StatusInternalServerError, common.NewErrorResponse(common.CodeError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"topic":  req.Topic,
		"sub_id": req.SubID,
	}))
}

func handleListSubscribers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	topic := r.URL.Query().Get("topic")
	if topic == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "topic is required"))
		return
	}

	subs, err := bus.ListSubscribers(topic)
	if err != nil {
		if errors.Is(err, eventbus.ErrTopicNotExist) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeTopicNotExist, ""))
			return
		}
		writeJSON(w, http.StatusInternalServerError, common.NewErrorResponse(common.CodeError, err.Error()))
		return
	}

	result := make([]common.SubscriberInfo, 0, len(subs))
	for _, s := range subs {
		result = append(result, common.SubscriberInfo{
			ID:       s.ID,
			Priority: s.Priority,
			Endpoint: s.Endpoint,
			IsActive: s.IsActive(),
			LastSeen: s.LastSeen().Unix(),
		})
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"topic":       topic,
		"subscribers": result,
	}))
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	var req common.PublishReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "invalid request body"))
		return
	}
	if req.Topic == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "topic is required"))
		return
	}

	state, err := bus.Publish(req.Topic, req.Payload, req.Async)
	if err != nil {
		if errors.Is(err, eventbus.ErrTopicNotExist) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeTopicNotExist, ""))
			return
		}
		writeJSON(w, http.StatusInternalServerError, common.NewErrorResponse(common.CodeError, err.Error()))
		return
	}

	if !req.Async {
		state.Wait()
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.EventResult{
		EventID:     state.Event.ID,
		Status:      state.Status.String(),
		Total:       state.Total,
		Success:     state.Success,
		Failed:      state.Failed,
		Intercepted: state.Intercepted,
	}))
}

func handleEventStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse(common.CodeInvalidParam, "method not allowed"))
		return
	}

	eventID := r.URL.Query().Get("event_id")
	if eventID == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidParam, "event_id is required"))
		return
	}

	state, err := bus.GetEventState(eventID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.NewErrorResponse(common.CodeEventNotExist, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.EventResult{
		EventID:     state.Event.ID,
		Status:      state.Status.String(),
		Total:       state.Total,
		Success:     state.Success,
		Failed:      state.Failed,
		Intercepted: state.Intercepted,
	}))
}
