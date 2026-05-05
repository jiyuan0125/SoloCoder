package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tag-manager/common"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) logAudit(operator, action, target, details string) {
	log := &common.AuditLog{
		ID:        generateID("audit"),
		Operator:  operator,
		Action:    action,
		Target:    target,
		Details:   details,
		Timestamp: nowUnix(),
	}
	h.store.AddAuditLog(log)
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.CreateTagRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := validateCreateTagRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if existing := h.store.GetTagByName(req.Name); existing != nil {
		writeError(w, http.StatusConflict, &DuplicateNameError{Name: req.Name})
		return
	}

	tag := &common.Tag{
		ID:          generateID("tag"),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		GroupID:     req.GroupID,
		CreatedAt:   nowUnix(),
		UpdatedAt:   nowUnix(),
		EnumConfig:  req.EnumConfig,
		NumberConfig: req.NumberConfig,
		TextConfig:  req.TextConfig,
	}

	if err := h.store.CreateTag(tag); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	details, _ := json.Marshal(map[string]interface{}{
		"name":        req.Name,
		"type":        req.Type,
		"description": req.Description,
	})
	h.logAudit(req.Operator, "CREATE", "tag:"+tag.ID, string(details))

	writeSuccess(w, common.CreateTagResponse{TagID: tag.ID})
}

func (h *Handler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	existing := h.store.GetTagByID(tagID)
	if existing == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	var req common.UpdateTagRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Operator == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "operator is required"})
		return
	}

	oldName := existing.Name

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.GroupID != "" {
		existing.GroupID = req.GroupID
	}
	if req.EnumConfig != nil {
		existing.EnumConfig = req.EnumConfig
	}
	if req.NumberConfig != nil {
		existing.NumberConfig = req.NumberConfig
	}
	if req.TextConfig != nil {
		existing.TextConfig = req.TextConfig
	}
	existing.UpdatedAt = nowUnix()

	if err := h.store.UpdateTag(existing, oldName); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	details, _ := json.Marshal(map[string]interface{}{
		"old_name":   oldName,
		"new_name":   existing.Name,
		"group_id":   existing.GroupID,
	})
	h.logAudit(req.Operator, "UPDATE", "tag:"+tagID, string(details))

	writeSuccess(w, nil)
}

func (h *Handler) GetTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tag := h.store.GetTagByID(tagID)
	if tag == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	writeSuccess(w, tag)
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	operator := r.URL.Query().Get("operator")
	if operator == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "operator is required"})
		return
	}

	tag := h.store.GetTagByID(tagID)
	if tag == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	impactedCount := h.store.DeleteAllTagRelations(tagID)

	h.store.DeleteTag(tagID)

	details, _ := json.Marshal(map[string]interface{}{
		"tag_name":         tag.Name,
		"impacted_users":   impactedCount,
	})
	h.logAudit(operator, "DELETE", "tag:"+tagID, string(details))

	writeSuccess(w, map[string]interface{}{
		"impacted_users": impactedCount,
	})
}

func (h *Handler) GetTagImpact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tag := h.store.GetTagByID(tagID)
	if tag == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	count := h.store.GetTagImpactCount(tagID)

	writeSuccess(w, common.GetTagImpactResponse{
		ImpactedUserCount: count,
	})
}

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	groupID := r.URL.Query().Get("group_id")

	var tags []*common.Tag
	if groupID != "" {
		tags = h.store.ListTagsByGroup(groupID)
	} else {
		tags = h.store.ListTags()
	}

	result := make([]common.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, *t)
	}

	writeSuccess(w, common.ListTagsResponse{Tags: result})
}

func (h *Handler) CreateTagGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.CreateTagGroupRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Operator == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "operator is required"})
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "group name is required"})
		return
	}

	group := &common.TagGroup{
		ID:          generateID("group"),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   nowUnix(),
	}

	h.store.CreateTagGroup(group)

	details, _ := json.Marshal(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	})
	h.logAudit(req.Operator, "CREATE", "group:"+group.ID, string(details))

	writeSuccess(w, common.CreateTagGroupResponse{GroupID: group.ID})
}

func (h *Handler) GetTagGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	groupID, err := getIDFromPath(r, "/groups")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	group := h.store.GetTagGroupByID(groupID)
	if group == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "group", ID: groupID})
		return
	}

	writeSuccess(w, group)
}

func (h *Handler) ListTagGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	groups := h.store.ListTagGroups()

	result := make([]common.TagGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}

	writeSuccess(w, common.ListTagGroupsResponse{Groups: result})
}

func (h *Handler) ApplyTagToUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tag := h.store.GetTagByID(tagID)
	if tag == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	var req common.ApplyTagToUserRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Operator == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "operator is required"})
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "user_id is required"})
		return
	}

	if err := validateTagValue(tag, req.TagValue); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	h.store.ApplyTag(req.UserID, tagID, req.TagValue)

	details, _ := json.Marshal(map[string]interface{}{
		"user_id":    req.UserID,
		"tag_name":   tag.Name,
		"tag_value":  req.TagValue,
	})
	h.logAudit(req.Operator, "APPLY", "tag:"+tagID, string(details))

	writeSuccess(w, nil)
}

func (h *Handler) BatchApplyTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tag := h.store.GetTagByID(tagID)
	if tag == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	var req common.BatchApplyTagRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Operator == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "operator is required"})
		return
	}
	if len(req.UserIDs) == 0 {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "at least one user_id is required"})
		return
	}
	if len(req.UserIDs) > MaxBatchUsers {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: fmt.Sprintf("maximum %d users per batch", MaxBatchUsers)})
		return
	}

	if err := validateTagValue(tag, req.TagValue); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	for _, userID := range req.UserIDs {
		h.store.ApplyTag(userID, tagID, req.TagValue)
	}

	details, _ := json.Marshal(map[string]interface{}{
		"user_count":  len(req.UserIDs),
		"tag_name":    tag.Name,
		"tag_value":   req.TagValue,
	})
	h.logAudit(req.Operator, "BATCH_APPLY", "tag:"+tagID, string(details))

	writeSuccess(w, map[string]interface{}{
		"processed_users": len(req.UserIDs),
	})
}

func (h *Handler) RemoveUserTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tagID, err := getIDFromPath(r, "/tags")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	userID := r.URL.Query().Get("user_id")
	operator := r.URL.Query().Get("operator")

	if userID == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "user_id is required"})
		return
	}
	if operator == "" {
		writeError(w, http.StatusBadRequest, &ValidationError{Message: "operator is required"})
		return
	}

	tag := h.store.GetTagByID(tagID)
	if tag == nil {
		writeError(w, http.StatusNotFound, &NotFoundError{Resource: "tag", ID: tagID})
		return
	}

	h.store.RemoveUserTag(userID, tagID)

	details, _ := json.Marshal(map[string]interface{}{
		"user_id":  userID,
		"tag_name": tag.Name,
	})
	h.logAudit(operator, "REMOVE", "tag:"+tagID, string(details))

	writeSuccess(w, nil)
}

func (h *Handler) GetUserTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	userID, err := getIDFromPath(r, "/users")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	userTags := h.store.GetUserTags(userID)

	result := make([]common.UserTagWithInfo, 0, len(userTags))
	for _, ut := range userTags {
		tag := h.store.GetTagByID(ut.TagID)
		if tag != nil {
			result = append(result, common.UserTagWithInfo{
				TagID:     ut.TagID,
				TagName:   tag.Name,
				TagType:   tag.Type,
				TagValue:  ut.TagValue,
				CreatedAt: ut.CreatedAt,
			})
		}
	}

	writeSuccess(w, common.GetUserTagsResponse{Tags: result})
}

func (h *Handler) FilterUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.FilterUsersRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := validateFilterGroup(&req.Filter); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	allUserTags := h.store.GetAllUserTags()
	allTags := make(map[string]*common.Tag)
	for _, tag := range h.store.ListTags() {
		allTags[tag.ID] = tag
	}

	var matchedUserIDs []string
	for userID, userTags := range allUserTags {
		match, err := evaluateFilter(&req.Filter, userID, userTags, allTags)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if match {
			matchedUserIDs = append(matchedUserIDs, userID)
		}
	}

	writeSuccess(w, common.FilterUsersResponse{UserIDs: matchedUserIDs})
}

func (h *Handler) GetTagStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	stats := h.store.GetTagUsageStats()

	writeSuccess(w, common.GetTagStatsResponse{Stats: stats})
}

func (h *Handler) GetTagTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	trends := h.store.GetTagTrend(30)

	writeSuccess(w, common.GetTagTrendResponse{Trends: trends})
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	logs := h.store.ListAuditLogs()

	result := make([]common.AuditLog, 0, len(logs))
	for _, l := range logs {
		result = append(result, *l)
	}

	writeSuccess(w, common.ListAuditLogsResponse{Logs: result})
}

func formatTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
