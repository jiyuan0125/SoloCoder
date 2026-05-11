package main

import (
	"encoding/json"
	"net/http"

	"github.com/drivingschool/common"
)

func setupExamRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/exams/plans/create", handleCreateExamPlan)
	mux.HandleFunc("/api/exams/plans/list", handleListExamPlans)
	mux.HandleFunc("/api/exams/book", handleBookExam)
	mux.HandleFunc("/api/exams/confirm", handleConfirmWaitingExam)
	mux.HandleFunc("/api/exams/cancel", handleCancelExamBooking)
	mux.HandleFunc("/api/exams/approve", handleApproveCancelExam)
	mux.HandleFunc("/api/exams/bookings", handleListExamBookings)
}

func handleCreateExamPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.CreateExamPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	plan, err := store.CreateExamPlan(req.Date, req.Subject, req.Venue, req.TotalQuota)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(plan))
}

func handleListExamPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	plans := store.ListExamPlans()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(plans))
}

func handleBookExam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.BookExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	booking, err := store.BookExam(req.StudentID, req.ExamPlanID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(booking))
}

func handleConfirmWaitingExam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.ConfirmWaitingExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	if err := store.ConfirmWaitingExam(req.ExamBookingID); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"message": "候补确认成功"}))
}

func handleCancelExamBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.CancelExamBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	needsApproval, err := store.CancelExamBooking(req.StudentID, req.ExamBookingID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	if needsApproval {
		writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
			"message":        "距离考试不足3天，需要管理员审批",
			"needs_approval": true,
		}))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"message":        "取消成功",
		"needs_approval": false,
	}))
}

func handleApproveCancelExam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.ApproveCancelExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	if err := store.ApproveCancelExam(req.ExamBookingID, req.Approved); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	if req.Approved {
		writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"message": "取消审批通过"}))
	} else {
		writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"message": "取消审批被拒绝"}))
	}
}

func handleListExamBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	studentID := r.URL.Query().Get("student_id")
	bookings := store.ListExamBookings(studentID)
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(bookings))
}
