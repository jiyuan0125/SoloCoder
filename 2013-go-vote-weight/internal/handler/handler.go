package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"votingsystem/internal/model"
	"votingsystem/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func parseID(path string) (int64, error) {
	return strconv.ParseInt(path, 10, 64)
}

func (h *Handler) ListVotings(w http.ResponseWriter, r *http.Request) {
	votings, err := h.svc.ListVotings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, votings)
}

func (h *Handler) CreateVoting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic    string `json:"topic"`
		Deadline string `json:"deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Topic == "" {
		writeError(w, http.StatusBadRequest, "topic is required")
		return
	}
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid deadline format, use RFC3339")
		return
	}

	v, err := h.svc.CreateVoting(req.Topic, deadline)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (h *Handler) GetVoting(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := parseID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid voting id")
		return
	}
	v, err := h.svc.GetVoting(id)
	if err != nil {
		if errors.Is(err, service.ErrVotingNotFound) {
			writeError(w, http.StatusNotFound, "voting not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) ListVotes(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := parseID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid voting id")
		return
	}
	votes, err := h.svc.ListVotes(id)
	if err != nil {
		if errors.Is(err, service.ErrVotingNotFound) {
			writeError(w, http.StatusNotFound, "voting not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, votes)
}

func (h *Handler) ListProxies(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := parseID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid voting id")
		return
	}
	proxies, err := h.svc.ListProxies(id)
	if err != nil {
		if errors.Is(err, service.ErrVotingNotFound) {
			writeError(w, http.StatusNotFound, "voting not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, proxies)
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := parseID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid voting id")
		return
	}
	logs, err := h.svc.ListAuditLogs(id)
	if err != nil {
		if errors.Is(err, service.ErrVotingNotFound) {
			writeError(w, http.StatusNotFound, "voting not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) Delegate(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := parseID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid voting id")
		return
	}
	var req struct {
		DelegatorID int64 `json:"delegator_id"`
		TrusteeID   int64 `json:"trustee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Delegate(id, req.DelegatorID, req.TrusteeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrVotingNotFound):
			writeError(w, http.StatusNotFound, "voting not found")
		case errors.Is(err, service.ErrVotingEnded):
			writeError(w, http.StatusBadRequest, "表决已结束")
		case errors.Is(err, service.ErrOwnerNotFound):
			writeError(w, http.StatusNotFound, "owner not found")
		case errors.Is(err, service.ErrAlreadyVoted):
			writeError(w, http.StatusConflict, "已投票")
		case errors.Is(err, service.ErrDuplicateProxy):
			writeError(w, http.StatusConflict, "重复委托")
		case errors.Is(err, service.ErrCannotDelegate):
			writeError(w, http.StatusBadRequest, "不能委托给自己")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) Vote(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := parseID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid voting id")
		return
	}
	var req struct {
		OwnerID int64             `json:"owner_id"`
		Choice  model.VoteChoice `json:"choice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Choice != model.VoteChoiceYes && req.Choice != model.VoteChoiceNo {
		writeError(w, http.StatusBadRequest, "choice must be 'yes' or 'no'")
		return
	}

	v, err := h.svc.Vote(id, req.OwnerID, req.Choice)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrVotingNotFound):
			writeError(w, http.StatusNotFound, "voting not found")
		case errors.Is(err, service.ErrVotingEnded):
			writeError(w, http.StatusBadRequest, "表决已结束")
		case errors.Is(err, service.ErrOwnerNotFound):
			writeError(w, http.StatusNotFound, "owner not found")
		case errors.Is(err, service.ErrAlreadyDelegated):
			writeError(w, http.StatusConflict, "已委托给他人")
		case errors.Is(err, service.ErrAlreadyVoted):
			writeError(w, http.StatusConflict, "已投票")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (h *Handler) ListOwners(w http.ResponseWriter, r *http.Request) {
	owners, err := h.svc.ListOwners()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, owners)
}

func (h *Handler) CreateOwner(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Area int    `json:"area"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	o, err := h.svc.CreateOwner(req.Name, req.Area)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, o)
}
