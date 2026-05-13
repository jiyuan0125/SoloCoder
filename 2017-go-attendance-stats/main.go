package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Workdays   []int  `json:"workdays"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Database   string `json:"database"`
	Port       int    `json:"port"`
}

type Employee struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AttendanceRecord struct {
	ID           int       `json:"id"`
	EmployeeID   int       `json:"employee_id"`
	PunchTime    time.Time `json:"punch_time"`
	PunchType    string    `json:"punch_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type MakeupRequest struct {
	ID          int       `json:"id"`
	EmployeeID  int       `json:"employee_id"`
	Date        string    `json:"date"`
	PunchType   string    `json:"punch_type"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MonthlyStats struct {
	EmployeeID     int `json:"employee_id"`
	Year           int `json:"year"`
	Month          int `json:"month"`
	AttendanceDays int `json:"attendance_days"`
	LateCount      int `json:"late_count"`
	EarlyLeaveCount int `json:"early_leave_count"`
	OvertimeHours  int `json:"overtime_hours"`
}

var config Config

func main() {
	loadConfig()
	initDB(config.Database)
	defer db.Close()

	initTestData()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/employees", handleEmployees)
	mux.HandleFunc("/api/employees/", handleEmployeeDetail)

	mux.HandleFunc("/api/attendance", handleAttendance)
	mux.HandleFunc("/api/attendance/", handleAttendanceDetail)
	mux.HandleFunc("/api/attendance/punch", handlePunch)

	mux.HandleFunc("/api/makeup", handleMakeup)
	mux.HandleFunc("/api/makeup/", handleMakeupDetail)

	mux.HandleFunc("/api/stats/monthly", handleMonthlyStats)

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func loadConfig() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}
}

func isWorkday(t time.Time) bool {
	weekday := int(t.Weekday())
	for _, wd := range config.Workdays {
		if wd == weekday {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}
