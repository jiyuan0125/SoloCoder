package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"coldchain/common"
	"coldchain/core"
)

type Handlers struct {
	service *core.Service
}

func NewHandlers(service *core.Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req common.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DeviceID == "" || req.CargoType == "" || req.StartWarehouse == "" || req.EndWarehouse == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	task, err := h.service.CreateTask(req.DeviceID, req.CargoType, req.StartWarehouse, req.EndWarehouse)
	if err != nil {
		var invalidCargoErr *core.InvalidCargoTypeError
		if errors.As(err, &invalidCargoErr) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		var existsErr *core.TaskExistsError
		if errors.As(err, &existsErr) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.CreateTaskResponse{
		TaskID:         task.ID,
		DeviceID:       task.DeviceID,
		CargoType:      string(task.CargoType),
		StartWarehouse: task.StartWarehouse,
		EndWarehouse:   task.EndWarehouse,
		StartTime:      task.StartTime,
		Status:         string(task.Status),
		TempMin:        task.TempMin,
		TempMax:        task.TempMax,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.service.ListTasks()

	resp := common.ListTasksResponse{
		Tasks: make([]common.TaskResponse, 0, len(tasks)),
	}

	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, common.TaskResponse{
			TaskID:         task.ID,
			DeviceID:       task.DeviceID,
			CargoType:      string(task.CargoType),
			StartWarehouse: task.StartWarehouse,
			EndWarehouse:   task.EndWarehouse,
			StartTime:      task.StartTime,
			EndTime:        task.EndTime,
			Status:         string(task.Status),
			TempMin:        task.TempMin,
			TempMax:        task.TempMax,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.TaskResponse{
		TaskID:         task.ID,
		DeviceID:       task.DeviceID,
		CargoType:      string(task.CargoType),
		StartWarehouse: task.StartWarehouse,
		EndWarehouse:   task.EndWarehouse,
		StartTime:      task.StartTime,
		EndTime:        task.EndTime,
		Status:         string(task.Status),
		TempMin:        task.TempMin,
		TempMax:        task.TempMax,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) EndTask(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID = strings.TrimSuffix(taskID, "/end")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	if err := h.service.EndTask(taskID); err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		var endedErr *core.TaskAlreadyEndedError
		if errors.As(err, &endedErr) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, common.SuccessResponse{Success: true})
}

func (h *Handlers) SubmitSensorData(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitSensorDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DeviceID == "" {
		writeError(w, http.StatusBadRequest, "device ID is required")
		return
	}

	data := &core.SensorData{
		DeviceID:    req.DeviceID,
		Timestamp:   req.Timestamp,
		Temperature: req.Temperature,
		Humidity:    req.Humidity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Speed:       req.Speed,
	}

	alert, err := h.service.SubmitSensorData(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.SubmitSensorDataResponse{
		Success: true,
	}

	if alert != nil {
		resp.Alert = &common.AlertInfo{
			ID:          alert.ID,
			TaskID:      alert.TaskID,
			Timestamp:   alert.Timestamp,
			Temperature: alert.Temperature,
			Level:       string(alert.Level),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetSensorData(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID = strings.TrimSuffix(taskID, "/data")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	data, err := h.service.GetSensorData(taskID)
	if err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.ListSensorDataResponse{
		Data: make([]common.SensorDataResponse, 0, len(data)),
	}

	for _, d := range data {
		resp.Data = append(resp.Data, common.SensorDataResponse{
			Timestamp:   d.Timestamp,
			Temperature: d.Temperature,
			Humidity:    d.Humidity,
			Latitude:    d.Latitude,
			Longitude:   d.Longitude,
			Speed:       d.Speed,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID = strings.TrimSuffix(taskID, "/alerts")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	alerts, err := h.service.GetAlerts(taskID)
	if err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.ListAlertsResponse{
		Alerts: make([]common.AlertInfo, 0, len(alerts)),
	}

	for _, a := range alerts {
		resp.Alerts = append(resp.Alerts, common.AlertInfo{
			ID:          a.ID,
			TaskID:      a.TaskID,
			Timestamp:   a.Timestamp,
			Temperature: a.Temperature,
			Level:       string(a.Level),
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID = strings.TrimSuffix(taskID, "/stats")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	stats, err := h.service.CalculateStats(taskID)
	if err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.TaskStatsResponse{
		TaskID:       stats.TaskID,
		TotalSamples: stats.TotalSamples,
		ValidSamples: stats.ValidSamples,
		PassRate:     stats.PassRate,
		NeedRecheck:  stats.NeedRecheck,
		TempMax:      stats.TempMax,
		TempMin:      stats.TempMin,
		TempAvg:      stats.TempAvg,
		TempStdDev:   stats.TempStdDev,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetReport(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID = strings.TrimSuffix(taskID, "/report")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	report, err := h.service.GenerateReport(taskID)
	if err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	alertInfos := make([]common.AlertInfo, 0, len(report.Alerts))
	for _, a := range report.Alerts {
		alertInfos = append(alertInfos, common.AlertInfo{
			ID:          a.ID,
			TaskID:      a.TaskID,
			Timestamp:   a.Timestamp,
			Temperature: a.Temperature,
			Level:       string(a.Level),
		})
	}

	resp := common.TaskReportResponse{
		TaskID: report.TaskID,
		Stats: common.TaskStatsResponse{
			TaskID:       report.Stats.TaskID,
			TotalSamples: report.Stats.TotalSamples,
			ValidSamples: report.Stats.ValidSamples,
			PassRate:     report.Stats.PassRate,
			NeedRecheck:  report.Stats.NeedRecheck,
			TempMax:      report.Stats.TempMax,
			TempMin:      report.Stats.TempMin,
			TempAvg:      report.Stats.TempAvg,
			TempStdDev:   report.Stats.TempStdDev,
		},
		Alerts:     alertInfos,
		ExportedAt: report.ExportedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path := r.URL.Path
	
	var exportType string
	if strings.HasSuffix(path, "/export/sensor") {
		exportType = "sensor"
		taskID = strings.TrimSuffix(taskID, "/export/sensor")
	} else if strings.HasSuffix(path, "/export/stats") {
		exportType = "stats"
		taskID = strings.TrimSuffix(taskID, "/export/stats")
	} else if strings.HasSuffix(path, "/export/alerts") {
		exportType = "alerts"
		taskID = strings.TrimSuffix(taskID, "/export/alerts")
	} else {
		writeError(w, http.StatusBadRequest, "invalid export type")
		return
	}

	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	if _, err := h.service.GetTask(taskID); err != nil {
		var notFoundErr *core.TaskNotFoundError
		if errors.As(err, &notFoundErr) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var buf bytes.Buffer
	exporter := h.service.Exporter()

	var err error
	switch exportType {
	case "sensor":
		err = exporter.ExportSensorDataToCSV(taskID, &buf)
	case "stats":
		err = exporter.ExportStatsToCSV(taskID, &buf)
	case "alerts":
		err = exporter.ExportAlertsToCSV(taskID, &buf)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+taskID+"_"+exportType+".csv")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{Error: message})
}
