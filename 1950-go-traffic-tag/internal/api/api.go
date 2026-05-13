package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"traffictag/internal/matcher"
	"traffictag/internal/middleware"
	"traffictag/internal/tagstore"
)

type Handler struct {
	ruleStore *matcher.RuleStore
	tagStore  *tagstore.TagStore
}

func NewHandler(ruleStore *matcher.RuleStore, tagStore *tagstore.TagStore) *Handler {
	return &Handler{
		ruleStore: ruleStore,
		tagStore:  tagStore,
	}
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

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/rules", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.listRules(w, r)
		case http.MethodPost:
			h.createRule(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/rules/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/rules/")
		if id == "" {
			writeError(w, http.StatusNotFound, "rule id required")
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.getRule(w, r, id)
		case http.MethodPut:
			h.updateRule(w, r, id)
		case http.MethodDelete:
			h.deleteRule(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.listTags(w, r)
	})

	mux.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.testRequest(w, r)
	})
}

func (h *Handler) listRules(w http.ResponseWriter, _ *http.Request) {
	rules := h.ruleStore.List()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

func (h *Handler) getRule(w http.ResponseWriter, _ *http.Request, id string) {
	rule, exists := h.ruleStore.Get(id)
	if !exists {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) createRule(w http.ResponseWriter, r *http.Request) {
	var rule matcher.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	if rule.ID == "" {
		writeError(w, http.StatusBadRequest, "rule id is required")
		return
	}

	if len(rule.Conditions) == 0 {
		writeError(w, http.StatusBadRequest, "at least one condition is required")
		return
	}

	if len(rule.Tags) == 0 {
		writeError(w, http.StatusBadRequest, "at least one tag is required")
		return
	}

	if _, exists := h.ruleStore.Get(rule.ID); exists {
		writeError(w, http.StatusConflict, "rule already exists")
		return
	}

	h.ruleStore.Add(&rule)
	writeJSON(w, http.StatusCreated, rule)
}

func (h *Handler) updateRule(w http.ResponseWriter, r *http.Request, id string) {
	var rule matcher.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	if len(rule.Conditions) == 0 {
		writeError(w, http.StatusBadRequest, "at least one condition is required")
		return
	}

	if len(rule.Tags) == 0 {
		writeError(w, http.StatusBadRequest, "at least one tag is required")
		return
	}

	if !h.ruleStore.Update(id, &rule) {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}

	rule.ID = id
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) deleteRule(w http.ResponseWriter, _ *http.Request, id string) {
	if !h.ruleStore.Remove(id) {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listTags(w http.ResponseWriter, _ *http.Request) {
	tags := h.tagStore.List()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
	})
}

func (h *Handler) testRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	io.Copy(io.Discard, r.Body)

	tags := h.ruleStore.MatchAll(r)
	headerTags := middleware.ParseTagsFromHeader(r.Header.Get(middleware.HeaderTrafficTags))

	if len(headerTags) > 0 {
		if tags == nil {
			tags = headerTags
		} else {
			for k, v := range headerTags {
				tags[k] = v
			}
		}
	}

	if len(tags) > 0 {
		h.tagStore.Increment(tags)
		w.Header().Set(middleware.HeaderTrafficTags, middleware.TagsToHeader(tags))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tags":           tags,
		"matched":        len(tags) > 0,
		"header_value":   middleware.TagsToHeader(tags),
	})
}
