package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"dental-clinic/models"
	"dental-clinic/store"

	"github.com/gorilla/mux"
)

type Server struct {
	store *store.Store
}

func NewServer() *Server {
	s := store.NewStore()
	s.SeedData()
	return &Server{store: s}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func (s *Server) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

type BookAppointmentRequest struct {
	PatientID  string   `json:"patient_id"`
	DoctorID   string   `json:"doctor_id"`
	Date       string   `json:"date"`
	StartTime  string   `json:"start_time"`
	Treatments []string `json:"treatments"`
}

func (s *Server) BookAppointment(w http.ResponseWriter, r *http.Request) {
	var req BookAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	_, err := s.store.GetPatient(req.PatientID)
	if err != nil {
		s.handleError(w, fmt.Errorf("patient not found"), http.StatusBadRequest)
		return
	}

	_, err = s.store.GetDoctor(req.DoctorID)
	if err != nil {
		s.handleError(w, fmt.Errorf("doctor not found"), http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		s.handleError(w, fmt.Errorf("invalid date format"), http.StatusBadRequest)
		return
	}

	timeParts := strings.Split(req.StartTime, ":")
	if len(timeParts) != 2 {
		s.handleError(w, fmt.Errorf("invalid time format"), http.StatusBadRequest)
		return
	}

	hour, _ := strconv.Atoi(timeParts[0])
	minute, _ := strconv.Atoi(timeParts[1])

	startTime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, date.Location())
	endTime := startTime.Add(30 * time.Minute)

	conflict, samePatient := s.store.CheckAppointmentConflict(req.DoctorID, startTime, req.PatientID)
	if conflict && !samePatient {
		s.handleError(w, fmt.Errorf("该时段已被其他患者预约"), http.StatusConflict)
		return
	}

	appointment, err := s.store.MergeOrCreateAppointment(req.PatientID, req.DoctorID, startTime, endTime, date, req.Treatments)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, appointment)
}

func (s *Server) ListDoctors(w http.ResponseWriter, r *http.Request) {
	doctors := s.store.ListDoctors()
	s.respondJSON(w, doctors)
}

func (s *Server) ListAppointments(w http.ResponseWriter, r *http.Request) {
	appointments := s.store.ListAppointments()
	s.respondJSON(w, appointments)
}

type CreatePlanRequest struct {
	PatientID       string        `json:"patient_id"`
	DiscountPercent int           `json:"discount_percent"`
	IsSurgical      bool          `json:"is_surgical"`
	Steps           []*models.Step `json:"steps"`
}

func (s *Server) CreateTreatmentPlan(w http.ResponseWriter, r *http.Request) {
	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	if req.DiscountPercent < 80 || req.DiscountPercent > 100 {
		s.handleError(w, fmt.Errorf("折扣比例超出允许范围（80%%-100%%）"), http.StatusBadRequest)
		return
	}

	for _, step := range req.Steps {
		if step.ExpectedDate == "" {
			s.handleError(w, fmt.Errorf("步骤预计执行日期不能为空"), http.StatusBadRequest)
			return
		}
	}

	plan, err := s.store.CreateTreatmentPlan(req.PatientID, req.DiscountPercent, req.IsSurgical, req.Steps)
	if err != nil {
		if err.Error() == "duplicate expected date in steps" {
			s.handleError(w, fmt.Errorf("同一计划内步骤预计日期不能重复"), http.StatusConflict)
			return
		}
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	totalFee, err := s.store.CalculateTotalFee(plan)
	if err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"plan":      plan,
		"total_fee": totalFee,
	}

	s.respondJSON(w, result)
}

func (s *Server) ListTreatmentPlans(w http.ResponseWriter, r *http.Request) {
	s.store.CheckOverduePlans()
	plans := s.store.ListTreatmentPlans()
	s.respondJSON(w, plans)
}

type UpdateStepRequest struct {
	StepID string `json:"step_id"`
	Status string `json:"status"`
}

func (s *Server) UpdateStepStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["id"]

	var req UpdateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	newStatus := models.StepStatus(req.Status)
	err := s.store.UpdateStepStatus(planID, req.StepID, newStatus)
	if err != nil {
		if err.Error() == "cannot revert completed step" || err.Error() == "must go through in_progress first" {
			s.handleError(w, err, http.StatusBadRequest)
			return
		}
		s.handleError(w, err, http.StatusNotFound)
		return
	}

	plan, _ := s.store.GetTreatmentPlan(planID)
	s.respondJSON(w, plan)
}

func (s *Server) ListFollowUps(w http.ResponseWriter, r *http.Request) {
	followUps := s.store.ListFollowUps()
	s.respondJSON(w, followUps)
}

type CompleteFollowUpRequest struct {
	Method   string `json:"method"`
	Feedback string `json:"feedback"`
	Notes    string `json:"notes"`
}

func (s *Server) CompleteFollowUp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followUpID := vars["id"]

	var req CompleteFollowUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	err := s.store.CompleteFollowUp(followUpID, models.FollowUpMethod(req.Method), models.FeedbackType(req.Feedback), req.Notes)
	if err != nil {
		s.handleError(w, err, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) BatchCompleteFollowUps(w http.ResponseWriter, r *http.Request) {
	body := struct {
		IDs        []string `json:"ids"`
		Method     string   `json:"method"`
		Feedback   string   `json:"feedback"`
		Notes      string   `json:"notes"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	successCount := 0
	for _, id := range body.IDs {
		err := s.store.CompleteFollowUp(id, models.FollowUpMethod(body.Method), models.FeedbackType(body.Feedback), body.Notes)
		if err == nil {
			successCount++
		}
	}

	s.respondJSON(w, map[string]int{"completed": successCount})
}

func (s *Server) ListTodos(w http.ResponseWriter, r *http.Request) {
	todos := s.store.ListTodos()
	s.respondJSON(w, todos)
}

func (s *Server) ListPatients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]*models.Patient{
		{ID: "p1", Name: "患者甲", Phone: "13800000001"},
		{ID: "p2", Name: "患者乙", Phone: "13800000002"},
		{ID: "p3", Name: "患者丙", Phone: "13800000003"},
	})
}

func (s *Server) GetTodayPatients(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	todayStr := now.Format("2006-01-02")

	appointments := s.store.ListAppointments()
	result := []map[string]interface{}{}

	for _, appt := range appointments {
		if appt.Date == todayStr {
			patient, _ := s.store.GetPatient(appt.PatientID)
			doctor, _ := s.store.GetDoctor(appt.DoctorID)
			result = append(result, map[string]interface{}{
				"appointment": appt,
				"patient":     patient,
				"doctor":      doctor,
			})
		}
	}

	s.respondJSON(w, result)
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "Server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	server := NewServer()
	router := mux.NewRouter()

	router.HandleFunc("/api/doctors", server.ListDoctors).Methods("GET")
	router.HandleFunc("/api/patients", server.ListPatients).Methods("GET")
	router.HandleFunc("/api/today-patients", server.GetTodayPatients).Methods("GET")

	router.HandleFunc("/api/appointments", server.ListAppointments).Methods("GET")
	router.HandleFunc("/api/appointments", server.BookAppointment).Methods("POST")

	router.HandleFunc("/api/treatment-plans", server.ListTreatmentPlans).Methods("GET")
	router.HandleFunc("/api/treatment-plans", server.CreateTreatmentPlan).Methods("POST")
	router.HandleFunc("/api/treatment-plans/{id}/steps", server.UpdateStepStatus).Methods("PUT")

	router.HandleFunc("/api/follow-ups", server.ListFollowUps).Methods("GET")
	router.HandleFunc("/api/follow-ups/{id}/complete", server.CompleteFollowUp).Methods("POST")
	router.HandleFunc("/api/follow-ups/batch-complete", server.BatchCompleteFollowUps).Methods("POST")

	router.HandleFunc("/api/todos", server.ListTodos).Methods("GET")

	http.Handle("/", enableCORS(router))

	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
