package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"equip-inspect/models"
	"equip-inspect/service"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, code int, message string, currentStatus ...string) {
	err := APIError{Code: code, Message: message}
	if len(currentStatus) > 0 {
		err.Status = currentStatus[0]
	}
	writeJSON(w, code, err)
}

func parseID(r *http.Request, param string) (int64, error) {
	idStr := r.PathValue(param)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", param)
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func readBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	inspectorIDStr := r.URL.Query().Get("inspector_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	inspectorID := int64(0)
	if inspectorIDStr != "" {
		inspectorID, _ = strconv.ParseInt(inspectorIDStr, 10, 64)
	}
	limit := 50
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	offset := 0
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	tasks, err := service.ListTasks(status, inspectorID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := service.GetTaskByID(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func AssignTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		InspectorID int64 `json:"inspector_id"`
	}
	if err = readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.InspectorID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid inspector_id")
		return
	}

	task, err := service.GetTaskByID(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	if task.InspectorID != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("task already assigned to inspector %d", *task.InspectorID))
		return
	}

	assignedTask, err := service.AssignTask(taskID, req.InspectorID)
	if err != nil {
		if strings.Contains(err.Error(), "task already taken") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "cannot assign") {
			writeError(w, http.StatusBadRequest, err.Error(), string(task.Status))
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assignedTask)
}

func CheckInPointHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := service.GetTaskByID(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	nextPoint, err := service.GetNextExpectedPoint(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req struct {
		PointID     int64 `json:"point_id"`
		InspectorID int64 `json:"inspector_id"`
	}
	if err = readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	checkedPoint, err := service.CheckInPoint(taskID, req.PointID, req.InspectorID)
	if err != nil {
		if nextPoint != nil && strings.Contains(err.Error(), "must check") {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":              http.StatusBadRequest,
				"message":           err.Error(),
				"next_expected_id":  nextPoint.PointID,
				"next_expected_order": nextPoint.OrderIndex,
			})
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, checkedPoint)
}

type SubmitRecordRequest struct {
	InspectorID int64                    `json:"inspector_id"`
	Records     []models.InspectionRecord `json:"records"`
}

func SubmitInspectionHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := service.GetTaskByID(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	var req SubmitRecordRequest
	if err = readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	score, err := service.SubmitInspectionRecords(taskID, req.InspectorID, req.Records)
	if err != nil {
		if strings.Contains(err.Error(), "invalid state") || strings.Contains(err.Error(), "cannot submit") {
			writeError(w, http.StatusBadRequest, err.Error(), string(task.Status))
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updatedTask, _ := service.GetTaskByID(taskID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"score": score,
		"task":  updatedTask,
	})
}

type TransitionRequest struct {
	NewStatus string `json:"new_status"`
	UserID    int64  `json:"user_id"`
	Comment   string `json:"comment"`
}

func TransitionTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := service.GetTaskByID(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	var req TransitionRequest
	if err = readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewStatus == "review_fail" && strings.TrimSpace(req.Comment) == "" {
		writeError(w, http.StatusBadRequest, "review_fail requires comment")
		return
	}

	updatedTask, err := service.TransitionTask(taskID, models.TaskStatus(req.NewStatus), req.UserID, req.Comment)
	if err != nil {
		if strings.Contains(err.Error(), "invalid state transition") {
			writeError(w, http.StatusBadRequest, err.Error(), string(task.Status))
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updatedTask)
}

func GetTaskPointsHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := service.GetTaskByID(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	points, err := service.GetTaskPoints(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	nextPoint, _ := service.GetNextExpectedPoint(taskID)
	var nextID *int64
	if nextPoint != nil {
		id := nextPoint.PointID
		nextID = &id
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"points":           points,
		"next_expected_id": nextID,
	})
}

func ListPlansHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 50
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	offset := 0
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	plans, err := service.ListPlans(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

type CreatePlanRequest struct {
	Name      string  `json:"name"`
	Frequency string  `json:"frequency"`
	PointIDs  []int64 `json:"point_ids"`
}

func CreatePlanHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatePlanRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Frequency == "" || len(req.PointIDs) == 0 {
		writeError(w, http.StatusBadRequest, "name, frequency, and point_ids are required")
		return
	}

	plan, err := service.CreatePlan(req.Name, req.Frequency, req.PointIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

func GenerateTasksHandler(w http.ResponseWriter, r *http.Request) {
	planID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tasks, err := service.GenerateTasksFromPlan(planID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tasks == nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "no new tasks needed at this time"})
		return
	}
	writeJSON(w, http.StatusCreated, tasks)
}

func ListRepairOrdersHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	deviceIDStr := r.URL.Query().Get("device_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	deviceID := int64(0)
	if deviceIDStr != "" {
		deviceID, _ = strconv.ParseInt(deviceIDStr, 10, 64)
	}
	limit := 50
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	offset := 0
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	orders, err := service.ListRepairOrders(status, deviceID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

type AssignRepairRequest struct {
	RepairerID int64 `json:"repairer_id"`
}

func AssignRepairHandler(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req AssignRepairRequest
	if err = readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := service.AssignRepairOrder(orderID, req.RepairerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func CompleteRepairHandler(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := service.CompleteRepairOrder(orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func GetStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	stats, err := service.GetStatistics(date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
