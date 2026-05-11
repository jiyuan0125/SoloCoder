package main

import (
	"encoding/json"
	"net/http"

	"github.com/drivingschool/common"
)

func setupCoachRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/coaches/add", handleAddCoach)
	mux.HandleFunc("/api/coaches/list", handleListCoaches)
	mux.HandleFunc("/api/coaches/schedule", handleSetCoachSchedule)
	mux.HandleFunc("/api/coaches/practice", handleBookPractice)
	mux.HandleFunc("/api/coaches/", handleGetCoach)
}

func handleAddCoach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.AddCoachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	coach, err := store.AddCoach(req.Name, req.TeachingType, req.Phone)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(coach))
}

func handleListCoaches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	coaches := store.ListCoaches()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(coaches))
}

func handleGetCoach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	id := r.URL.Path[len("/api/coaches/"):]
	if id == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("缺少教练ID"))
		return
	}

	coach, err := store.GetCoach(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(coach))
}

func handleSetCoachSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.SetCoachScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	if err := store.SetCoachSchedule(req.CoachID, req.WeekStart, req.Slots); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"message": "排班设置成功"}))
}

func handleBookPractice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.BookPracticeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	booking, err := store.BookPractice(req.StudentID, req.CoachID, req.TimeSlot, req.Subject)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(booking))
}
