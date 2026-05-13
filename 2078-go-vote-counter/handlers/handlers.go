package handlers

import (
	"encoding/json"
	"encoding/csv"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"vote-counter/db"
)

type CreatePollRequest struct {
	Title         string   `json:"title"`
	IsSingleChoice bool     `json:"is_single_choice"`
	MaxChoices    int      `json:"max_choices"`
	IsAnonymous   bool     `json:"is_anonymous"`
	Deadline      string   `json:"deadline"`
	Options       []string `json:"options"`
}

type UpdateVoteRequest struct {
	ParticipantID string  `json:"participant_id"`
	OptionIDs     []int64 `json:"option_ids"`
}

type OptionResponse struct {
	ID       int64   `json:"id"`
	Text     string  `json:"text"`
	Votes    int     `json:"votes"`
	Percent  float64 `json:"percent"`
}

type PollResponse struct {
	ID            int64            `json:"id"`
	Title         string           `json:"title"`
	IsSingleChoice bool            `json:"is_single_choice"`
	MaxChoices    int              `json:"max_choices"`
	IsAnonymous   bool             `json:"is_anonymous"`
	Deadline      time.Time        `json:"deadline"`
	Status        string           `json:"status"`
	TotalVotes    int              `json:"total_votes"`
	Options       []OptionResponse `json:"options"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondJSON(w, status, errorResponse{Error: message})
}

func parsePollID(path string) (int64, string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/polls/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", false
	}

	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	return id, subPath, true
}

func PollHandler(w http.ResponseWriter, r *http.Request) {
	pollID, subPath, ok := parsePollID(r.URL.Path)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid poll ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if subPath == "export" {
			ExportPoll(w, r, pollID)
		} else {
			GetPoll(w, r, pollID)
		}
	case http.MethodPost, http.MethodPut:
		if subPath == "vote" {
			UpdateVote(w, r, pollID)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func CreatePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req CreatePollRequest
	if err := json.Unmarshal(body, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if len(req.Options) < 2 {
		respondError(w, http.StatusBadRequest, "At least 2 options are required")
		return
	}

	if req.IsSingleChoice && len(req.Options) < 1 {
		respondError(w, http.StatusBadRequest, "Single choice poll requires at least 1 option")
		return
	}

	if !req.IsSingleChoice && req.MaxChoices <= 0 {
		respondError(w, http.StatusBadRequest, "Max choices must be greater than 0 for multiple choice poll")
		return
	}

	if !req.IsSingleChoice && req.MaxChoices > len(req.Options) {
		respondError(w, http.StatusBadRequest, "Max choices cannot exceed number of options")
		return
	}

	if req.IsSingleChoice {
		req.MaxChoices = 1
	}

	var deadline time.Time
	if req.Deadline == "" {
		respondError(w, http.StatusBadRequest, "Deadline is required")
		return
	}

	deadline, err = time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid deadline format, expected RFC3339")
		return
	}

	if deadline.Before(time.Now()) {
		respondError(w, http.StatusBadRequest, "Deadline must be in the future")
		return
	}

	poll := &db.Poll{
		Title:         req.Title,
		IsSingleChoice: req.IsSingleChoice,
		MaxChoices:    req.MaxChoices,
		IsAnonymous:   req.IsAnonymous,
		Deadline:      deadline,
	}

	createdPoll, createdOptions, err := db.CreatePoll(poll, req.Options)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create poll: "+err.Error())
		return
	}

	optionsResp := make([]OptionResponse, len(createdOptions))
	for i, opt := range createdOptions {
		optionsResp[i] = OptionResponse{
			ID:      opt.ID,
			Text:    opt.Text,
			Votes:   0,
			Percent: 0,
		}
	}

	resp := PollResponse{
		ID:             createdPoll.ID,
		Title:          createdPoll.Title,
		IsSingleChoice: createdPoll.IsSingleChoice,
		MaxChoices:     createdPoll.MaxChoices,
		IsAnonymous:    createdPoll.IsAnonymous,
		Deadline:       createdPoll.Deadline,
		Status:         "active",
		TotalVotes:     0,
		Options:        optionsResp,
	}

	respondJSON(w, http.StatusCreated, resp)
}

func GetPoll(w http.ResponseWriter, r *http.Request, pollID int64) {
	poll, err := db.GetPollByID(pollID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get poll: "+err.Error())
		return
	}
	if poll == nil {
		respondError(w, http.StatusNotFound, "Poll not found")
		return
	}

	options, err := db.GetOptionsByPollID(pollID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get options: "+err.Error())
		return
	}

	totalVotes := 0
	for _, opt := range options {
		totalVotes += opt.Votes
	}

	status := "active"
	if time.Now().After(poll.Deadline) {
		status = "ended"
	}

	optionsResp := make([]OptionResponse, len(options))
	for i, opt := range options {
		percent := 0.0
		if totalVotes > 0 {
			percent = float64(opt.Votes) / float64(totalVotes) * 100
		}
		optionsResp[i] = OptionResponse{
			ID:      opt.ID,
			Text:    opt.Text,
			Votes:   opt.Votes,
			Percent: percent,
		}
	}

	resp := PollResponse{
		ID:             poll.ID,
		Title:          poll.Title,
		IsSingleChoice: poll.IsSingleChoice,
		MaxChoices:     poll.MaxChoices,
		IsAnonymous:    poll.IsAnonymous,
		Deadline:       poll.Deadline,
		Status:         status,
		TotalVotes:     totalVotes,
		Options:        optionsResp,
	}

	respondJSON(w, http.StatusOK, resp)
}

func UpdateVote(w http.ResponseWriter, r *http.Request, pollID int64) {
	db.Lock()
	defer db.Unlock()

	poll, err := db.GetPollByID(pollID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get poll: "+err.Error())
		return
	}
	if poll == nil {
		respondError(w, http.StatusNotFound, "Poll not found")
		return
	}

	if time.Now().After(poll.Deadline) {
		respondError(w, http.StatusForbidden, "Poll has ended")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req UpdateVoteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.ParticipantID == "" {
		respondError(w, http.StatusBadRequest, "Participant ID is required")
		return
	}

	if len(req.OptionIDs) == 0 {
		respondError(w, http.StatusBadRequest, "At least one option is required")
		return
	}

	if poll.IsSingleChoice && len(req.OptionIDs) > 1 {
		respondError(w, http.StatusBadRequest, "Single choice poll allows only 1 option")
		return
	}

	if !poll.IsSingleChoice && len(req.OptionIDs) > poll.MaxChoices {
		respondError(w, http.StatusBadRequest, "Exceeded maximum allowed choices: "+strconv.Itoa(poll.MaxChoices))
		return
	}

	options, err := db.GetOptionsByPollID(pollID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get options: "+err.Error())
		return
	}

	validOptionIDs := make(map[int64]bool)
	for _, opt := range options {
		validOptionIDs[opt.ID] = true
	}

	for _, optID := range req.OptionIDs {
		if !validOptionIDs[optID] {
			respondError(w, http.StatusBadRequest, "Invalid option ID: "+strconv.FormatInt(optID, 10))
			return
		}
	}

	seen := make(map[int64]bool)
	for _, optID := range req.OptionIDs {
		if seen[optID] {
			respondError(w, http.StatusBadRequest, "Duplicate option ID: "+strconv.FormatInt(optID, 10))
			return
		}
		seen[optID] = true
	}

	existingVote, err := db.GetVoteByParticipant(pollID, req.ParticipantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check existing vote: "+err.Error())
		return
	}

	var previousOptionIDs []int64
	if existingVote != nil {
		previousOptionIDs = db.OptionIDsFromString(existingVote.OptionIDs)
	}

	optionIDSet := make(map[int64]bool)
	for _, id := range req.OptionIDs {
		optionIDSet[id] = true
	}

	optionsToSubtract := make([]int64, 0)
	for _, id := range previousOptionIDs {
		if !optionIDSet[id] {
			optionsToSubtract = append(optionsToSubtract, id)
		}
	}

	optionsToAdd := make([]int64, 0)
	previousSet := make(map[int64]bool)
	for _, id := range previousOptionIDs {
		previousSet[id] = true
	}
	for _, id := range req.OptionIDs {
		if !previousSet[id] {
			optionsToAdd = append(optionsToAdd, id)
		}
	}

	if len(optionsToSubtract) > 0 || len(optionsToAdd) > 0 {
		if err := db.UpdateOptionsVotes(pollID, optionsToSubtract, optionsToAdd); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update votes: "+err.Error())
			return
		}
	}

	if existingVote == nil {
		if err := db.CreateVote(pollID, req.ParticipantID, req.OptionIDs); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create vote: "+err.Error())
			return
		}
	} else {
		if err := db.UpdateVote(existingVote.ID, req.OptionIDs); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update vote: "+err.Error())
			return
		}
	}

	GetPoll(w, r, pollID)
}

func ExportPoll(w http.ResponseWriter, r *http.Request, pollID int64) {
	poll, err := db.GetPollByID(pollID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get poll: "+err.Error())
		return
	}
	if poll == nil {
		respondError(w, http.StatusNotFound, "Poll not found")
		return
	}

	options, err := db.GetOptionsByPollID(pollID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get options: "+err.Error())
		return
	}

	totalVotes := 0
	for _, opt := range options {
		totalVotes += opt.Votes
	}

	status := "进行中"
	if time.Now().After(poll.Deadline) {
		status = "已结束"
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=poll_"+strconv.FormatInt(pollID, 10)+".csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"投票标题", poll.Title, "", ""})
	writer.Write([]string{"状态", status, "", ""})
	writer.Write([]string{"截止时间", poll.Deadline.Format("2006-01-02 15:04:05"), "", ""})
	pollType := "单选"
	if !poll.IsSingleChoice {
		pollType = "多选(最多" + strconv.Itoa(poll.MaxChoices) + "项)"
	}
	writer.Write([]string{"投票类型", pollType, "", ""})
	writer.Write([]string{"总票数", strconv.Itoa(totalVotes), "", ""})
	writer.Write([]string{"", "", "", ""})
	writer.Write([]string{"选项ID", "选项内容", "票数", "占比"})

	for _, opt := range options {
		percent := "0.00%"
		if totalVotes > 0 {
			percentVal := float64(opt.Votes) / float64(totalVotes) * 100
			percent = strconv.FormatFloat(percentVal, 'f', 2, 64) + "%"
		}
		writer.Write([]string{
			strconv.FormatInt(opt.ID, 10),
			opt.Text,
			strconv.Itoa(opt.Votes),
			percent,
		})
	}
}
