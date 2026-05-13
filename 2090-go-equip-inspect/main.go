package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	db *Database
}

func NewServer(db *Database) *Server {
	return &Server{db: db}
}

func (s *Server) getNextRunDate(currentDate time.Time, frequency string) time.Time {
	switch frequency {
	case ScheduleFrequencyDaily:
		return currentDate.AddDate(0, 0, 1)
	case ScheduleFrequencyWeekly:
		return currentDate.AddDate(0, 0, 7)
	case ScheduleFrequencyMonthly:
		return currentDate.AddDate(0, 1, 0)
	default:
		return currentDate.AddDate(0, 0, 1)
	}
}

func (s *Server) RunScheduledTasks() error {
	now := time.Now()
	today := now.Format("2006-01-02")

	tx, err := s.db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT id, route_id, frequency, next_run_date FROM schedules
		WHERE active = 1 AND date(next_run_date) <= date(?)
	`, today)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Schedule struct {
		ID          int64
		RouteID     int64
		Frequency   string
		NextRunDate string
	}

	var schedules []Schedule
	for rows.Next() {
		var sch Schedule
		if err := rows.Scan(&sch.ID, &sch.RouteID, &sch.Frequency, &sch.NextRunDate); err != nil {
			continue
		}
		schedules = append(schedules, sch)
	}

	for _, sch := range schedules {
		nextRun, err := time.Parse("2006-01-02", sch.NextRunDate)
		if err != nil {
			continue
		}

		var dueDate time.Time
		switch sch.Frequency {
		case ScheduleFrequencyDaily:
			dueDate = nextRun.Add(24 * time.Hour)
		case ScheduleFrequencyWeekly:
			dueDate = nextRun.Add(7 * 24 * time.Hour)
		case ScheduleFrequencyMonthly:
			dueDate = nextRun.AddDate(0, 1, 0)
		}

		var existingCount int
		tx.QueryRow(`
			SELECT COUNT(*) FROM tasks
			WHERE schedule_id = ? AND date(created_at) = date(?)
		`, sch.ID, nextRun.Format("2006-01-02")).Scan(&existingCount)

		if existingCount == 0 {
			_, err = tx.Exec(`
				INSERT INTO tasks (schedule_id, route_id, status, due_date)
				VALUES (?, ?, ?, ?)
			`, sch.ID, sch.RouteID, TaskStatusPending, dueDate)
			if err != nil {
				log.Printf("Failed to create task for schedule %d: %v", sch.ID, err)
				continue
			}
			logOperation(fmt.Sprintf("Auto-generated task for schedule %d", sch.ID))
		}

		newNextRun := s.getNextRunDate(nextRun, sch.Frequency)
		_, err = tx.Exec(`
			UPDATE schedules SET next_run_date = ? WHERE id = ?
		`, newNextRun.Format("2006-01-02"), sch.ID)
		if err != nil {
			log.Printf("Failed to update schedule %d: %v", sch.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	s.db.UpdateStatistics()
	return nil
}

func (s *Server) CheckOverdueTasks() ([]map[string]interface{}, error) {
	now := time.Now()

	rows, err := s.db.db.Query(`
		SELECT id, route_id, status, inspector_id, due_date
		FROM tasks
		WHERE status NOT IN ('closed', 'review_pass')
		AND due_date IS NOT NULL
		AND datetime(due_date) < datetime(?)
		AND reminder_sent = 0
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []map[string]interface{}
	var taskIDs []int64

	for rows.Next() {
		var id int64
		var routeID int64
		var status string
		var inspectorID sql.NullInt64
		var dueDate string

		if err := rows.Scan(&id, &routeID, &status, &inspectorID, &dueDate); err != nil {
			continue
		}

		reminder := map[string]interface{}{
			"task_id":     id,
			"route_id":    routeID,
			"status":      status,
			"due_date":    dueDate,
			"overdue_since": now.Format(time.RFC3339),
		}
		if inspectorID.Valid {
			reminder["inspector_id"] = inspectorID.Int64
		}

		reminders = append(reminders, reminder)
		taskIDs = append(taskIDs, id)
	}

	for _, id := range taskIDs {
		s.db.db.Exec(`UPDATE tasks SET reminder_sent = 1 WHERE id = ?`, id)
		logOperation(fmt.Sprintf("Sent reminder for overdue task %d", id))
	}

	return reminders, nil
}

func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateDevice(w, r)
		} else if r.Method == http.MethodGet {
			s.ListDevices(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/inspection-points", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateInspectionPoint(w, r)
		} else if r.Method == http.MethodGet {
			s.ListInspectionPoints(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/inspection-items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateInspectionItem(w, r)
		} else if r.Method == http.MethodGet {
			s.ListInspectionItems(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateRoute(w, r)
		} else if r.Method == http.MethodGet {
			s.ListRoutes(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateUser(w, r)
		} else if r.Method == http.MethodGet {
			s.ListUsers(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/schedules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CreateSchedule(w, r)
		} else if r.Method == http.MethodGet {
			s.ListSchedules(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/schedules/quota", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			s.UpdateScheduleQuota(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.ListTasks(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/claim", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.ClaimTask(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/checkin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.Checkin(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.SubmitTask(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/transition", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.TransitionTask(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.GetTask(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/maintenance/assign", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.AssignMaintenance(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/maintenance/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CompleteMaintenance(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/maintenance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.ListMaintenanceRecords(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/review", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.ReviewTask(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.ListReviewRecords(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.GetStatistics(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/scheduler/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.RunScheduler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/reminders/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.CheckReminders(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	return mux
}

func (s *Server) StartBackgroundScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.RunScheduledTasks(); err != nil {
			log.Printf("Scheduler error: %v", err)
		}
		if _, err := s.CheckOverdueTasks(); err != nil {
			log.Printf("Reminder check error: %v", err)
		}
	}
}

func main() {
	db, err := NewDatabase("./equipment_inspect.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)

	go server.StartBackgroundScheduler()

	handler := server.SetupRoutes()

	addr := ":8080"
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
