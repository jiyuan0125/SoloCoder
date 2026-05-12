package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"education-platform/database"
	"education-platform/models"
	"education-platform/utils"

	"github.com/gorilla/mux"
)

func UpdateWatchProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID, err := strconv.ParseInt(vars["student_id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid student ID")
		return
	}

	courseID, err := strconv.ParseInt(vars["course_id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid course ID")
		return
	}

	var req struct {
		WatchedSeconds int64 `json:"watched_seconds"`
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

	var videoDuration int64
	err = tx.QueryRow("SELECT video_duration FROM courses WHERE id = ? AND course_type = ?", courseID, models.CourseTypeRecord).Scan(&videoDuration)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "not_found", "Recorded course not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	var existingID int64
	var currentWatched int64
	var isCompleted bool
	err = tx.QueryRow("SELECT id, watched_seconds, is_completed FROM watch_progress WHERE student_id = ? AND course_id = ?", studentID, courseID).Scan(&existingID, &currentWatched, &isCompleted)

	var newWatched int64
	if err == nil {
		newWatched = currentWatched
		if req.WatchedSeconds > currentWatched {
			newWatched = req.WatchedSeconds
		}
	} else if err == sql.ErrNoRows {
		newWatched = req.WatchedSeconds
	} else {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	progress := float64(newWatched) / float64(videoDuration) * 100
	shouldMarkComplete := progress >= 90 && !isCompleted

	now := time.Now().Format(time.RFC3339)
	if err == nil {
		if shouldMarkComplete {
			_, err = tx.Exec(`UPDATE watch_progress SET watched_seconds = ?, is_completed = 1, completed_at = ?, updated_at = ? WHERE id = ?`,
				newWatched, now, now, existingID)
		} else {
			_, err = tx.Exec(`UPDATE watch_progress SET watched_seconds = ?, updated_at = ? WHERE id = ?`,
				newWatched, now, existingID)
		}
	} else {
		if shouldMarkComplete {
			_, err = tx.Exec(`INSERT INTO watch_progress (student_id, course_id, watched_seconds, is_completed, completed_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`,
				studentID, courseID, newWatched, now, now)
		} else {
			_, err = tx.Exec(`INSERT INTO watch_progress (student_id, course_id, watched_seconds, is_completed, updated_at) VALUES (?, ?, ?, 0, ?)`,
				studentID, courseID, newWatched, now)
		}
	}

	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	tx.Commit()

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"watched_seconds": newWatched,
		"progress":        fmt.Sprintf("%.2f%%", progress),
		"is_completed":    shouldMarkComplete || isCompleted,
	})
}

func GetWatchProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID, err := strconv.ParseInt(vars["student_id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid student ID")
		return
	}

	rows, err := database.DB.Query(`SELECT wp.id, wp.student_id, wp.course_id, wp.watched_seconds, wp.is_completed, wp.completed_at, wp.updated_at,
		c.name, c.video_duration, c.course_type, c.instructor
		FROM watch_progress wp JOIN courses c ON wp.course_id = c.id WHERE wp.student_id = ?`, studentID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	type ProgressDetail struct {
		models.WatchProgress
		CourseName     string  `json:"course_name"`
		Instructor     string  `json:"instructor"`
		CourseType     string  `json:"course_type"`
		VideoDuration  *int64  `json:"video_duration,omitempty"`
		ProgressPercent float64 `json:"progress_percent"`
	}

	progresses := []ProgressDetail{}
	for rows.Next() {
		var p ProgressDetail
		var completedAt sql.NullString
		var videoDuration sql.NullInt64
		err := rows.Scan(&p.ID, &p.StudentID, &p.CourseID, &p.WatchedSeconds, &p.IsCompleted,
			&completedAt, &p.UpdatedAt, &p.CourseName, &videoDuration, &p.CourseType, &p.Instructor)
		if err != nil {
			continue
		}
		if completedAt.Valid {
			if t, err := time.Parse(time.RFC3339, completedAt.String); err == nil {
				p.CompletedAt = &t
			}
		}
		if videoDuration.Valid && videoDuration.Int64 > 0 {
			p.VideoDuration = &videoDuration.Int64
			p.ProgressPercent = float64(p.WatchedSeconds) / float64(videoDuration.Int64) * 100
		}
		progresses = append(progresses, p)
	}

	utils.JSONResponse(w, http.StatusOK, progresses)
}

func CreateReview(w http.ResponseWriter, r *http.Request) {
	var review models.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if review.Rating < 1 || review.Rating > 5 {
		utils.JSONError(w, http.StatusBadRequest, "invalid_rating", "Rating must be between 1 and 5")
		return
	}

	if len(review.Comment) > 200 {
		utils.JSONError(w, http.StatusBadRequest, "comment_too_long", "Comment must be 200 characters or less")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer tx.Rollback()

	var isCompleted bool
	err = tx.QueryRow("SELECT is_completed FROM watch_progress WHERE student_id = ? AND course_id = ?", review.StudentID, review.CourseID).Scan(&isCompleted)
	if err != nil || !isCompleted {
		utils.JSONError(w, http.StatusBadRequest, "not_completed", "Must complete course before reviewing")
		return
	}

	var existingID int64
	err = tx.QueryRow("SELECT id FROM reviews WHERE student_id = ? AND course_id = ?", review.StudentID, review.CourseID).Scan(&existingID)
	if err == nil {
		utils.JSONError(w, http.StatusConflict, "already_reviewed", "Already reviewed this course")
		return
	}

	result, err := tx.Exec(`INSERT INTO reviews (student_id, course_id, rating, comment) VALUES (?, ?, ?, ?)`,
		review.StudentID, review.CourseID, review.Rating, review.Comment)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	id, _ := result.LastInsertId()
	review.ID = id
	tx.Commit()

	utils.JSONResponse(w, http.StatusCreated, review)
}

func GetCourseReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid course ID")
		return
	}

	rows, err := database.DB.Query(`SELECT r.id, r.student_id, r.course_id, r.rating, r.comment, r.created_at, s.nickname
		FROM reviews r JOIN students s ON r.student_id = s.id WHERE r.course_id = ?`, courseID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	type ReviewDetail struct {
		models.Review
		StudentNickname string `json:"student_nickname"`
	}

	reviews := []ReviewDetail{}
	for rows.Next() {
		var r ReviewDetail
		err := rows.Scan(&r.ID, &r.StudentID, &r.CourseID, &r.Rating, &r.Comment, &r.CreatedAt, &r.StudentNickname)
		if err != nil {
			continue
		}
		reviews = append(reviews, r)
	}

	utils.JSONResponse(w, http.StatusOK, reviews)
}

func GetSubjectStats(w http.ResponseWriter, r *http.Request) {
	subjectNames := map[string]string{
		"chinese":   "语文",
		"math":      "数学",
		"english":   "英语",
		"physics":   "物理",
		"chemistry": "化学",
		"biology":   "生物",
		"history":   "历史",
		"geography": "地理",
		"politics":  "政治",
	}

	rows, err := database.DB.Query(`
		SELECT 
			c.subject,
			COUNT(DISTINCT c.id) as course_count,
			COUNT(DISTINCT e.id) as enrollment_count
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		WHERE c.subject IS NOT NULL
		GROUP BY c.subject
		ORDER BY enrollment_count DESC
	`)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	stats := []models.CourseStats{}
	for rows.Next() {
		var s models.CourseStats
		var subject string
		err := rows.Scan(&subject, &s.CourseCount, &s.EnrollmentCount)
		if err != nil {
			continue
		}
		s.Subject = subject
		s.SubjectName = subjectNames[subject]
		stats = append(stats, s)
	}

	utils.JSONResponse(w, http.StatusOK, stats)
}

func GetInstructorStats(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT 
			c.instructor,
			COUNT(DISTINCT c.id) as course_count,
			AVG(r.rating) as avg_rating
		FROM courses c
		LEFT JOIN reviews r ON c.id = r.course_id
		GROUP BY c.instructor
		ORDER BY course_count DESC
	`)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	stats := []models.InstructorStats{}
	for rows.Next() {
		var s models.InstructorStats
		err := rows.Scan(&s.InstructorName, &s.CourseCount, &s.AvgRating)
		if err != nil {
			continue
		}
		stats = append(stats, s)
	}

	utils.JSONResponse(w, http.StatusOK, stats)
}

func GetAllLiveStats(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT ls.id, ls.course_id, ls.peak_online, ls.average_online_time, ls.danmaku_count, ls.vote_participation, ls.created_at, c.name
		FROM live_stats ls JOIN courses c ON ls.course_id = c.id
	`)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	type LiveStatDetail struct {
		models.LiveStats
		CourseName string `json:"course_name"`
	}

	stats := []LiveStatDetail{}
	for rows.Next() {
		var s LiveStatDetail
		err := rows.Scan(&s.ID, &s.CourseID, &s.PeakOnline, &s.AverageOnlineTime, &s.DanmakuCount, &s.VoteParticipation, &s.CreatedAt, &s.CourseName)
		if err != nil {
			continue
		}
		stats = append(stats, s)
	}

	utils.JSONResponse(w, http.StatusOK, stats)
}
