package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"housekeeping-training/common"
	"housekeeping-training/core"
)

type Server struct {
	store *core.Store
}

func NewServer() *Server {
	store := core.NewStore()
	store.InitDefaultCourses()
	return &Server{store: store}
}

func writeResponse(w http.ResponseWriter, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Response{
		Success: success,
		Message: message,
		Data:    data,
	})
}

func (s *Server) handleCourses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		courses := s.store.ListCourses()
		writeResponse(w, true, "", courses)
	case http.MethodPost:
		var req common.CreateCourseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		course, err := s.store.CreateCourse(req.Name, req.Level, req.Hours, req.Instructor, req.Fee, req.MaxCapacity)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "course created", course)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		schedules := s.store.ListSchedules()
		writeResponse(w, true, "", schedules)
	case http.MethodPost:
		var req common.CreateScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		schedule, err := s.store.CreateSchedule(req.CourseID, req.StartDate, req.TimeSlot, req.Classroom, req.TotalClasses)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "schedule created", schedule)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStudents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		students := s.store.ListStudents()
		writeResponse(w, true, "", students)
	case http.MethodPost:
		var req common.CreateStudentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		student, err := s.store.CreateStudent(req.Name, req.IDCard, req.Phone, req.Education)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "student created", student)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEnrollments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		enrollments := s.store.ListEnrollments()
		writeResponse(w, true, "", enrollments)
	case http.MethodPost:
		var req common.EnrollRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		enrollment, err := s.store.EnrollStudent(req.StudentID, req.ScheduleID)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "enrolled successfully", enrollment)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePaymentFirst(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/pay-first/")
	if id == "" {
		writeResponse(w, false, "enrollment ID required", nil)
		return
	}
	err := s.store.PayFirstInstallment(id)
	if err != nil {
		writeResponse(w, false, err.Error(), nil)
		return
	}
	writeResponse(w, true, "first installment paid", nil)
}

func (s *Server) handlePaymentSecond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/pay-second/")
	if id == "" {
		writeResponse(w, false, "enrollment ID required", nil)
		return
	}
	err := s.store.PaySecondInstallment(id)
	if err != nil {
		writeResponse(w, false, err.Error(), nil)
		return
	}
	writeResponse(w, true, "second installment paid", nil)
}

func (s *Server) handleAttendance(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		enrollmentID := r.URL.Query().Get("enrollment_id")
		if enrollmentID != "" {
			records := s.store.ListAttendanceRecords(enrollmentID)
			writeResponse(w, true, "", records)
		} else {
			writeResponse(w, true, "", nil)
		}
	case http.MethodPost:
		var req common.RecordAttendanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		record, err := s.store.RecordAttendance(req.EnrollmentID, req.ClassDate, req.Status)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "attendance recorded", record)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAttendanceRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/attendance-rate/")
	if id == "" {
		writeResponse(w, false, "enrollment ID required", nil)
		return
	}
	rate, err := s.store.GetAttendanceRate(id)
	if err != nil {
		writeResponse(w, false, err.Error(), nil)
		return
	}
	canTake, reason, _ := s.store.CanTakeExam(id)
	writeResponse(w, true, "", map[string]interface{}{
		"rate":           rate,
		"can_take_exam":  canTake,
		"reason":         reason,
	})
}

func (s *Server) handleExams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		exams := s.store.ListExams()
		writeResponse(w, true, "", exams)
	case http.MethodPost:
		var req common.RecordExamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		canTake, reason, err := s.store.CanTakeExam(req.EnrollmentID)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		if !canTake && !req.IsRetake {
			writeResponse(w, false, reason, nil)
			return
		}
		exam, err := s.store.RecordExam(req.EnrollmentID, req.ExamDate, req.WrittenScore, req.PracticalScore, req.IsRetake)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "exam recorded", exam)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCertificates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	studentID := r.URL.Query().Get("student_id")
	if studentID != "" {
		certs := s.store.GetStudentCertificates(studentID)
		writeResponse(w, true, "", certs)
	} else {
		certs := s.store.ListCertificates()
		writeResponse(w, true, "", certs)
	}
}

func (s *Server) handleEmployers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		employers := s.store.ListEmployers()
		writeResponse(w, true, "", employers)
	case http.MethodPost:
		var req common.CreateEmployerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		employer, err := s.store.CreateEmployer(req.Name, req.Contact, req.Phone)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "employer created", employer)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jobs := s.store.ListJobPostings()
		writeResponse(w, true, "", jobs)
	case http.MethodPost:
		var req common.CreateJobPostingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		job, err := s.store.CreateJobPosting(req.EmployerID, req.Position, req.RequiredCert, req.MinSalary, req.MaxSalary, req.WorkLocation)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "job posting created", job)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jobID := r.URL.Query().Get("generate_for_job")
		if jobID != "" {
			recs, err := s.store.GenerateRecommendations(jobID)
			if err != nil {
				writeResponse(w, false, err.Error(), nil)
				return
			}
			writeResponse(w, true, "recommendations generated", recs)
		} else {
			recs := s.store.ListRecommendations()
			writeResponse(w, true, "", recs)
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProcessRecommendation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/recommendations/")
	id = strings.TrimSuffix(id, "/process")
	if id == "" {
		writeResponse(w, false, "recommendation ID required", nil)
		return
	}
	var req common.ProcessRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, false, "invalid request body", nil)
		return
	}
	rec, err := s.store.ProcessRecommendation(id, req.InterviewDate, req.Result, req.Notes)
	if err != nil {
		writeResponse(w, false, err.Error(), nil)
		return
	}
	writeResponse(w, true, "recommendation updated", rec)
}

func (s *Server) handleRefunds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		refunds := s.store.ListRefunds()
		writeResponse(w, true, "", refunds)
	case http.MethodPost:
		var req common.RefundRequestReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(w, false, "invalid request body", nil)
			return
		}
		refund, err := s.store.RequestRefund(req.EnrollmentID, req.Reason)
		if err != nil {
			writeResponse(w, false, err.Error(), nil)
			return
		}
		writeResponse(w, true, "refund request created", refund)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "server port")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("PORT")
		if envPort != "" {
			p, err := strconv.Atoi(envPort)
			if err == nil && p > 0 {
				port = p
			}
		}
	}
	if port == 0 {
		port = 8080
	}

	server := NewServer()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/courses", server.handleCourses)
	mux.HandleFunc("/api/schedules", server.handleSchedules)
	mux.HandleFunc("/api/students", server.handleStudents)
	mux.HandleFunc("/api/enrollments", server.handleEnrollments)
	mux.HandleFunc("/api/pay-first/", server.handlePaymentFirst)
	mux.HandleFunc("/api/pay-second/", server.handlePaymentSecond)
	mux.HandleFunc("/api/attendance", server.handleAttendance)
	mux.HandleFunc("/api/attendance-rate/", server.handleAttendanceRate)
	mux.HandleFunc("/api/exams", server.handleExams)
	mux.HandleFunc("/api/certificates", server.handleCertificates)
	mux.HandleFunc("/api/employers", server.handleEmployers)
	mux.HandleFunc("/api/jobs", server.handleJobs)
	mux.HandleFunc("/api/recommendations", server.handleRecommendations)
	mux.HandleFunc("/api/recommendations/", server.handleProcessRecommendation)
	mux.HandleFunc("/api/refunds", server.handleRefunds)

	fmt.Printf("Server starting on port %d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
