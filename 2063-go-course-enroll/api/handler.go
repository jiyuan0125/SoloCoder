package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"course-enroll/modules/course"
	"course-enroll/modules/enrollment"
	"course-enroll/modules/queue"
	"course-enroll/modules/resource"
	"course-enroll/storage"
)

type Handler struct {
	courseMgr *course.Manager
	engine    *enrollment.Engine
	queueSys  *queue.System
	resourceMgr *resource.Manager
	store     storage.Store
}

func NewHandler(courseMgr *course.Manager, engine *enrollment.Engine, queueSys *queue.System, resourceMgr *resource.Manager, store storage.Store) *Handler {
	return &Handler{
		courseMgr:   courseMgr,
		engine:      engine,
		queueSys:    queueSys,
		resourceMgr: resourceMgr,
		store:       store,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/courses", h.handleCourses)
	mux.HandleFunc("/api/courses/", h.handleCourseByID)
	
	mux.HandleFunc("/api/students", h.handleStudents)
	mux.HandleFunc("/api/students/", h.handleStudentByID)
	
	mux.HandleFunc("/api/enroll", h.handleEnroll)
	mux.HandleFunc("/api/drop", h.handleDrop)
	
	mux.HandleFunc("/api/queue/", h.handleQueue)
	
	mux.HandleFunc("/api/resources/", h.handleResource)
	
	mux.HandleFunc("/health", h.handleHealth)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleCourses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		courses, err := h.courseMgr.ListCourses(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, courses)
	case http.MethodPost:
		h.createCourse(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createCourse(w http.ResponseWriter, r *http.Request) {
	var req course.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "empty request body")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	courseInfo, err := h.courseMgr.CreateCourse(r.Context(), &req)
	if err != nil {
		if errors.Is(err, course.ErrInvalidCapacity) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, course.ErrInvalidTime) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, courseInfo)
}

func (h *Handler) handleCourseByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/api/courses/")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseInfo, err := h.courseMgr.GetCourse(r.Context(), id)
	if err != nil {
		if errors.Is(err, course.ErrCourseNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, courseInfo)
}

func (h *Handler) handleStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := h.engine.CreateStudent(r.Context(), req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "name": req.Name})
}

func (h *Handler) handleStudentByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/students/"), "/")
	if len(parts) < 1 {
		http.NotFound(w, r)
		return
	}

	studentID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 2 && parts[1] == "complete" && r.Method == http.MethodPost {
		h.markCourseCompleted(w, r, studentID)
		return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		h.getStudentSchedule(w, r, studentID)
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) markCourseCompleted(w http.ResponseWriter, r *http.Request, studentID int64) {
	var req struct {
		CourseID int64 `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.engine.MarkCourseCompleted(r.Context(), studentID, req.CourseID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "marked as completed"})
}

func (h *Handler) getStudentSchedule(w http.ResponseWriter, r *http.Request, studentID int64) {
	enrollments, err := h.store.ListStudentEnrollments(r.Context(), studentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type courseSchedule struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}

	var schedule []courseSchedule
	for _, e := range enrollments {
		c, err := h.store.GetCourse(r.Context(), e.CourseID)
		if err != nil || c == nil {
			continue
		}
		schedule = append(schedule, courseSchedule{
			ID:        c.ID,
			Name:      c.Name,
			StartTime: c.StartTime,
			EndTime:   c.EndTime,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"student_id": studentID,
		"schedule":   schedule,
	})
}

func (h *Handler) handleEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StudentID int64 `json:"student_id"`
		CourseID  int64 `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, enrollErr := h.engine.Enroll(r.Context(), req.StudentID, req.CourseID)
	if enrollErr != nil {
		if enrollErr.MissingPrereqs != nil {
			writeJSON(w, enrollErr.Status, map[string]interface{}{
				"error":           enrollErr.Message,
				"missing_prereqs": enrollErr.MissingPrereqs,
			})
			return
		}
		if enrollErr.Conflicts != nil {
			writeJSON(w, enrollErr.Status, map[string]interface{}{
				"error":     enrollErr.Message,
				"conflicts": enrollErr.Conflicts,
			})
			return
		}
		writeError(w, enrollErr.Status, enrollErr.Message)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleDrop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StudentID int64 `json:"student_id"`
		CourseID  int64 `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, dropErr := h.engine.Drop(r.Context(), req.StudentID, req.CourseID)
	if dropErr != nil {
		writeError(w, dropErr.Status, dropErr.Message)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/queue/")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	items, err := h.queueSys.ListQueue(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"course_id": id,
		"queue":     items,
	})
}

func (h *Handler) handleResource(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/resources/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	resourceType := parts[0]
	resourceID := parts[1]

	if r.Method == http.MethodPost && len(parts) == 2 {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := h.resourceMgr.CreateResource(r.Context(), &resource.ResourceInfo{
			ID:   resourceID,
			Type: resourceType,
			Name: req.Name,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{"id": resourceID, "type": resourceType})
		return
	}

	if r.Method == http.MethodGet && len(parts) == 2 {
		summary, err := h.resourceMgr.GetResourceSummary(r.Context(), resourceID, resourceType)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, summary)
		return
	}

	http.NotFound(w, r)
}

func parseIDFromPath(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, "/")
	return strconv.ParseInt(idStr, 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
