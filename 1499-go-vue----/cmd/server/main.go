package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"community-activity-platform/pkg/common"
	"community-activity-platform/pkg/core"
)

type Server struct {
	storage             *core.Storage
	participantService  *core.ParticipantService
	activityService     *core.ActivityService
	registrationService *core.RegistrationService
	checkInService      *core.CheckInService
	feedbackService     *core.FeedbackService
}

func NewServer() *Server {
	storage := core.NewStorage()
	return &Server{
		storage:             storage,
		participantService:  core.NewParticipantService(storage),
		activityService:     core.NewActivityService(storage),
		registrationService: core.NewRegistrationService(storage),
		checkInService:      core.NewCheckInService(storage),
		feedbackService:     core.NewFeedbackService(storage),
	}
}

func (s *Server) writeResponse(w http.ResponseWriter, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	resp := common.APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := common.APIResponse{
		Success: false,
		Message: err.Error(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) RegisterParticipant(w http.ResponseWriter, r *http.Request) {
	var req common.RegisterParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	participant, err := s.participantService.Register(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "participant registered successfully", participant)
}

func (s *Server) GetParticipant(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	phone := r.URL.Query().Get("phone")

	var participant *common.Participant
	var err error

	if id != "" {
		participant, err = s.participantService.GetByID(id)
	} else if phone != "" {
		participant, err = s.participantService.GetByPhone(phone)
	} else {
		s.writeError(w, http.StatusBadRequest, common.ErrMissingRequiredField)
		return
	}

	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, true, "", participant)
}

func (s *Server) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var req common.CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	activities, err := s.activityService.Create(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "activities created successfully", activities)
}

func (s *Server) GetActivity(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, common.ErrMissingRequiredField)
		return
	}

	activity, err := s.activityService.GetDetail(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, true, "", activity)
}

func (s *Server) ListActivities(w http.ResponseWriter, r *http.Request) {
	var req common.ListActivitiesRequest
	req.Page = 1
	req.PageSize = 20

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := common.ActivityStatus(statusStr)
		req.Status = &status
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		actType := common.ActivityType(typeStr)
		req.Type = &actType
	}

	activities, err := s.activityService.List(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "", activities)
}

func (s *Server) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, common.ErrMissingRequiredField)
		return
	}

	var req common.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	activity, err := s.activityService.Update(id, req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "activity updated successfully", activity)
}

func (s *Server) CancelActivity(w http.ResponseWriter, r *http.Request) {
	var req common.CancelActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	activity, err := s.activityService.Cancel(req.ActivityID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "activity cancelled successfully", activity)
}

func (s *Server) RegisterActivity(w http.ResponseWriter, r *http.Request) {
	var req common.RegisterActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	registration, err := s.registrationService.Register(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "registration successful", registration)
}

func (s *Server) CancelRegistration(w http.ResponseWriter, r *http.Request) {
	var req common.CancelRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.registrationService.Cancel(req.RegistrationID); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "registration cancelled successfully", nil)
}

func (s *Server) CheckIn(w http.ResponseWriter, r *http.Request) {
	var req common.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	registration, err := s.checkInService.CheckIn(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "check-in successful", registration)
}

func (s *Server) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	feedback, err := s.feedbackService.Submit(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "feedback submitted successfully", feedback)
}

func (s *Server) GetActivitySummary(w http.ResponseWriter, r *http.Request) {
	activityID := r.URL.Query().Get("activity_id")
	if activityID == "" {
		s.writeError(w, http.StatusBadRequest, common.ErrMissingRequiredField)
		return
	}

	summary, err := s.feedbackService.GetSummary(activityID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, true, "", summary)
}

func (s *Server) ListAnalysisTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := s.feedbackService.ListAnalysisTodos()
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(w, true, "", todos)
}

func main() {
	var port int
	var envPort string

	flag.IntVar(&port, "port", 8080, "Server port")
	flag.Parse()

	if envPort = os.Getenv("APP_PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}

	server := NewServer()

	mux := http.NewServeMux()

	mux.HandleFunc("/participant/register", server.RegisterParticipant)
	mux.HandleFunc("/participant", server.GetParticipant)

	mux.HandleFunc("/activity/create", server.CreateActivity)
	mux.HandleFunc("/activity", server.GetActivity)
	mux.HandleFunc("/activity/list", server.ListActivities)
	mux.HandleFunc("/activity/update", server.UpdateActivity)
	mux.HandleFunc("/activity/cancel", server.CancelActivity)

	mux.HandleFunc("/registration/register", server.RegisterActivity)
	mux.HandleFunc("/registration/cancel", server.CancelRegistration)

	mux.HandleFunc("/checkin", server.CheckIn)

	mux.HandleFunc("/feedback/submit", server.SubmitFeedback)
	mux.HandleFunc("/activity/summary", server.GetActivitySummary)
	mux.HandleFunc("/analysis/todos", server.ListAnalysisTodos)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on port %d...", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}
