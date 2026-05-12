package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"education-platform/database"
	"education-platform/handlers"
	ws "education-platform/websocket"

	"github.com/gorilla/mux"
)

func main() {
	port := getPort()
	dbPath := "./education.db"

	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Printf("Database initialized at: %s\n", dbPath)

	r := mux.NewRouter()

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/courses", handlers.CreateCourse).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/courses", handlers.ListCourses).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/courses/{id}", handlers.GetCourse).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/courses/{id}/status", handlers.UpdateLiveCourseStatus).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/courses/{id}/stats", handlers.GetLiveStats).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/courses/{id}/reviews", handlers.GetCourseReviews).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/students", handlers.RegisterStudent).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/students", handlers.ListStudents).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/students/{id}", handlers.GetStudent).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/students/{id}/enrollments", handlers.GetStudentEnrollments).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/recharge", handlers.RechargeCard).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/enroll", handlers.EnrollCourse).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/progress/{student_id}/{course_id}", handlers.UpdateWatchProgress).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/progress/{student_id}", handlers.GetWatchProgress).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/reviews", handlers.CreateReview).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/stats/subjects", handlers.GetSubjectStats).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/stats/instructors", handlers.GetInstructorStats).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/stats/live", handlers.GetAllLiveStats).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/courses/{id}/ws", ws.HandleWebSocket)
	r.HandleFunc("/api/courses/{id}/vote", ws.CreateVote).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/courses/{id}/votes", ws.GetActiveVotes).Methods("GET", "OPTIONS")

	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Printf("API: http://localhost:%s/api/\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}

	for i, arg := range os.Args {
		if arg == "--port" && i+1 < len(os.Args) {
			if _, err := strconv.Atoi(os.Args[i+1]); err == nil {
				return os.Args[i+1]
			}
		}
		if len(arg) > 7 && arg[:7] == "--port=" {
			if port := arg[7:]; port != "" {
				return port
			}
		}
	}

	return "8300"
}
