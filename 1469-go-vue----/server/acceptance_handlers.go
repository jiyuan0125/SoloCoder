package main

import (
	"encoding/json"
	"net/http"

	"supervision-log-system/common"
)

func (r *Router) createAcceptanceRecord(w http.ResponseWriter, req *http.Request) {
	var reqData common.CreateAcceptanceRecordRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	record, err := r.acceptanceService.Create(&reqData)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	sendSuccess(w, record)
}

func (r *Router) listAcceptanceRecords(w http.ResponseWriter) {
	records := r.acceptanceService.List()
	sendSuccess(w, records)
}

func (r *Router) createRectificationNotice(w http.ResponseWriter, req *http.Request) {
	var reqData common.CreateRectificationNoticeRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	notice, err := r.acceptanceService.CreateRectificationNotice(&reqData)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, notice)
}

func (r *Router) completeRectification(w http.ResponseWriter, req *http.Request) {
	var reqData common.CompleteRectificationRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	if err := r.acceptanceService.CompleteRectification(&reqData); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, map[string]bool{"completed": true})
}

func (r *Router) recheckAcceptance(w http.ResponseWriter, req *http.Request) {
	var reqData common.RecheckAcceptanceRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	record, err := r.acceptanceService.Recheck(&reqData)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, record)
}
