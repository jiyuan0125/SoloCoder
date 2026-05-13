package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	dbPath               = "parking.db"
	freeMinutes          = 30
	ratePerHour          = 500
	dailyCap             = 4000
	memberDiscount       = 0.8
	dayStartHour         = 0
	dayStartMinute       = 0
)

type VehicleType string

const (
	NormalVehicle VehicleType = "normal"
	MemberVehicle VehicleType = "member"
	MonthlyVehicle VehicleType = "monthly"
)

type ApprovalStatus string

const (
	StatusPending   ApprovalStatus = "pending"
	StatusApplied   ApprovalStatus = "applied"
	StatusReviewed  ApprovalStatus = "reviewed"
	StatusRechecked ApprovalStatus = "rechecked"
	StatusApproved  ApprovalStatus = "approved"
	StatusCompleted ApprovalStatus = "completed"
)

type ResourceType string

const (
	ResourceLot      ResourceType = "parking_lot"
	ResourceVehicle  ResourceType = "vehicle"
	ResourceMembership ResourceType = "membership"
)

type ParkingRecord struct {
	ID           int64     `json:"id"`
	PlateNumber  string    `json:"plate_number"`
	EntryTime    time.Time `json:"entry_time"`
	ExitTime     *time.Time `json:"exit_time,omitempty"`
	Fee          int64     `json:"fee,omitempty"`
	VehicleType  VehicleType `json:"vehicle_type"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   int64     `json:"resource_id"`
	HasFreePeriod bool     `json:"has_free_period"`
}

type Approval struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Status        ApprovalStatus `json:"status"`
	ResourceType  ResourceType   `json:"resource_type"`
	ResourceID    int64          `json:"resource_id"`
	Applicant     string         `json:"applicant"`
	FirstReviewer string         `json:"first_reviewer,omitempty"`
	SecondReviewer string        `json:"second_reviewer,omitempty"`
	Approver      string         `json:"approver,omitempty"`
	ApplicantTime *time.Time     `json:"applicant_time,omitempty"`
	FirstReviewTime *time.Time   `json:"first_review_time,omitempty"`
	SecondReviewTime *time.Time  `json:"second_review_time,omitempty"`
	ApprovalTime  *time.Time     `json:"approval_time,omitempty"`
	CompleteTime  *time.Time     `json:"complete_time,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

type Vehicle struct {
	PlateNumber string      `json:"plate_number"`
	VehicleType VehicleType `json:"vehicle_type"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ResourceSummary struct {
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   int64        `json:"resource_id"`
	TotalRecords int64        `json:"total_records"`
	TotalFee     int64        `json:"total_fee"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		panic(fmt.Sprintf("无法打开数据库: %v", err))
	}
	defer db.Close()

	err = initDB()
	if err != nil {
		panic(fmt.Sprintf("初始化数据库失败: %v", err))
	}

	http.HandleFunc("/vehicle", vehicleHandler)
	http.HandleFunc("/vehicle/", vehicleDetailHandler)
	http.HandleFunc("/parking/entry", entryHandler)
	http.HandleFunc("/parking/exit", exitHandler)
	http.HandleFunc("/parking/record/", recordDetailHandler)
	http.HandleFunc("/parking/history/", historyHandler)
	http.HandleFunc("/approval", approvalHandler)
	http.HandleFunc("/approval/", approvalDetailHandler)
	http.HandleFunc("/approval/advance/", approvalAdvanceHandler)
	http.HandleFunc("/resource/summary", resourceSummaryHandler)
	http.HandleFunc("/vehicles", vehiclesHandler)

	fmt.Println("停车场计费系统启动，监听端口 8102")
	err = http.ListenAndServe(":8102", nil)
	if err != nil {
		panic(fmt.Sprintf("启动服务器失败: %v", err))
	}
}

func initDB() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS vehicles (
			plate_number TEXT PRIMARY KEY,
			vehicle_type TEXT NOT NULL DEFAULT 'normal',
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS parking_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plate_number TEXT NOT NULL,
			entry_time TIMESTAMP NOT NULL,
			exit_time TIMESTAMP,
			fee INTEGER DEFAULT 0,
			vehicle_type TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id INTEGER NOT NULL DEFAULT 1,
			has_free_period INTEGER NOT NULL DEFAULT 1,
			FOREIGN KEY (plate_number) REFERENCES vehicles(plate_number)
		)`,
		`CREATE TABLE IF NOT EXISTS approvals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			resource_type TEXT NOT NULL,
			resource_id INTEGER NOT NULL DEFAULT 1,
			applicant TEXT NOT NULL,
			first_reviewer TEXT,
			second_reviewer TEXT,
			approver TEXT,
			applicant_time TIMESTAMP,
			first_review_time TIMESTAMP,
			second_review_time TIMESTAMP,
			approval_time TIMESTAMP,
			complete_time TIMESTAMP,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_parking_plate ON parking_records(plate_number)`,
		`CREATE INDEX IF NOT EXISTS idx_parking_resource ON parking_records(resource_type, resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_resource ON approvals(resource_type, resource_id)`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

type EntryRequest struct {
	PlateNumber  string       `json:"plate_number"`
	EntryTime    string       `json:"entry_time"`
	VehicleType  VehicleType  `json:"vehicle_type"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   int64        `json:"resource_id"`
}

type ExitRequest struct {
	PlateNumber string `json:"plate_number"`
	ExitTime    string `json:"exit_time"`
}

func entryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	var req EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if req.PlateNumber == "" {
		respondError(w, http.StatusBadRequest, "车牌号不能为空")
		return
	}

	vehicleType := req.VehicleType
	if vehicleType == "" {
		vehicleType = NormalVehicle
	}

	resourceType := req.ResourceType
	if resourceType == "" {
		resourceType = ResourceLot
	}

	resourceID := req.ResourceID
	if resourceID == 0 {
		resourceID = 1
	}

	var entryTime time.Time
	var err error
	if req.EntryTime != "" {
		entryTime, err = time.ParseInLocation("2006-01-02 15:04:05", req.EntryTime, time.Local)
		if err != nil {
			entryTime, err = time.Parse(time.RFC3339, req.EntryTime)
			if err != nil {
				respondError(w, http.StatusBadRequest, "无效的时间格式")
				return
			}
		}
	} else {
		entryTime = time.Now()
	}

	tx, err := db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT OR IGNORE INTO vehicles (plate_number, vehicle_type, created_at) VALUES (?, ?, ?)`,
		req.PlateNumber, vehicleType, time.Now())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	if vehicleType != NormalVehicle {
		_, err = tx.Exec(`UPDATE vehicles SET vehicle_type = ? WHERE plate_number = ?`, vehicleType, req.PlateNumber)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "服务器错误")
			return
		}
	}

	var existingCount int
	startOfDay := time.Date(entryTime.Year(), entryTime.Month(), entryTime.Day(), dayStartHour, dayStartMinute, 0, 0, entryTime.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err = tx.QueryRow(`SELECT COUNT(*) FROM parking_records WHERE plate_number = ? AND entry_time >= ? AND entry_time < ? AND exit_time IS NOT NULL`,
		req.PlateNumber, startOfDay, endOfDay).Scan(&existingCount)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	hasFreePeriod := existingCount == 0

	result, err := tx.Exec(`INSERT INTO parking_records (plate_number, entry_time, vehicle_type, resource_type, resource_id, has_free_period) VALUES (?, ?, ?, ?, ?, ?)`,
		req.PlateNumber, entryTime, vehicleType, resourceType, resourceID, hasFreePeriod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	id, _ := result.LastInsertId()
	err = tx.Commit()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	record := ParkingRecord{
		ID:           id,
		PlateNumber:  req.PlateNumber,
		EntryTime:    entryTime,
		VehicleType:  vehicleType,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		HasFreePeriod: hasFreePeriod,
	}

	respondJSON(w, http.StatusCreated, record)
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	var req ExitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if req.PlateNumber == "" {
		respondError(w, http.StatusBadRequest, "车牌号不能为空")
		return
	}

	var exitTime time.Time
	var err error
	if req.ExitTime != "" {
		exitTime, err = time.ParseInLocation("2006-01-02 15:04:05", req.ExitTime, time.Local)
		if err != nil {
			exitTime, err = time.Parse(time.RFC3339, req.ExitTime)
			if err != nil {
				respondError(w, http.StatusBadRequest, "无效的时间格式")
				return
			}
		}
	} else {
		exitTime = time.Now()
	}

	tx, err := db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer tx.Rollback()

	var record ParkingRecord
	var entryTimeStr string
	var hasFreePeriodInt int
	err = tx.QueryRow(`SELECT id, plate_number, entry_time, vehicle_type, resource_type, resource_id, has_free_period 
		FROM parking_records WHERE plate_number = ? AND exit_time IS NULL ORDER BY entry_time DESC LIMIT 1`,
		req.PlateNumber).Scan(&record.ID, &record.PlateNumber, &entryTimeStr, &record.VehicleType,
		&record.ResourceType, &record.ResourceID, &hasFreePeriodInt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusBadRequest, "未找到入场记录")
			return
		}
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	record.EntryTime, err = time.ParseInLocation("2006-01-02 15:04:05", entryTimeStr, time.Local)
	if err != nil {
		record.EntryTime, err = time.Parse(time.RFC3339, entryTimeStr)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "服务器错误")
			return
		}
	}

	record.HasFreePeriod = hasFreePeriodInt == 1

	if exitTime.Before(record.EntryTime) {
		respondError(w, http.StatusBadRequest, "出场时间不能早于入场时间")
		return
	}

	var vehicleType VehicleType
	err = tx.QueryRow(`SELECT vehicle_type FROM vehicles WHERE plate_number = ?`, req.PlateNumber).Scan(&vehicleType)
	if err == nil {
		record.VehicleType = vehicleType
	}

	fee := calculateFee(record.EntryTime, exitTime, record.VehicleType, record.HasFreePeriod)

	record.ExitTime = &exitTime
	record.Fee = fee

	_, err = tx.Exec(`UPDATE parking_records SET exit_time = ?, fee = ? WHERE id = ?`,
		exitTime, fee, record.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	err = tx.Commit()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	respondJSON(w, http.StatusOK, record)
}

func calculateFee(entryTime, exitTime time.Time, vehicleType VehicleType, hasFreePeriod bool) int64 {
	if vehicleType == MonthlyVehicle {
		return 0
	}

	duration := exitTime.Sub(entryTime)
	minutes := int64(duration.Minutes())

	if hasFreePeriod && minutes <= freeMinutes {
		return 0
	}

	effectiveMinutes := minutes
	if hasFreePeriod {
		effectiveMinutes = minutes - freeMinutes
	}

	if effectiveMinutes <= 0 {
		return 0
	}

	hours := (effectiveMinutes + 59) / 60

	var totalFee int64 = 0
	remainingHours := hours

	currentDay := time.Date(entryTime.Year(), entryTime.Month(), entryTime.Day(), dayStartHour, dayStartMinute, 0, 0, entryTime.Location())
	endDay := time.Date(exitTime.Year(), exitTime.Month(), exitTime.Day(), dayStartHour, dayStartMinute, 0, 0, exitTime.Location())

	for remainingHours > 0 {
		dayEnd := currentDay.Add(24 * time.Hour)
		if currentDay.Equal(endDay) {
			dayFee := int64(min(remainingHours*ratePerHour, dailyCap))
			totalFee += dayFee
			remainingHours = 0
		} else {
			entryTimeOnDay := max(entryTime, currentDay)
			remainingMinutesInDay := int64(dayEnd.Sub(entryTimeOnDay).Minutes())
			hoursInDay := (remainingMinutesInDay + 59) / 60
			if hoursInDay > remainingHours {
				hoursInDay = remainingHours
			}
			dayFee := int64(min(hoursInDay*ratePerHour, dailyCap))
			totalFee += dayFee
			remainingHours -= hoursInDay
			currentDay = currentDay.Add(24 * time.Hour)
		}
	}

	if vehicleType == MemberVehicle {
		totalDays := int64(endDay.Sub(currentDay).Hours()/24) + 1
		if endDay.After(currentDay) {
			totalDays = int64(endDay.Sub(entryTime).Hours()/24) + 1
		}
		if totalDays < 1 {
			totalDays = 1
		}

		cappedDays := hours / 8
		if hours%8 > 0 {
			cappedDays++
		}
		if cappedDays < 1 {
			cappedDays = 1
		}

		regularFee := hours * ratePerHour

		if regularFee > dailyCap*cappedDays {
			capAmount := dailyCap * cappedDays
			nonCapAmount := regularFee - (dailyCap * (cappedDays - 1))
			discountedCap := dailyCap * (cappedDays - 1)
			discountedNonCap := int64(float64(nonCapAmount) * memberDiscount)
			totalFee = discountedCap + discountedNonCap
			if totalFee > capAmount {
				totalFee = capAmount
			}
		} else {
			totalFee = int64(float64(regularFee) * memberDiscount)
		}
	}

	return totalFee
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}
	plateNumber := parts[3]

	rows, err := db.Query(`SELECT id, plate_number, entry_time, exit_time, fee, vehicle_type, resource_type, resource_id, has_free_period 
		FROM parking_records WHERE plate_number = ? ORDER BY entry_time DESC`, plateNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer rows.Close()

	records := make([]ParkingRecord, 0)
	var totalFee int64 = 0

	for rows.Next() {
		var record ParkingRecord
		var entryTimeStr, exitTimeStr sql.NullString
		var fee sql.NullInt64
		var hasFreePeriodInt int

		err := rows.Scan(&record.ID, &record.PlateNumber, &entryTimeStr, &exitTimeStr,
			&fee, &record.VehicleType, &record.ResourceType, &record.ResourceID, &hasFreePeriodInt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "服务器错误")
			return
		}

		if entryTimeStr.Valid {
			record.EntryTime, err = time.ParseInLocation("2006-01-02 15:04:05", entryTimeStr.String, time.Local)
			if err != nil {
				record.EntryTime, _ = time.Parse(time.RFC3339, entryTimeStr.String)
			}
		}

		if exitTimeStr.Valid {
			exitTime, err := time.ParseInLocation("2006-01-02 15:04:05", exitTimeStr.String, time.Local)
			if err != nil {
				exitTime, _ = time.Parse(time.RFC3339, exitTimeStr.String)
			}
			record.ExitTime = &exitTime
		}

		if fee.Valid {
			record.Fee = fee.Int64
			totalFee += fee.Int64
		}

		record.HasFreePeriod = hasFreePeriodInt == 1
		records = append(records, record)
	}

	if len(records) == 0 {
		records = make([]ParkingRecord, 0)
	}

	response := map[string]interface{}{
		"records":    records,
		"total_fee":  totalFee,
	}

	respondJSON(w, http.StatusOK, response)
}

func recordDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的记录ID")
		return
	}

	var record ParkingRecord
	var entryTimeStr, exitTimeStr sql.NullString
	var fee sql.NullInt64
	var hasFreePeriodInt int

	err = db.QueryRow(`SELECT id, plate_number, entry_time, exit_time, fee, vehicle_type, resource_type, resource_id, has_free_period 
		FROM parking_records WHERE id = ?`, id).Scan(&record.ID, &record.PlateNumber, &entryTimeStr, &exitTimeStr,
		&fee, &record.VehicleType, &record.ResourceType, &record.ResourceID, &hasFreePeriodInt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "记录不存在")
			return
		}
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	if entryTimeStr.Valid {
		record.EntryTime, err = time.ParseInLocation("2006-01-02 15:04:05", entryTimeStr.String, time.Local)
		if err != nil {
			record.EntryTime, _ = time.Parse(time.RFC3339, entryTimeStr.String)
		}
	}

	if exitTimeStr.Valid {
		exitTime, err := time.ParseInLocation("2006-01-02 15:04:05", exitTimeStr.String, time.Local)
		if err != nil {
			exitTime, _ = time.Parse(time.RFC3339, exitTimeStr.String)
		}
		record.ExitTime = &exitTime
	}

	if fee.Valid {
		record.Fee = fee.Int64
	}

	record.HasFreePeriod = hasFreePeriodInt == 1
	respondJSON(w, http.StatusOK, record)
}

func vehicleHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createVehicle(w, r)
	case http.MethodGet:
		listVehicles(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func createVehicle(w http.ResponseWriter, r *http.Request) {
	var vehicle Vehicle
	if err := json.NewDecoder(r.Body).Decode(&vehicle); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if vehicle.PlateNumber == "" {
		respondError(w, http.StatusBadRequest, "车牌号不能为空")
		return
	}

	if vehicle.VehicleType == "" {
		vehicle.VehicleType = NormalVehicle
	}

	now := time.Now()
	vehicle.CreatedAt = now

	_, err := db.Exec(`INSERT OR REPLACE INTO vehicles (plate_number, vehicle_type, created_at) VALUES (?, ?, ?)`,
		vehicle.PlateNumber, vehicle.VehicleType, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	respondJSON(w, http.StatusCreated, vehicle)
}

func listVehicles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT plate_number, vehicle_type, created_at FROM vehicles ORDER BY created_at DESC`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer rows.Close()

	vehicles := make([]Vehicle, 0)
	for rows.Next() {
		var vehicle Vehicle
		var createdAtStr string
		err := rows.Scan(&vehicle.PlateNumber, &vehicle.VehicleType, &createdAtStr)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "服务器错误")
			return
		}

		vehicle.CreatedAt, err = time.ParseInLocation("2006-01-02 15:04:05", createdAtStr, time.Local)
		if err != nil {
			vehicle.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		}

		vehicles = append(vehicles, vehicle)
	}

	if len(vehicles) == 0 {
		vehicles = make([]Vehicle, 0)
	}

	respondJSON(w, http.StatusOK, vehicles)
}

func vehiclesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}
	listVehicles(w, r)
}

func vehicleDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}
	plateNumber := parts[3]

	var vehicle Vehicle
	var createdAtStr string
	err := db.QueryRow(`SELECT plate_number, vehicle_type, created_at FROM vehicles WHERE plate_number = ?`, plateNumber).
		Scan(&vehicle.PlateNumber, &vehicle.VehicleType, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "车辆不存在")
			return
		}
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	vehicle.CreatedAt, err = time.ParseInLocation("2006-01-02 15:04:05", createdAtStr, time.Local)
	if err != nil {
		vehicle.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	}

	respondJSON(w, http.StatusOK, vehicle)
}

type ApprovalRequest struct {
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   int64        `json:"resource_id"`
	Applicant    string       `json:"applicant"`
}

func approvalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createApproval(w, r)
	case http.MethodGet:
		listApprovals(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func createApproval(w http.ResponseWriter, r *http.Request) {
	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "标题不能为空")
		return
	}

	if req.Applicant == "" {
		respondError(w, http.StatusBadRequest, "申请人不能为空")
		return
	}

	resourceType := req.ResourceType
	if resourceType == "" {
		resourceType = ResourceLot
	}

	resourceID := req.ResourceID
	if resourceID == 0 {
		resourceID = 1
	}

	now := time.Now()
	result, err := db.Exec(`INSERT INTO approvals (title, description, status, resource_type, resource_id, applicant, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`, req.Title, req.Description, StatusPending, resourceType, resourceID, req.Applicant, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	id, _ := result.LastInsertId()
	approval := Approval{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      StatusPending,
		ResourceType: resourceType,
		ResourceID:  resourceID,
		Applicant:   req.Applicant,
		CreatedAt:   now,
	}

	respondJSON(w, http.StatusCreated, approval)
}

func listApprovals(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, title, description, status, resource_type, resource_id, applicant, 
		first_reviewer, second_reviewer, approver, applicant_time, first_review_time, 
		second_review_time, approval_time, complete_time, created_at FROM approvals`

	var args []interface{}
	resourceType := r.URL.Query().Get("resource_type")
	resourceIDStr := r.URL.Query().Get("resource_id")
	status := r.URL.Query().Get("status")

	whereClauses := make([]string, 0)
	if resourceType != "" {
		whereClauses = append(whereClauses, "resource_type = ?")
		args = append(args, resourceType)
	}
	if resourceIDStr != "" {
		whereClauses = append(whereClauses, "resource_id = ?")
		args = append(args, resourceIDStr)
	}
	if status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer rows.Close()

	approvals := make([]Approval, 0)
	for rows.Next() {
		var approval Approval
		var applicantTime, firstReviewTime, secondReviewTime, approvalTime, completeTime, createdAtStr sql.NullString

		err := rows.Scan(&approval.ID, &approval.Title, &approval.Description, &approval.Status,
			&approval.ResourceType, &approval.ResourceID, &approval.Applicant,
			&approval.FirstReviewer, &approval.SecondReviewer, &approval.Approver,
			&applicantTime, &firstReviewTime, &secondReviewTime, &approvalTime, &completeTime, &createdAtStr)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "服务器错误")
			return
		}

		approval.ApplicantTime = parseNullableTime(applicantTime)
		approval.FirstReviewTime = parseNullableTime(firstReviewTime)
		approval.SecondReviewTime = parseNullableTime(secondReviewTime)
		approval.ApprovalTime = parseNullableTime(approvalTime)
		approval.CompleteTime = parseNullableTime(completeTime)
		approval.CreatedAt = parseTime(createdAtStr.String)

		approvals = append(approvals, approval)
	}

	if len(approvals) == 0 {
		approvals = make([]Approval, 0)
	}

	respondJSON(w, http.StatusOK, approvals)
}

func parseNullableTime(ns sql.NullString) *time.Time {
	if !ns.Valid {
		return nil
	}
	t := parseTime(ns.String)
	return &t
}

func parseTime(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}

func approvalDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的审批ID")
		return
	}

	var approval Approval
	var applicantTime, firstReviewTime, secondReviewTime, approvalTime, completeTime, createdAtStr sql.NullString

	err = db.QueryRow(`SELECT id, title, description, status, resource_type, resource_id, applicant, 
		first_reviewer, second_reviewer, approver, applicant_time, first_review_time, 
		second_review_time, approval_time, complete_time, created_at FROM approvals WHERE id = ?`, id).
		Scan(&approval.ID, &approval.Title, &approval.Description, &approval.Status,
			&approval.ResourceType, &approval.ResourceID, &approval.Applicant,
			&approval.FirstReviewer, &approval.SecondReviewer, &approval.Approver,
			&applicantTime, &firstReviewTime, &secondReviewTime, &approvalTime, &completeTime, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "审批不存在")
			return
		}
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	approval.ApplicantTime = parseNullableTime(applicantTime)
	approval.FirstReviewTime = parseNullableTime(firstReviewTime)
	approval.SecondReviewTime = parseNullableTime(secondReviewTime)
	approval.ApprovalTime = parseNullableTime(approvalTime)
	approval.CompleteTime = parseNullableTime(completeTime)
	approval.CreatedAt = parseTime(createdAtStr.String)

	respondJSON(w, http.StatusOK, approval)
}

type AdvanceApprovalRequest struct {
	Reviewer string `json:"reviewer"`
}

func approvalAdvanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的审批ID")
		return
	}

	var req AdvanceApprovalRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "无效的请求参数")
			return
		}
	}

	tx, err := db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer tx.Rollback()

	var currentStatus ApprovalStatus
	err = tx.QueryRow(`SELECT status FROM approvals WHERE id = ?`, id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "审批不存在")
			return
		}
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	var nextStatus ApprovalStatus
	var updateField string
	var reviewerField string

	switch currentStatus {
	case StatusPending:
		nextStatus = StatusApplied
		updateField = "applicant_time"
	case StatusApplied:
		nextStatus = StatusReviewed
		updateField = "first_review_time"
		reviewerField = "first_reviewer"
	case StatusReviewed:
		nextStatus = StatusRechecked
		updateField = "second_review_time"
		reviewerField = "second_reviewer"
	case StatusRechecked:
		nextStatus = StatusApproved
		updateField = "approval_time"
		reviewerField = "approver"
	case StatusApproved:
		nextStatus = StatusCompleted
		updateField = "complete_time"
	case StatusCompleted:
		respondError(w, http.StatusBadRequest, "审批已完成，无法继续推进")
		return
	default:
		respondError(w, http.StatusBadRequest, "未知的审批状态")
		return
	}

	now := time.Now()
	var result sql.Result

	if reviewerField != "" && req.Reviewer != "" {
		result, err = tx.Exec(`UPDATE approvals SET status = ?, `+updateField+` = ?, `+reviewerField+` = ? WHERE id = ?`,
			nextStatus, now, req.Reviewer, id)
	} else {
		result, err = tx.Exec(`UPDATE approvals SET status = ?, `+updateField+` = ? WHERE id = ?`,
			nextStatus, now, id)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		respondError(w, http.StatusNotFound, "审批不存在")
		return
	}

	err = tx.Commit()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": nextStatus,
	})
}

func resourceSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	resourceType := r.URL.Query().Get("resource_type")
	resourceIDStr := r.URL.Query().Get("resource_id")

	query := `SELECT resource_type, resource_id, COUNT(*) as total_records, COALESCE(SUM(fee), 0) as total_fee 
		FROM parking_records WHERE exit_time IS NOT NULL`

	var args []interface{}
	whereClauses := make([]string, 0)

	if resourceType != "" {
		whereClauses = append(whereClauses, "resource_type = ?")
		args = append(args, resourceType)
	}
	if resourceIDStr != "" {
		whereClauses = append(whereClauses, "resource_id = ?")
		args = append(args, resourceIDStr)
	}

	if len(whereClauses) > 0 {
		query += " AND " + strings.Join(whereClauses, " AND ")
	}

	query += " GROUP BY resource_type, resource_id"

	rows, err := db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	defer rows.Close()

	summaries := make([]ResourceSummary, 0)
	for rows.Next() {
		var summary ResourceSummary
		err := rows.Scan(&summary.ResourceType, &summary.ResourceID, &summary.TotalRecords, &summary.TotalFee)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "服务器错误")
			return
		}
		summaries = append(summaries, summary)
	}

	if len(summaries) == 0 {
		summaries = make([]ResourceSummary, 0)
	}

	respondJSON(w, http.StatusOK, summaries)
}
