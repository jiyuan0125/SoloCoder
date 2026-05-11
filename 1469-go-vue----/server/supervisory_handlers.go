package main

import (
	"encoding/json"
	"net/http"

	"supervision-log-system/common"
)

func (r *Router) createSupervisoryRecord(w http.ResponseWriter, req *http.Request) {
	var reqData common.CreateSupervisoryRecordRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	record, err := r.supervisoryService.Create(&reqData)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, record)
}

func (r *Router) listSupervisoryRecords(w http.ResponseWriter) {
	records := r.supervisoryService.List()
	sendSuccess(w, records)
}

func (r *Router) submitSupervisoryRecord(w http.ResponseWriter, req *http.Request) {
	var reqData common.IDResponse
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	if err := r.supervisoryService.Submit(reqData.ID); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, map[string]bool{"submitted": true})
}

func (r *Router) addSupervisorySupplement(w http.ResponseWriter, req *http.Request) {
	var reqData common.AddSupervisorySupplementRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	if err := r.supervisoryService.AddSupplement(&reqData); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, map[string]bool{"added": true})
}
