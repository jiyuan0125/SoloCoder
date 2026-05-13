package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type SeatType string

const (
	SeatTypeNormal  SeatType = "normal"
	SeatTypeQuiet   SeatType = "quiet"
	SeatTypeWindow  SeatType = "window"
)

type ReservationStatus string

const (
	StatusReserved ReservationStatus = "reserved"
	StatusChecked  ReservationStatus = "checked"
	StatusExpired  ReservationStatus = "expired"
	StatusCancelled ReservationStatus = "cancelled"
)

type Seat struct {
	ID   int64    `json:"id"`
	Code string   `json:"code"`
	Type SeatType `json:"type"`
}

type Reservation struct {
	ID             int64             `json:"id"`
	SeatID         int64             `json:"seat_id"`
	SeatCode       string            `json:"seat_code"`
	SeatType       SeatType          `json:"seat_type"`
	StudentID      string            `json:"student_id"`
	Date           string            `json:"date"`
	StartTime      string            `json:"start_time"`
	EndTime        string            `json:"end_time"`
	Status         ReservationStatus `json:"status"`
	ReservedAt     time.Time         `json:"reserved_at"`
	CheckedAt      *time.Time        `json:"checked_at,omitempty"`
	CancelledAt    *time.Time        `json:"cancelled_at,omitempty"`
}

type SeatStatus struct {
	Seat    Seat                `json:"seat"`
	Status  ReservationStatus   `json:"status"`
	IsFree  bool                `json:"is_free"`
}

type SeatStatusResponse struct {
	Seat
	Status ReservationStatus `json:"status"`
	IsFree bool              `json:"is_free"`
}

const (
	reservationWindowStart = 7
	cancelWindowMinutes    = 30
	checkInWindowMinutes   = 15
	defaultStartTime       = "08:00"
	defaultEndTime         = "22:00"
)

var (
	db      *sql.DB
	seatMux sync.RWMutex
)

func main() {
	initDB()
	defer db.Close()

	go startAutoReleaseWorker()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/seats", listSeats)
	mux.HandleFunc("/api/seats/", seatHandler)
	mux.HandleFunc("/api/reservations", reservationsHandler)
	mux.HandleFunc("/api/reservations/student/", studentReservationsHandler)
	mux.HandleFunc("/api/reservations/", reservationHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./roomrush.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTables()
	seedData()
}

func createTables() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS seats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS reservations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seat_id INTEGER NOT NULL,
			student_id TEXT NOT NULL,
			date TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'reserved',
			reserved_at DATETIME NOT NULL,
			checked_at DATETIME,
			cancelled_at DATETIME,
			FOREIGN KEY (seat_id) REFERENCES seats(id),
			UNIQUE(seat_id, date, status)
		);

		CREATE INDEX IF NOT EXISTS idx_reservations_student_date ON reservations(student_id, date);
		CREATE INDEX IF NOT EXISTS idx_reservations_seat_date ON reservations(seat_id, date);
		CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func seedData() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM seats").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count seats: %v", err)
	}

	if count > 0 {
		return
	}

	seats := []struct {
		code string
		typ  SeatType
	}{
		{"A01", SeatTypeNormal}, {"A02", SeatTypeNormal}, {"A03", SeatTypeNormal},
		{"A04", SeatTypeNormal}, {"A05", SeatTypeNormal},
		{"B01", SeatTypeQuiet}, {"B02", SeatTypeQuiet}, {"B03", SeatTypeQuiet},
		{"B04", SeatTypeQuiet}, {"B05", SeatTypeQuiet},
		{"C01", SeatTypeWindow}, {"C02", SeatTypeWindow}, {"C03", SeatTypeWindow},
		{"C04", SeatTypeWindow}, {"C05", SeatTypeWindow},
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO seats (code, type) VALUES (?, ?)")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, s := range seats {
		if _, err := stmt.Exec(s.code, s.typ); err != nil {
			log.Fatalf("Failed to insert seat: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Printf("Inserted %d seats", len(seats))
}

func listSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}

	rows, err := db.Query(`
		SELECT s.id, s.code, s.type
		FROM seats s
		ORDER BY s.code
	`)
	if err != nil {
		log.Printf("Failed to query seats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	seats := make([]Seat, 0)
	for rows.Next() {
		var s Seat
		if err := rows.Scan(&s.ID, &s.Code, &s.Type); err != nil {
			log.Printf("Failed to scan seat: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		seats = append(seats, s)
	}

	seatStatuses := make([]SeatStatusResponse, 0, len(seats))
	for _, seat := range seats {
		status, isFree := getSeatStatus(seat.ID, date)
		seatStatuses = append(seatStatuses, SeatStatusResponse{
			Seat:   seat,
			Status: status,
			IsFree: isFree,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"date":  date,
		"seats": seatStatuses,
	})
}

func seatHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/seats/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 || parts[0] == "" {
		listSeats(w, r)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid seat ID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 && parts[1] != "" {
		if parts[1] == "status" {
			seatStatusHandler(w, r, id)
		} else if r.Method == http.MethodPost {
			seatActionHandler(w, r, id, parts[1])
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		getSeatDetail(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func seatStatusHandler(w http.ResponseWriter, r *http.Request, seatID int64) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}

	var seat Seat
	err := db.QueryRow("SELECT id, code, type FROM seats WHERE id = ?", seatID).
		Scan(&seat.ID, &seat.Code, &seat.Type)
	if err == sql.ErrNoRows {
		http.Error(w, "Seat not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query seat: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	status, isFree := getSeatStatus(seat.ID, date)
	writeJSON(w, http.StatusOK, SeatStatusResponse{
		Seat:   seat,
		Status: status,
		IsFree: isFree,
	})
}

func getSeatDetail(w http.ResponseWriter, r *http.Request, seatID int64) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}

	var seat Seat
	err := db.QueryRow("SELECT id, code, type FROM seats WHERE id = ?", seatID).
		Scan(&seat.ID, &seat.Code, &seat.Type)
	if err == sql.ErrNoRows {
		http.Error(w, "Seat not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query seat: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	reservations, err := getSeatReservations(seatID, date)
	if err != nil {
		log.Printf("Failed to get seat reservations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	status, isFree := getSeatStatus(seatID, date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"seat":         seat,
		"date":         date,
		"status":       status,
		"is_free":      isFree,
		"reservations": reservations,
	})
}

func seatActionHandler(w http.ResponseWriter, r *http.Request, seatID int64, action string) {
	switch action {
	case "reserve":
		reserveSeat(w, r, seatID)
	default:
		http.Error(w, "Action not found", http.StatusNotFound)
	}
}

func reserveSeat(w http.ResponseWriter, r *http.Request, seatID int64) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StudentID string `json:"student_id"`
		Date      string `json:"date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.StudentID == "" || req.Date == "" {
		http.Error(w, "student_id and date are required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	tomorrow := now.Add(24 * time.Hour).Format("2006-01-02")
	if req.Date != tomorrow {
		http.Error(w, "Can only reserve for tomorrow", http.StatusBadRequest)
		return
	}

	allowedTime := time.Date(now.Year(), now.Month(), now.Day(), reservationWindowStart, 0, 0, 0, now.Location())
	if now.Before(allowedTime) {
		http.Error(w, "Reservation window opens at 7:00 AM", http.StatusForbidden)
		return
	}

	seatMux.Lock()
	defer seatMux.Unlock()

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var seat Seat
	err = tx.QueryRow("SELECT id, code, type FROM seats WHERE id = ?", seatID).
		Scan(&seat.ID, &seat.Code, &seat.Type)
	if err == sql.ErrNoRows {
		http.Error(w, "Seat not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query seat: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var studentCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM reservations
		WHERE student_id = ? AND date = ? AND status IN ('reserved', 'checked')
	`, req.StudentID, req.Date).Scan(&studentCount)
	if err != nil {
		log.Printf("Failed to check student reservations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if studentCount > 0 {
		http.Error(w, "Student already has a reservation for this date", http.StatusConflict)
		return
	}

	var seatCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM reservations
		WHERE seat_id = ? AND date = ? AND status IN ('reserved', 'checked')
	`, seatID, req.Date).Scan(&seatCount)
	if err != nil {
		log.Printf("Failed to check seat reservations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if seatCount > 0 {
		status, _ := getSeatStatus(seatID, req.Date)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Seat is already reserved",
			"seat_status": status,
		})
		return
	}

	reservation := Reservation{
		SeatID:     seatID,
		SeatCode:   seat.Code,
		SeatType:   seat.Type,
		StudentID:  req.StudentID,
		Date:       req.Date,
		StartTime:  defaultStartTime,
		EndTime:    defaultEndTime,
		Status:     StatusReserved,
		ReservedAt: now,
	}

	result, err := tx.Exec(`
		INSERT INTO reservations (seat_id, student_id, date, start_time, end_time, status, reserved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, seatID, req.StudentID, req.Date, defaultStartTime, defaultEndTime, StatusReserved, now)
	if err != nil {
		log.Printf("Failed to insert reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	reservation.ID, _ = result.LastInsertId()

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, reservation)
}

func reservationsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listReservations(w, r)
	case http.MethodPost:
		createReservation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func reservationHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/reservations/")
	parts := strings.SplitN(path, "/", 2)
	
	if len(parts) == 0 || parts[0] == "" {
		listReservations(w, r)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid reservation ID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 && parts[1] != "" {
		reservationActionHandler(w, r, id, parts[1])
		return
	}

	switch r.Method {
	case http.MethodGet:
		getReservation(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func reservationActionHandler(w http.ResponseWriter, r *http.Request, reservationID int64, action string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "cancel":
		cancelReservation(w, r, reservationID)
	case "check-in":
		checkInReservation(w, r, reservationID)
	default:
		http.Error(w, "Action not found", http.StatusNotFound)
	}
}

func studentReservationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	studentID := strings.TrimPrefix(r.URL.Path, "/api/reservations/student/")
	if studentID == "" {
		http.Error(w, "Student ID is required", http.StatusBadRequest)
		return
	}

	date := r.URL.Query().Get("date")
	
	query := `
		SELECT r.id, r.seat_id, s.code, s.type, r.student_id, r.date, 
		       r.start_time, r.end_time, r.status, r.reserved_at, r.checked_at, r.cancelled_at
		FROM reservations r
		JOIN seats s ON r.seat_id = s.id
		WHERE r.student_id = ?
	`
	args := []interface{}{studentID}
	
	if date != "" {
		query += " AND r.date = ?"
		args = append(args, date)
	}
	query += " ORDER BY r.date DESC, r.reserved_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Failed to query student reservations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	reservations := make([]Reservation, 0)
	for rows.Next() {
		var r Reservation
		var checkedAt, cancelledAt sql.NullTime
		err := rows.Scan(&r.ID, &r.SeatID, &r.SeatCode, &r.SeatType, &r.StudentID, &r.Date,
			&r.StartTime, &r.EndTime, &r.Status, &r.ReservedAt, &checkedAt, &cancelledAt)
		if err != nil {
			log.Printf("Failed to scan reservation: %v", err)
			continue
		}
		if checkedAt.Valid {
			r.CheckedAt = &checkedAt.Time
		}
		if cancelledAt.Valid {
			r.CancelledAt = &cancelledAt.Time
		}
		reservations = append(reservations, r)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"student_id":   studentID,
		"reservations": reservations,
	})
}

func listReservations(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	status := r.URL.Query().Get("status")

	query := `
		SELECT r.id, r.seat_id, s.code, s.type, r.student_id, r.date,
		       r.start_time, r.end_time, r.status, r.reserved_at, r.checked_at, r.cancelled_at
		FROM reservations r
		JOIN seats s ON r.seat_id = s.id
		WHERE 1=1
	`
	args := []interface{}{}

	if date != "" {
		query += " AND r.date = ?"
		args = append(args, date)
	}
	if status != "" {
		query += " AND r.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY r.date DESC, r.reserved_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Failed to query reservations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	reservations := make([]Reservation, 0)
	for rows.Next() {
		var r Reservation
		var checkedAt, cancelledAt sql.NullTime
		err := rows.Scan(&r.ID, &r.SeatID, &r.SeatCode, &r.SeatType, &r.StudentID, &r.Date,
			&r.StartTime, &r.EndTime, &r.Status, &r.ReservedAt, &checkedAt, &cancelledAt)
		if err != nil {
			log.Printf("Failed to scan reservation: %v", err)
			continue
		}
		if checkedAt.Valid {
			r.CheckedAt = &checkedAt.Time
		}
		if cancelledAt.Valid {
			r.CancelledAt = &cancelledAt.Time
		}
		reservations = append(reservations, r)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"reservations": reservations,
	})
}

func createReservation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SeatID    int64  `json:"seat_id"`
		StudentID string `json:"student_id"`
		Date      string `json:"date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SeatID == 0 || req.StudentID == "" || req.Date == "" {
		http.Error(w, "seat_id, student_id and date are required", http.StatusBadRequest)
		return
	}

	reserveSeat(w, r, req.SeatID)
}

func getReservation(w http.ResponseWriter, r *http.Request, reservationID int64) {
	var rsv Reservation
	var checkedAt, cancelledAt sql.NullTime
	
	err := db.QueryRow(`
		SELECT r.id, r.seat_id, s.code, s.type, r.student_id, r.date,
		       r.start_time, r.end_time, r.status, r.reserved_at, r.checked_at, r.cancelled_at
		FROM reservations r
		JOIN seats s ON r.seat_id = s.id
		WHERE r.id = ?
	`, reservationID).Scan(&rsv.ID, &rsv.SeatID, &rsv.SeatCode, &rsv.SeatType, &rsv.StudentID, &rsv.Date,
		&rsv.StartTime, &rsv.EndTime, &rsv.Status, &rsv.ReservedAt, &checkedAt, &cancelledAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Reservation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if checkedAt.Valid {
		rsv.CheckedAt = &checkedAt.Time
	}
	if cancelledAt.Valid {
		rsv.CancelledAt = &cancelledAt.Time
	}

	writeJSON(w, http.StatusOK, rsv)
}

func cancelReservation(w http.ResponseWriter, r *http.Request, reservationID int64) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	seatMux.Lock()
	defer seatMux.Unlock()

	now := time.Now()
	
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var rsv Reservation
	var checkedAt, cancelledAt sql.NullTime
	
	err = tx.QueryRow(`
		SELECT r.id, r.seat_id, s.code, s.type, r.student_id, r.date,
		       r.start_time, r.end_time, r.status, r.reserved_at, r.checked_at, r.cancelled_at
		FROM reservations r
		JOIN seats s ON r.seat_id = s.id
		WHERE r.id = ?
	`, reservationID).Scan(&rsv.ID, &rsv.SeatID, &rsv.SeatCode, &rsv.SeatType, &rsv.StudentID, &rsv.Date,
		&rsv.StartTime, &rsv.EndTime, &rsv.Status, &rsv.ReservedAt, &checkedAt, &cancelledAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Reservation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rsv.Status != StatusReserved {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Cannot cancel reservation in current state",
			"seat_status": rsv.Status,
		})
		return
	}

	cancelDeadline := rsv.ReservedAt.Add(time.Duration(cancelWindowMinutes) * time.Minute)
	if now.After(cancelDeadline) {
		http.Error(w, "Cancellation window expired (30 minutes after reservation)", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(`
		UPDATE reservations SET status = ?, cancelled_at = ? WHERE id = ?
	`, StatusCancelled, now, reservationID)
	if err != nil {
		log.Printf("Failed to cancel reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rsv.Status = StatusCancelled
	rsv.CancelledAt = &now

	writeJSON(w, http.StatusOK, rsv)
}

func checkInReservation(w http.ResponseWriter, r *http.Request, reservationID int64) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	seatMux.Lock()
	defer seatMux.Unlock()

	now := time.Now()
	
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var rsv Reservation
	var checkedAt, cancelledAt sql.NullTime
	
	err = tx.QueryRow(`
		SELECT r.id, r.seat_id, s.code, s.type, r.student_id, r.date,
		       r.start_time, r.end_time, r.status, r.reserved_at, r.checked_at, r.cancelled_at
		FROM reservations r
		JOIN seats s ON r.seat_id = s.id
		WHERE r.id = ?
	`, reservationID).Scan(&rsv.ID, &rsv.SeatID, &rsv.SeatCode, &rsv.SeatType, &rsv.StudentID, &rsv.Date,
		&rsv.StartTime, &rsv.EndTime, &rsv.Status, &rsv.ReservedAt, &checkedAt, &cancelledAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Reservation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rsv.Status != StatusReserved {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Cannot check in reservation in current state",
			"seat_status": rsv.Status,
		})
		return
	}

	startTime, err := time.Parse("15:04", rsv.StartTime)
	if err != nil {
		log.Printf("Failed to parse start time: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	reservationDate, _ := time.Parse("2006-01-02", rsv.Date)
	checkInStart := time.Date(reservationDate.Year(), reservationDate.Month(), reservationDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
	checkInDeadline := checkInStart.Add(time.Duration(checkInWindowMinutes) * time.Minute)

	if now.Before(checkInStart) {
		http.Error(w, "Check-in not available yet", http.StatusBadRequest)
		return
	}

	if now.After(checkInDeadline) {
		http.Error(w, "Check-in window expired (15 minutes after start time)", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(`
		UPDATE reservations SET status = ?, checked_at = ? WHERE id = ?
	`, StatusChecked, now, reservationID)
	if err != nil {
		log.Printf("Failed to check in reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rsv.Status = StatusChecked
	rsv.CheckedAt = &now

	writeJSON(w, http.StatusOK, rsv)
}

func getSeatStatus(seatID int64, date string) (ReservationStatus, bool) {
	var status ReservationStatus
	var reservedAt sql.NullTime
	var startTime string

	err := db.QueryRow(`
		SELECT r.status, r.reserved_at, r.start_time
		FROM reservations r
		WHERE r.seat_id = ? AND r.date = ? AND r.status IN ('reserved', 'checked')
		ORDER BY r.reserved_at DESC
		LIMIT 1
	`, seatID, date).Scan(&status, &reservedAt, &startTime)

	if err == sql.ErrNoRows {
		return "", true
	}
	if err != nil {
		log.Printf("Failed to query seat status: %v", err)
		return "", false
	}

	if status == StatusChecked {
		return StatusChecked, false
	}

	now := time.Now()
	reservationDate, _ := time.Parse("2006-01-02", date)
	startTimeParsed, _ := time.Parse("15:04", startTime)
	checkInDeadline := time.Date(reservationDate.Year(), reservationDate.Month(), reservationDate.Day(),
		startTimeParsed.Hour(), startTimeParsed.Minute(), 0, 0, now.Location()).
		Add(time.Duration(checkInWindowMinutes) * time.Minute)

	if now.After(checkInDeadline) {
		return StatusExpired, true
	}

	return StatusReserved, false
}

func getSeatReservations(seatID int64, date string) ([]Reservation, error) {
	rows, err := db.Query(`
		SELECT r.id, r.seat_id, s.code, s.type, r.student_id, r.date,
		       r.start_time, r.end_time, r.status, r.reserved_at, r.checked_at, r.cancelled_at
		FROM reservations r
		JOIN seats s ON r.seat_id = s.id
		WHERE r.seat_id = ? AND r.date = ?
		ORDER BY r.reserved_at DESC
	`, seatID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reservations := make([]Reservation, 0)
	for rows.Next() {
		var r Reservation
		var checkedAt, cancelledAt sql.NullTime
		err := rows.Scan(&r.ID, &r.SeatID, &r.SeatCode, &r.SeatType, &r.StudentID, &r.Date,
			&r.StartTime, &r.EndTime, &r.Status, &r.ReservedAt, &checkedAt, &cancelledAt)
		if err != nil {
			continue
		}
		if checkedAt.Valid {
			r.CheckedAt = &checkedAt.Time
		}
		if cancelledAt.Valid {
			r.CancelledAt = &cancelledAt.Time
		}
		reservations = append(reservations, r)
	}

	return reservations, nil
}

func startAutoReleaseWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		autoReleaseExpiredReservations()
	}
}

func autoReleaseExpiredReservations() {
	now := time.Now()
	today := now.Format("2006-01-02")

	seatMux.Lock()
	defer seatMux.Unlock()

	rows, err := db.Query(`
		SELECT r.id, r.start_time
		FROM reservations r
		WHERE r.status = 'reserved' AND r.date = ?
	`, today)
	if err != nil {
		log.Printf("Failed to query expired reservations: %v", err)
		return
	}
	defer rows.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback()

	for rows.Next() {
		var id int64
		var startTime string
		if err := rows.Scan(&id, &startTime); err != nil {
			continue
		}

		startTimeParsed, _ := time.Parse("15:04", startTime)
		checkInDeadline := time.Date(now.Year(), now.Month(), now.Day(),
			startTimeParsed.Hour(), startTimeParsed.Minute(), 0, 0, now.Location()).
			Add(time.Duration(checkInWindowMinutes) * time.Minute)

		if now.After(checkInDeadline) {
			_, err := tx.Exec(`
				UPDATE reservations SET status = ? WHERE id = ?
			`, StatusExpired, id)
			if err != nil {
				log.Printf("Failed to expire reservation %d: %v", id, err)
			} else {
				log.Printf("Auto-expired reservation %d", id)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit auto-release transaction: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
