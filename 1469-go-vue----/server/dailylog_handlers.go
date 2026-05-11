package main

import (
	"encoding/json"
	"net/http"

	"supervision-log-system/common"
)

func (r *Router) submitDailyLog(w http.ResponseWriter, req *http.Request) {
	var reqData common.SubmitDailyLogRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	log, err := r.dailyLogService.Submit(&reqData)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, log)
}

func (r *Router) getDailyLog(w http.ResponseWriter, req *http.Request) {
	var reqData common.GetDailyLogRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	log, err := r.dailyLogService.Get(reqData.Supervisor, reqData.LogDate)
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	sendSuccess(w, log)
}

func (r *Router) generateDailyLogDraft(w http.ResponseWriter, req *http.Request) {
	var reqData common.GetDailyLogRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	log, err := r.dailyLogService.GenerateDraft(reqData.Supervisor, reqData.LogDate)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, log)
}
