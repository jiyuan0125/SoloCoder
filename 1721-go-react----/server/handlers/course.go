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

func CreateCourse(w http.ResponseWriter, r *http.Request) {
	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if course.Price < 0 {
		utils.JSONError(w, http.StatusBadRequest, "negative_price", "Price cannot be negative")
		return
	}

	if course.CourseType == models.CourseTypeLive && len(course.Name) == 0 {
		utils.JSONError(w, http.StatusBadRequest, "invalid_course", "Course name is required")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRow("SELECT id FROM courses WHERE name = ?", course.Name).Scan(&existingID)
	if err == nil {
		utils.JSONError(w, http.StatusConflict, "duplicate_name", "Course name already exists")
		return
	} else if err != sql.ErrNoRows {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	var result sql.Result
	if course.CourseType == models.CourseTypeLive {
		var startTimeStr *string
		if course.StartTime != nil {
			s := course.StartTime.Format(time.RFC3339)
			startTimeStr = &s
		}
		result, err = tx.Exec(`INSERT INTO courses (name, instructor, course_type, price, subject, start_time, duration, max_online, live_status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			course.Name, course.Instructor, course.CourseType, course.Price,
			course.Subject, startTimeStr, course.Duration, course.MaxOnline, models.LiveStatusNotStarted)
	} else {
		result, err = tx.Exec(`INSERT INTO courses (name, instructor, course_type, price, subject, video_duration, description) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			course.Name, course.Instructor, course.CourseType, course.Price,
			course.Subject, course.VideoDuration, course.Description)
	}

	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	courseID, _ := result.LastInsertId()
	tx.Commit()

	course.ID = courseID
	utils.JSONResponse(w, http.StatusCreated, course)
}

func ListCourses(w http.ResponseWriter, r *http.Request) {
	subject := r.URL.Query().Get("subject")
	courseType := r.URL.Query().Get("type")

	query := `SELECT id, name, instructor, course_type, price, subject, 
		start_time, duration, max_online, live_status, video_duration, description, 
		created_at, updated_at FROM courses WHERE 1=1`
	args := []interface{}{}

	if subject != "" {
		query += " AND subject = ?"
		args = append(args, subject)
	}
	if courseType != "" {
		query += " AND course_type = ?"
		args = append(args, courseType)
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}
	defer rows.Close()

	courses := []models.Course{}
	for rows.Next() {
		var c models.Course
		var startTime, createdAt, updatedAt sql.NullString
		var duration, maxOnline, videoDuration sql.NullInt64
		var liveStatus, subject, description sql.NullString

		err := rows.Scan(&c.ID, &c.Name, &c.Instructor, &c.CourseType, &c.Price, &subject,
			&startTime, &duration, &maxOnline, &liveStatus, &videoDuration, &description,
			&createdAt, &updatedAt)
		if err != nil {
			continue
		}

		if subject.Valid {
			c.Subject = models.Subject(subject.String)
		}
		if startTime.Valid {
			if t, err := time.Parse(time.RFC3339, startTime.String); err == nil {
				c.StartTime = &t
			}
		}
		if duration.Valid {
			d := int(duration.Int64)
			c.Duration = &d
		}
		if maxOnline.Valid {
			m := int(maxOnline.Int64)
			c.MaxOnline = &m
		}
		if liveStatus.Valid {
			s := models.LiveCourseStatus(liveStatus.String)
			c.LiveStatus = &s
		}
		if videoDuration.Valid {
			c.VideoDuration = &videoDuration.Int64
		}
		if description.Valid {
			c.Description = &description.String
		}

		courses = append(courses, c)
	}

	utils.JSONResponse(w, http.StatusOK, courses)
}

func GetCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid course ID")
		return
	}

	var c models.Course
	var startTime, createdAt, updatedAt sql.NullString
	var duration, maxOnline, videoDuration sql.NullInt64
	var liveStatus, subject, description sql.NullString

	err = database.DB.QueryRow(`SELECT id, name, instructor, course_type, price, subject, 
		start_time, duration, max_online, live_status, video_duration, description, 
		created_at, updated_at FROM courses WHERE id = ?`, id).Scan(
		&c.ID, &c.Name, &c.Instructor, &c.CourseType, &c.Price, &subject,
		&startTime, &duration, &maxOnline, &liveStatus, &videoDuration, &description,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "not_found", "Course not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if subject.Valid {
		c.Subject = models.Subject(subject.String)
	}
	if startTime.Valid {
		if t, err := time.Parse(time.RFC3339, startTime.String); err == nil {
			c.StartTime = &t
		}
	}
	if duration.Valid {
		d := int(duration.Int64)
		c.Duration = &d
	}
	if maxOnline.Valid {
		m := int(maxOnline.Int64)
		c.MaxOnline = &m
	}
	if liveStatus.Valid {
		s := models.LiveCourseStatus(liveStatus.String)
		c.LiveStatus = &s
	}
	if videoDuration.Valid {
		c.VideoDuration = &videoDuration.Int64
	}
	if description.Valid {
		c.Description = &description.String
	}

	utils.JSONResponse(w, http.StatusOK, c)
}

func UpdateLiveCourseStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid course ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	validStatuses := map[string]bool{
		"not_started": true,
		"in_progress": true,
		"ended":       true,
		"has_replay":  true,
	}

	if !validStatuses[req.Status] {
		utils.JSONError(w, http.StatusBadRequest, "invalid_status", "Invalid status")
		return
	}

	result, err := database.DB.Exec("UPDATE courses SET live_status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", req.Status, courseID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		utils.JSONError(w, http.StatusNotFound, "not_found", "Course not found")
		return
	}

	if req.Status == "ended" {
		generateLiveStats(courseID)
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func generateLiveStats(courseID int64) {
	var danmakuCount int64
	database.DB.QueryRow("SELECT COUNT(*) FROM danmakus WHERE course_id = ?", courseID).Scan(&danmakuCount)

	var peakOnline int
	database.DB.QueryRow(`SELECT COUNT(DISTINCT student_id) FROM live_sessions WHERE course_id = ? GROUP BY DATE(join_time) ORDER BY COUNT(*) DESC LIMIT 1`, courseID).Scan(&peakOnline)

	var avgDuration float64
	database.DB.QueryRow(`SELECT AVG(duration_seconds) FROM live_sessions WHERE course_id = ? AND duration_seconds > 0`, courseID).Scan(&avgDuration)

	var totalParticipants, totalVoters int
	database.DB.QueryRow(`SELECT COUNT(DISTINCT student_id) FROM live_sessions WHERE course_id = ?`, courseID).Scan(&totalParticipants)
	database.DB.QueryRow(`SELECT COUNT(DISTINCT vp.student_id) FROM vote_participants vp JOIN votes v ON vp.vote_id = v.id WHERE v.course_id = ?`, courseID).Scan(&totalVoters)

	var participationRate float64
	if totalParticipants > 0 {
		participationRate = float64(totalVoters) / float64(totalParticipants) * 100
	}

	database.DB.Exec(`INSERT INTO live_stats (course_id, peak_online, average_online_time, danmaku_count, vote_participation) 
		VALUES (?, ?, ?, ?, ?) 
		ON CONFLICT(course_id) DO UPDATE SET 
		peak_online = excluded.peak_online,
		average_online_time = excluded.average_online_time,
		danmaku_count = excluded.danmaku_count,
		vote_participation = excluded.vote_participation`,
		courseID, peakOnline, avgDuration, danmakuCount, participationRate)

	fmt.Printf("Generated stats for course %d: peak=%d, avg=%.2f, danmaku=%d, participation=%.2f%%\n",
		courseID, peakOnline, avgDuration, danmakuCount, participationRate)
}

func GetLiveStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid_id", "Invalid course ID")
		return
	}

	var stats models.LiveStats
	err = database.DB.QueryRow(`SELECT id, course_id, peak_online, average_online_time, danmaku_count, vote_participation, created_at 
		FROM live_stats WHERE course_id = ?`, courseID).Scan(
		&stats.ID, &stats.CourseID, &stats.PeakOnline, &stats.AverageOnlineTime,
		&stats.DanmakuCount, &stats.VoteParticipation, &stats.CreatedAt)

	if err == sql.ErrNoRows {
		utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "No stats available yet"})
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, stats)
}
