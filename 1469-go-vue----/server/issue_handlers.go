package main

import (
	"encoding/json"
	"net/http"

	"supervision-log-system/common"
)

func (r *Router) createIssue(w http.ResponseWriter, req *http.Request) {
	var reqData common.CreateIssueRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	issue, err := r.issueService.Create(&reqData)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, issue)
}

func (r *Router) listIssues(w http.ResponseWriter) {
	issues := r.issueService.List()
	sendSuccess(w, issues)
}

func (r *Router) updateIssueStatus(w http.ResponseWriter, req *http.Request) {
	var reqData common.UpdateIssueStatusRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	if err := r.issueService.UpdateStatus(&reqData); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, map[string]bool{"updated": true})
}

func (r *Router) checkOverdueIssues(w http.ResponseWriter) {
	r.issueService.CheckOverdue()
	sendSuccess(w, map[string]bool{"checked": true})
}
