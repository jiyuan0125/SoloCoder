package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"education-platform/database"
	"education-platform/models"
	"education-platform/utils"

	"github.com/gorilla/mux"
)

func RegisterStudent(w http.ResponseWriter, r *http.Request) {
	var student models.Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if student.Nickname == "" || student.Phone == "" || student.Grade == "" {
		utils.JSONError(w, http.StatusBadRequest, "invalid_student", "Nickname, phone, and grade are required")
		return
	}

	validGrade := false
	for _, g := range models.AllGrades {
		if g == student.Grade {
			validGrade = true
			break
		}
	}
	if !validGrade {
		utils.JSONError(w, http.StatusBadRequest, "invalid_grade", "Invalid grade")
		return
	}

	result, err := database.DB.Exec(`INSERT INTO students (nickname, phone, grade, balance) VALUES (?, ?, ?, 0)`,
		student.Nickname, student.Phone, student.Grade)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	id, _ := result.LastInsertId()
	student.ID = id
	student.Balance = 0

	utils.JSONResponse(w, http.StatusCreated, student)
}

func ListStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`SELECT id, nickname, phone, grade, balance, created_at FROM students`)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	students := []models.Student{}
	for rows.Next() {
		var s models.Student
		err := rows.Scan(&s.ID, &s.Nickname, &s.Phone, &s.Grade, &s.Balance, &s.CreatedAt)
		if err != nil {
			continue
		}
		students = append(students, s)
	}

	utils.JSONResponse(w, http.StatusOK, students)
}

func GetStudent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid student ID")
		return
	}

	var s models.Student
	err = database.DB.QueryRow(`SELECT id, nickname, phone, grade, balance, created_at FROM students WHERE id = ?`, id).Scan(
		&s.ID, &s.Nickname, &s.Phone, &s.Grade, &s.Balance, &s.CreatedAt)

	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "not_found", "Student not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, s)
}

func RechargeCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StudentID  int64  `json:"student_id"`
		CardNumber string `json:"card_number"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer tx.Rollback()

	var card models.RechargeCard
	var usedBy sql.NullInt64
	var usedAt sql.NullString
	err = tx.QueryRow(`SELECT id, card_number, password, amount, is_used, used_by, used_at, purchase_at, expire_at 
		FROM recharge_cards WHERE card_number = ?`, req.CardNumber).Scan(
		&card.ID, &card.CardNumber, &card.Password, &card.Amount, &card.IsUsed, &usedBy, &usedAt, &card.PurchaseAt, &card.ExpireAt)

	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "not_found", "Recharge card not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if card.Password != req.Password {
		utils.JSONError(w, http.StatusBadRequest, "invalid_password", "Invalid card password")
		return
	}

	if card.IsUsed {
		utils.JSONError(w, http.StatusConflict, "card_used", "Card has already been used")
		return
	}

	if time.Now().After(card.ExpireAt) {
		utils.JSONError(w, http.StatusBadRequest, "card_expired", "Card has expired")
		return
	}

	var studentExists bool
	tx.QueryRow("SELECT 1 FROM students WHERE id = ?", req.StudentID).Scan(&studentExists)
	if !studentExists {
		utils.JSONError(w, http.StatusNotFound, "student_not_found", "Student not found")
		return
	}

	now := time.Now()
	_, err = tx.Exec(`UPDATE recharge_cards SET is_used = 1, used_by = ?, used_at = ? WHERE id = ?`,
		req.StudentID, now.Format(time.RFC3339), card.ID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	_, err = tx.Exec(`UPDATE students SET balance = balance + ? WHERE id = ?`, card.Amount, req.StudentID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	tx.Commit()

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Recharge successful",
		"amount":  card.Amount,
	})
}

func EnrollCourse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StudentID int64 `json:"student_id"`
		CourseID  int64 `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRow("SELECT id FROM enrollments WHERE student_id = ? AND course_id = ?", req.StudentID, req.CourseID).Scan(&existingID)
	if err == nil {
		utils.JSONError(w, http.StatusConflict, "already_enrolled", "Already enrolled in this course")
		return
	} else if err != sql.ErrNoRows {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	var course models.Course
	var price int64
	err = tx.QueryRow("SELECT id, price FROM courses WHERE id = ?", req.CourseID).Scan(&course.ID, &price)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "course_not_found", "Course not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	var balance int64
	err = tx.QueryRow("SELECT balance FROM students WHERE id = ?", req.StudentID).Scan(&balance)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "student_not_found", "Student not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if price > balance {
		utils.JSONError(w, http.StatusBadRequest, "insufficient_balance", "Insufficient balance")
		return
	}

	_, err = tx.Exec(`INSERT INTO enrollments (student_id, course_id, price_paid) VALUES (?, ?, ?)`,
		req.StudentID, req.CourseID, price)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	_, err = tx.Exec(`UPDATE students SET balance = balance - ? WHERE id = ?`, price, req.StudentID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	tx.Commit()

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Enrolled successfully",
	})
}

func GetStudentEnrollments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid student ID")
		return
	}

	rows, err := database.DB.Query(`SELECT e.id, e.course_id, e.price_paid, e.enrolled_at, 
		c.name, c.instructor, c.course_type, c.subject
		FROM enrollments e JOIN courses c ON e.course_id = c.id 
		WHERE e.student_id = ?`, studentID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	type EnrollmentDetail struct {
		ID         int64     `json:"id"`
		CourseID   int64     `json:"course_id"`
		CourseName string    `json:"course_name"`
		Instructor string    `json:"instructor"`
		CourseType string    `json:"course_type"`
		Subject    string    `json:"subject"`
		PricePaid  int64     `json:"price_paid"`
		EnrolledAt time.Time `json:"enrolled_at"`
	}

	enrollments := []EnrollmentDetail{}
	for rows.Next() {
		var e EnrollmentDetail
		err := rows.Scan(&e.ID, &e.CourseID, &e.PricePaid, &e.EnrolledAt,
			&e.CourseName, &e.Instructor, &e.CourseType, &e.Subject)
		if err != nil {
			continue
		}
		enrollments = append(enrollments, e)
	}

	utils.JSONResponse(w, http.StatusOK, enrollments)
}
