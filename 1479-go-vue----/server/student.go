package main

import (
	"encoding/json"
	"net/http"

	"github.com/drivingschool/common"
)

func setupStudentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/students/register", handleRegisterStudent)
	mux.HandleFunc("/api/students/list", handleListStudents)
	mux.HandleFunc("/api/students/status", handleUpdateSubjectStatus)
	mux.HandleFunc("/api/students/hours", handleAddStudyHours)
	mux.HandleFunc("/api/students/", handleGetStudent)
}

func handleRegisterStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.RegisterStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	student, err := store.RegisterStudent(req.Name, req.IDCard, req.Phone, req.VehicleType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(student))
}

func handleListStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	students := store.ListStudents()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(students))
}

func handleGetStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	id := r.URL.Path[len("/api/students/"):]
	if id == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("缺少学员ID"))
		return
	}

	student, err := store.GetStudent(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(student))
}

func handleUpdateSubjectStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.UpdateSubjectStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	if err := store.UpdateSubjectStatus(req.StudentID, req.Subject, req.Status); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"message": "状态更新成功"}))
}

func handleAddStudyHours(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("方法不允许"))
		return
	}

	var req common.AddStudyHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("无效的请求体"))
		return
	}

	if err := store.AddStudyHours(req.StudentID, req.Subject, req.Hours); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string]string{"message": "学时添加成功"}))
}
